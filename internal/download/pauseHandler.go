package download

import(
	"context"
	"sync"
	"fmt"
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

    h.ctx, h.cancel = context.WithCancel(context.Background())
    h.PauseChan = make(chan struct{})
    h.ResumeChan <- struct{}{} // Signal resume
    return nil
}

func (h *DownloadHandler) restartIncompleteParts() error {
    fmt.Println("Restarting incomplete parts...")
    h.State.Mutex.Lock()
    if h.State.CurrentByte >= h.State.TotalBytes {
        h.State.Mutex.Unlock()
        return nil // Nothing to restart if complete
    }
    incomplete := h.State.IncompleteParts
    h.State.IncompleteParts = nil
    h.State.Mutex.Unlock()

    if len(incomplete) == 0 {
        return nil // No incomplete parts to restart
    }

    jobs := make(chan chunk, h.WORKERS_COUNT)
    errChan := make(chan error, h.WORKERS_COUNT)
    pauseAck := make(chan bool, h.WORKERS_COUNT)

    var wg sync.WaitGroup
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, jobs, errChan, pauseAck, &wg)
    }

    go func() {
        for _, chunk := range incomplete {
            jobs <- chunk
        }
        close(jobs)
    }()

    wg.Wait()
    close(errChan)
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    return nil
}