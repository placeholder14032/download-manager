package download

import(
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

    // Create new pause channel
    h.PauseChan = make(chan struct{})

    // Get incomplete parts and clear them from state
    incompleteParts := make([]chunk, len(h.State.IncompleteParts))
    copy(incompleteParts, h.State.IncompleteParts)
    h.State.IncompleteParts = make([]chunk, 0)
    h.State.Mutex.Unlock()

    return h.resumingDownload(d, incompleteParts)
}

func (h *DownloadHandler) resumingDownload(d Download, incompleteParts []chunk) error {
    //  check if it's already completed
    completedCount := 0
    h.State.Mutex.Lock()
    for _, completed := range h.State.Completed {
        if completed {
            completedCount++
        }
    }
    h.State.Mutex.Unlock()

    // if already completed -> so we should just combine parts
    if completedCount == h.PartsCount {
        return h.combineParts(&d, int(h.State.TotalBytes))
    }

    // else we should resume now we will do the same as in StartDownloading we are defining and alligning jobs
    // then we will start workers and distribute jobs
    jobs := make(chan chunk, h.WORKERS_COUNT)
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool)
    pauseAck := make(chan bool, h.WORKERS_COUNT)
    var wg sync.WaitGroup

    // start workers
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, &d, jobs, errChan, pauseAck, &wg)
    }

    // job distribution
    go func() {
        defer close(jobs)
        
        //  handle incomplete parts first
        for _, part := range incompleteParts {
            
            h.State.Mutex.Lock()
            isCompleted := h.State.Completed[part.Start/h.CHUNK_SIZE]
            h.State.Mutex.Unlock()
            
            if !isCompleted {
                select {
                case <-h.PauseChan:
                    h.State.Mutex.Lock()
                    h.State.IncompleteParts = append(h.State.IncompleteParts, part)
                    h.State.Mutex.Unlock()
                    return
                case jobs <- part:
                    fmt.Printf("Resuming part %d (bytes %d-%d)\n", 
                        part.Start/h.CHUNK_SIZE, part.Start, part.End)
                }
            }
        }

        // then we will handle the remaining parts

        h.State.Mutex.Lock()
        currentByte := h.State.CurrentByte
        h.State.Mutex.Unlock()

        for currentByte < h.State.TotalBytes {
            partIndex := int(currentByte) / h.CHUNK_SIZE
            
            h.State.Mutex.Lock()
            isCompleted := h.State.Completed[partIndex]
            h.State.Mutex.Unlock()

            if !isCompleted {
                end := currentByte + int64(h.CHUNK_SIZE)
                if end > h.State.TotalBytes {
                    end = h.State.TotalBytes
                }

                chunk := chunk{Start: int(currentByte), End: int(end - 1)}
                select {
                case <-h.PauseChan:
                    h.State.Mutex.Lock()
                    h.State.CurrentByte = currentByte
                    h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
                    h.State.Mutex.Unlock()
                    return
                case jobs <- chunk:
                    fmt.Printf("Downloading part %d (bytes %d-%d)\n", 
                        partIndex, currentByte, end-1)
                }
            }
            currentByte += int64(h.CHUNK_SIZE)
        }
    }()

    // Wait and handle completion
    go func() {
        wg.Wait()
        if !h.State.IsPaused {
            close(errChan)
            done <- true
        }
    }()

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

