package download

import (
    "fmt"
    "sync"
)

func (h *DownloadHandler) Pause() error {
    if h.State == nil {
        return fmt.Errorf("download not initialized")
    }

    h.State.Mutex.Lock()
    defer h.State.Mutex.Unlock()

    if h.State.IsPaused {
        return fmt.Errorf("download is already paused")
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
        return fmt.Errorf("download not initialized")
    }

    h.State.Mutex.Lock()
    if !h.State.IsPaused {
        h.State.Mutex.Unlock()
        return fmt.Errorf("download is not paused")
    }

    h.State.IsPaused = false
    h.PauseChan = make(chan struct{})
    incompleteParts := h.prepareResume()
    h.State.Mutex.Unlock()

    return h.resumeDownload(d, incompleteParts)
}

func (h *DownloadHandler) prepareResume() []chunk {
    incompleteParts := make([]chunk, len(h.State.IncompleteParts))
    copy(incompleteParts, h.State.IncompleteParts)
    h.State.IncompleteParts = make([]chunk, 0)
    return incompleteParts
}

func (h *DownloadHandler) resumeDownload(d Download, incompleteParts []chunk) error {
    if h.isDownloadComplete() {
        return h.combineParts(&d, int(h.State.TotalBytes))
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
    // Kindda the same logic as downloadHandler startinfDownload
    //  3 channels like downloadHandler
    jobs := make(chan chunk, h.WORKERS_COUNT)
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool)
    pauseAck := make(chan bool, h.WORKERS_COUNT)
    var wg sync.WaitGroup

    // starting workers
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, &d, jobs, errChan, pauseAck, &wg)
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
            return h.combineParts(&d, int(h.State.TotalBytes))
        }
    case <-done:
        return h.combineParts(&d, int(h.State.TotalBytes))
    }
    return nil
}