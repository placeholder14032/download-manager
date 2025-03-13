package download

import (
    "fmt"
    "sync"
)

func (h *DownloadHandler) Pause() error {
    if h.Download.State == nil {
        return fmt.Errorf("Download not initialized")
    }

    h.Download.State.Mutex.Lock()
    defer h.Download.State.Mutex.Unlock()

    if h.Download.State.IsPaused {
        return fmt.Errorf("Download is already paused")
    }

    h.Download.State.IsPaused = true
    
    if h.PauseChan != nil {
        close(h.PauseChan)
        fmt.Println("Download pause signal sent")
    }

    return nil
}

func (h *DownloadHandler) Resume(d Download) error {
    if h.Download.State == nil {
        return fmt.Errorf("Download not initialized")
    }

    h.Download.State.Mutex.Lock()
    if !h.Download.State.IsPaused {
        h.Download.State.Mutex.Unlock()
        return fmt.Errorf("Download is not paused")
    }

    h.Download.State.IsPaused = false
    h.PauseChan = make(chan struct{})
    incompleteParts := h.prepareResume()
    h.Download.State.Mutex.Unlock()

    return h.resumeDownload(d, incompleteParts)
}

func (h *DownloadHandler) prepareResume() []chunk {
    incompleteParts := make([]chunk, len(h.Download.State.IncompleteParts))
    copy(incompleteParts, h.Download.State.IncompleteParts)
    h.Download.State.IncompleteParts = make([]chunk, 0)
    return incompleteParts
}

func (h *DownloadHandler) resumeDownload(d Download, incompleteParts []chunk) error {
    if h.isDownloadComplete() {
        return h.combineParts(int(h.Download.State.TotalBytes))
    }

    return h.resumeDownloadWorkers(d, incompleteParts) 
}

func (h *DownloadHandler) isDownloadComplete() bool {
    h.Download.State.Mutex.Lock()
    defer h.Download.State.Mutex.Unlock()
    
    completedCount := 0
    for _, completed := range h.Download.State.Completed {
        if completed {
            completedCount++
        }
    }
    return completedCount == h.PartsCount
}

func (h *DownloadHandler) resumeDownloadWorkers(d Download, incompleteParts []chunk) error {
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
    h.Download.State.Mutex.Lock()
    currentByte := h.Download.State.CurrentByte
    h.Download.State.Mutex.Unlock()

    for currentByte < h.Download.State.TotalBytes {
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
    h.Download.State.Mutex.Lock()
    defer h.Download.State.Mutex.Unlock()
    return h.Download.State.Completed[partIndex]
}

func (h *DownloadHandler) createChunk(currentByte int64) chunk {
    end := currentByte + int64(h.CHUNK_SIZE)
    if end > h.Download.State.TotalBytes {
        end = h.Download.State.TotalBytes
    }
    return chunk{Start: int(currentByte), End: int(end - 1)}
}

func (h *DownloadHandler) sendJob(jobs chan chunk, part chunk) bool {
    partIndex := part.Start / h.CHUNK_SIZE
    select {
    case <-h.PauseChan:
        h.Download.State.Mutex.Lock()
        h.Download.State.IncompleteParts = append(h.Download.State.IncompleteParts, part)
        h.Download.State.Mutex.Unlock()
        return false
    case jobs <- part:
        fmt.Printf("Downloading part %d (bytes %d-%d)\n", 
            partIndex, part.Start, part.End)
        return true
    }
}

func (h *DownloadHandler) handleCompletion(errChan chan error, done chan bool, wg *sync.WaitGroup) {
    wg.Wait()
    if !h.Download.State.IsPaused {
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
        if !h.Download.State.IsPaused {
            return h.combineParts(int(h.Download.State.TotalBytes))
        }
    case <-done:
        return h.combineParts(int(h.Download.State.TotalBytes))
    }
    return nil
}