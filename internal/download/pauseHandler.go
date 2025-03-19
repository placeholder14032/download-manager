package download

import (
	"fmt"
	"os"
	"sync"
)

func (h *DownloadHandler) Pause() error {
    if h.State == nil {
        return fmt.Errorf("Download not initialized")
    }

    h.State.Mutex.Lock()
    defer h.State.Mutex.Unlock()

    if h.State.IsPaused {
        return fmt.Errorf("Download is already paused")
    }

    h.State.IsPaused = true
    
    if h.PauseChan != nil {
        close(h.PauseChan)
        fmt.Println("Download pause signal sent")
    }

    return nil
}

func (h *DownloadHandler) Resume(d Download) error {
    if h.State == nil {
        return fmt.Errorf("Download not initialized")
    }

    h.State.Mutex.Lock()
    if !h.State.IsPaused {
        h.State.Mutex.Unlock()
        return fmt.Errorf("Download is not paused")
    }

    // Re-initialize state if needed
    if h.PartsCount == 0 {
        // Get content length again
        _, contentLength, err := h.IsAcceptRangeSupported()
        if err != nil {
            h.State.Mutex.Unlock()
            return fmt.Errorf("failed to get content length: %v", err)
        }

        h.PartsCount = (contentLength + h.CHUNK_SIZE - 1) / h.CHUNK_SIZE
        if h.State.Completed == nil {
            h.State.Completed = make([]bool, h.PartsCount)
        }
        h.State.TotalBytes = int64(contentLength)
    }

    // Validate existing part files
    for i := 0; i < h.PartsCount; i++ {
        partFileName := fmt.Sprintf("%s.part%d", h.FilePath, i)
        if _, err := os.Stat(partFileName); os.IsNotExist(err) {
            // Create empty file if missing
            file, err := os.OpenFile(partFileName, os.O_WRONLY|os.O_CREATE, 0644)
            if err != nil {
                h.State.Mutex.Unlock()
                return fmt.Errorf("failed to create missing part file %s: %v", partFileName, err)
            }
            file.Close()
            h.State.Completed[i] = false
        }
    }

    h.State.IsPaused = false
    h.PauseChan = make(chan struct{})
    incompleteParts := h.prepareResume()
    h.State.Mutex.Unlock()

    return h.resumeDownload(d, incompleteParts)
}

func (h *DownloadHandler) prepareResume() []chunk {
    // Verify all "completed" parts actually exist with correct size
    for i := 0; i < h.PartsCount; i++ {
        if h.State.Completed[i] {
            partFileName := fmt.Sprintf("%s.part%d", h.FilePath, i)
            info, err := os.Stat(partFileName)
            
            start := i * h.CHUNK_SIZE
            end := start + h.CHUNK_SIZE
            if end > int(h.State.TotalBytes) {
                end = int(h.State.TotalBytes)
            }
            expectedSize := end - start

            if err != nil || info.Size() != int64(expectedSize) {
                fmt.Printf("Part %d marked incomplete: size %d, expected %d\n", i, info.Size(), expectedSize)
                h.State.Completed[i] = false
            }
        }
    }

    // If no incomplete parts are tracked, reconstruct them from Completed array
    if len(h.State.IncompleteParts) == 0 {
        for i := 0; i < h.PartsCount; i++ {
            if !h.State.Completed[i] {
                start := i * h.CHUNK_SIZE
                end := start + h.CHUNK_SIZE
                if end > int(h.State.TotalBytes) {
                    end = int(h.State.TotalBytes)
                }
                h.State.IncompleteParts = append(h.State.IncompleteParts, chunk{
                    Start: start,
                    End:   end - 1,
                })
            }
        }
    }

    incompleteParts := make([]chunk, len(h.State.IncompleteParts))
    copy(incompleteParts, h.State.IncompleteParts)
    h.State.IncompleteParts = make([]chunk, 0)
    return incompleteParts
}

func (h *DownloadHandler) resumeDownload(d Download, incompleteParts []chunk) error {
    if h.isDownloadComplete() {
        return h.combineParts(int(h.State.TotalBytes))
    }

    return h.resumeDownloadWorkers(d, incompleteParts) 
}

func (h *DownloadHandler) isDownloadComplete() bool {
    h.State.Mutex.Lock()
    defer h.State.Mutex.Unlock()
    
    completedCount := 0
    for _, completed := range h.State.Completed {
        if completed {
            completedCount++
        }
    }
    return completedCount == h.PartsCount
}

func (h *DownloadHandler) resumeDownloadWorkers(d Download, incompleteParts []chunk) error {
    // Validate completed parts before resuming
    if err := h.validateCompletedParts(); err != nil {
        return fmt.Errorf("failed to validate completed parts: %v", err)
    }

    // Kindda the same logic as DownloadHandler startinfDownload
    //  3 channels like DownloadHandler
    jobs := make(chan chunk, h.WORKERS_COUNT)
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool)
    pauseAck := make(chan bool, h.WORKERS_COUNT)
    var wg sync.WaitGroup

    // starting workers
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, jobs, errChan, pauseAck, &wg)
    }

    // Distribute jobs like startingDownload
    go h.distributeResumeJobs(jobs, incompleteParts)

    go h.handleCompletion(errChan, done, &wg)

    return h.waitForDownloadResult(d, errChan, done)
}

func (h *DownloadHandler) distributeResumeJobs(jobs chan chunk, incompleteParts []chunk) {
    defer close(jobs)

    // Handle incomplete parts first
    for _, part := range incompleteParts {
        if !h.isPartCompleted(part.Start/h.CHUNK_SIZE) {
            if !h.sendJob(jobs, part) {
                return
            }
        }
    }

    // Handle remaining parts
    h.State.Mutex.Lock()
    currentByte := h.State.CurrentByte
    h.State.Mutex.Unlock()

    for currentByte < h.State.TotalBytes {
        partIndex := int(currentByte) / h.CHUNK_SIZE
        if !h.isPartCompleted(partIndex) {
            chunk := h.createChunk(currentByte)
            if !h.sendJob(jobs, chunk) {
                return
            }
        }
        currentByte += int64(h.CHUNK_SIZE)
    }
}

func (h *DownloadHandler) isPartCompleted(partIndex int) bool {
    h.State.Mutex.Lock()
    defer h.State.Mutex.Unlock()
    return h.State.Completed[partIndex]
}

func (h *DownloadHandler) createChunk(currentByte int64) chunk {
    end := currentByte + int64(h.CHUNK_SIZE)
    if end > h.State.TotalBytes {
        end = h.State.TotalBytes
    }
    return chunk{Start: int(currentByte), End: int(end - 1)}
}

func (h *DownloadHandler) sendJob(jobs chan chunk, part chunk) bool {
    partIndex := part.Start / h.CHUNK_SIZE
    select {
    case <-h.PauseChan:
        h.State.Mutex.Lock()
        h.State.IncompleteParts = append(h.State.IncompleteParts, part)
        h.State.Mutex.Unlock()
        return false
    case jobs <- part:
        fmt.Printf("Downloading part %d (bytes %d-%d)\n", 
            partIndex, part.Start, part.End)
        return true
    }
}

func (h *DownloadHandler) handleCompletion(errChan chan error, done chan bool, wg *sync.WaitGroup) {
    wg.Wait()
    if !h.State.IsPaused {
        close(errChan)
        done <- true
    }
}

func (h *DownloadHandler) waitForDownloadResult(d Download, errChan chan error, done chan bool) error {
    select {
    case err := <-errChan:
        if err != nil {
            return err
        }
        if !h.State.IsPaused {
            return h.combineParts(int(h.State.TotalBytes))
        }
    case <-done:
        return h.combineParts(int(h.State.TotalBytes))
    }
    return nil
}


func (h *DownloadHandler) validateCompletedParts() error {
    h.State.Mutex.Lock()
    defer h.State.Mutex.Unlock()

    for i := 0; i < h.PartsCount; i++ {
        if !h.State.Completed[i] {
            continue // Skip parts that aren't marked as completed
        }

        partFileName := fmt.Sprintf("%s.part%d", h.FilePath, i)
        info, err := os.Stat(partFileName)
        if err != nil {
            if os.IsNotExist(err) {
                // Part file doesn't exist, mark as not completed
                fmt.Printf("Part %d does not exist, marking for re-download\n", i)
                h.State.Completed[i] = false
                continue
            }
            return fmt.Errorf("failed to stat part file %s: %v", partFileName, err)
        }

        expectedSize := h.CHUNK_SIZE
        if i == h.PartsCount-1 {
            expectedSize = int(h.State.TotalBytes % int64(h.CHUNK_SIZE))
            if expectedSize == 0 {
                expectedSize = h.CHUNK_SIZE
            }
        }

        if info.Size() != int64(expectedSize) {
            fmt.Printf("Part %d size mismatch: got %d, want %d, marking for re-download\n", i, info.Size(), expectedSize)
            h.State.Completed[i] = false
        }
    }
    return nil
}