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
		fmt.Println("Ignoring pause request - download already complete")
		return
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
	close(h.ResumeChan)
	h.ResumeChan = make(chan struct{})

	return h.restartDownload()
}

func (h *DownloadHandler) restartDownload() error {
	jobs := make(chan chunk, h.WORKERS_COUNT)
	errChan := make(chan error, h.WORKERS_COUNT)
	done := make(chan bool, 1)
	pauseAck := make(chan bool, h.WORKERS_COUNT)

	var wg sync.WaitGroup
	for i := 0; i < h.WORKERS_COUNT; i++ {
		wg.Add(1)
		go h.worker(i, jobs, errChan, pauseAck, &wg)
	}

	go func() {
		defer close(jobs)
		h.State.Mutex.Lock()
		incomplete := h.State.IncompleteParts
		h.State.IncompleteParts = nil
		h.State.Mutex.Unlock()

		for _, chunk := range incomplete {
			jobs <- chunk
		}

		if h.State.CurrentByte < h.State.TotalBytes {
			h.distributeRemainingJobs(jobs)
		}
	}()

	go func() {
		wg.Wait()
		h.State.Mutex.Lock()
		if h.State.CurrentByte >= h.State.TotalBytes {
			done <- true
		}
		h.State.Mutex.Unlock()
	}()

	select {
	case <-done:
		close(errChan)
		for err := range errChan {
			if err != nil {
				return err
			}
		}
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