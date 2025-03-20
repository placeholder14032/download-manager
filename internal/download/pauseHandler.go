package download

import(
	"context"
	"sync"
)

func (h *DownloadHandler) Pause() {
    h.State.Mutex.Lock()
    if h.State.CurrentByte >= h.State.TotalBytes {
        h.State.Mutex.Unlock()
        return // Ignore if complete
    }
    h.State.IsPaused = true
    h.State.Mutex.Unlock()

    close(h.PauseChan)
    h.cancel()
}

func (h *DownloadHandler) Resume() error {
    h.State.Mutex.Lock()
    if !h.State.IsPaused {
        h.State.Mutex.Unlock()
        return nil
    }
    h.State.IsPaused = false
    h.State.Mutex.Unlock()

    // Create new context and channels
    h.ctx, h.cancel = context.WithCancel(context.Background())
    h.PauseChan = make(chan struct{})
    close(h.ResumeChan) // Signal all waiting workers
    h.ResumeChan = make(chan struct{}) // Create new channel for future pauses
    
    // Restart work from where we left off
    return h.restartDownload()
}

func (h *DownloadHandler) restartDownload() error {
    // Restart with a fresh worker pool and job distribution
    jobs := make(chan chunk, h.WORKERS_COUNT)
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool, 1)
    
    // Start fresh worker pool
    var wg sync.WaitGroup
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, jobs, errChan,done, &wg)
    }
    
    // First handle any incomplete parts
    go func() {
        h.State.Mutex.Lock()
        incomplete := h.State.IncompleteParts
        h.State.IncompleteParts = nil
        h.State.Mutex.Unlock()
        
        for _, chunk := range incomplete {
            jobs <- chunk
        }
        
        // Then continue with remaining parts
        if h.State.CurrentByte < h.State.TotalBytes {
            h.distributeRemainingJobs(jobs)
        }
        close(jobs)
    }()
    
    // Set up completion monitoring
    go func() {
        wg.Wait()
        h.State.Mutex.Lock()
        if h.State.CurrentByte >= h.State.TotalBytes {
            done <- true
        }
        h.State.Mutex.Unlock()
    }()
    
    // Wait for completion or error
    select {
    case <-done:
        return h.combineParts(h.State.TotalBytes)
    case err := <-errChan:
        return err
    }
}

func (h *DownloadHandler) distributeRemainingJobs(jobs chan<- chunk) {
    currentByte := h.State.CurrentByte
    for currentByte < h.State.TotalBytes {
        end := currentByte + h.CHUNK_SIZE
        if end > h.State.TotalBytes {
            end = h.State.TotalBytes
        }
        jobs <- chunk{Start: currentByte, End: end - 1}
        currentByte = end
    }
}