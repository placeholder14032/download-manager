package download

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// start Downloading function parts
func (h *DownloadHandler) startWorkers( wg *sync.WaitGroup, jobs <-chan chunk, errChan chan<- error, pauseAck chan<- bool) {
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, jobs, errChan, pauseAck, wg)
    }
}

func (h *DownloadHandler) distributeJobs(jobs chan<- chunk, contentLength int) {
    defer close(jobs)
    currentByte := h.State.CurrentByte

    for currentByte < int64(contentLength) {
        // Check pause state first
        h.State.Mutex.Lock()
        if h.State.IsPaused {
            h.State.CurrentByte = currentByte
            h.State.Mutex.Unlock()
            return
        }
        h.State.Mutex.Unlock()

        end := currentByte + int64(h.CHUNK_SIZE)
        if end > int64(contentLength) {
            end = int64(contentLength)
        }

        chunk := chunk{Start: int(currentByte), End: int(end - 1)}
        
        select {
        case <-h.PauseChan:
            h.State.Mutex.Lock()
            h.State.CurrentByte = currentByte
            h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
            h.State.Mutex.Unlock()
            fmt.Printf("Distributor paused at byte %d\n", currentByte)
            return
        case jobs <- chunk:
            currentByte = end
            fmt.Printf("Dispatched chunk %d-%d\n", chunk.Start, chunk.End)
        }
    }
}

func (h *DownloadHandler) waitForCompletion(wg *sync.WaitGroup, errChan chan<- error, done chan<- bool) {
    wg.Wait()
    if !h.State.IsPaused {
        close(errChan)
        done <- true
    }
}

func (h *DownloadHandler) handleDownloadCompletion(contentLength int, errChan <-chan error, done <-chan bool) error {
    select {
    case err := <-errChan:
        if err != nil {
            return err
        }
        if !h.State.IsPaused {
            return h.combineParts( contentLength)
        }
        return nil
    case <-done:
        return h.combineParts( contentLength)
    }
}

func (h *DownloadHandler) worker(id int, jobs <-chan chunk, errChan chan<- error, pauseAck chan<- bool, wg *sync.WaitGroup) {
    defer wg.Done()

    for chunk := range jobs {
        // Check pause state
        h.State.Mutex.Lock()
        if h.State.IsPaused {
            h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
            h.State.Mutex.Unlock()
            fmt.Printf("Worker %d stopped due to pause state\n", id)
            pauseAck <- true
            return
        }
        h.State.Mutex.Unlock()

        select {
        case <-h.PauseChan:
            h.State.Mutex.Lock()
            h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
            h.State.Mutex.Unlock()
            fmt.Printf("Worker %d paused at chunk %d-%d\n", id, chunk.Start, chunk.End)
            pauseAck <- true
            return
        default:
            if err := h.downloadWithRanges(chunk.Start, chunk.End); err != nil {
                fmt.Printf("Worker %d: Failed chunk %d-%d: %v\n", id, chunk.Start, chunk.End, err)
                h.State.Mutex.Lock()
                h.State.IncompleteParts = append(h.State.IncompleteParts, chunk) // Requeue failed chunk
                h.State.Mutex.Unlock()
                errChan <- fmt.Errorf("worker %d failed: %v", id, err)
                continue // Don’t exit, keep processing
            }
            fmt.Printf("Worker %d: Successfully downloaded chunk %d-%d\n", id, chunk.Start, chunk.End)
            h.State.Mutex.Lock()
            partIndex := chunk.Start / h.CHUNK_SIZE
            if partIndex < len(h.State.Completed) {
                h.State.Completed[partIndex] = true
                chunkSize := int64(chunk.End - chunk.Start + 1)
                h.State.CurrentByte += chunkSize
                h.Progress.UpdateBytesDone(h.State.CurrentByte)
            }
            h.State.Mutex.Unlock()
        }
    }
    fmt.Printf("Worker %d: Exiting normally\n", id)
    pauseAck <- true
}

func (h *DownloadHandler) combineParts( contentLength int) error {
    c :=  NewPartsCombiner()
    return c.CombineParts(h.FilePath, contentLength, h.PartsCount)
}

func (h *DownloadHandler) IsAcceptRangeSupported() (bool, int, error) {
    req, err := http.NewRequest("HEAD", h.URL, nil)
    if err != nil {
        return false, 0, fmt.Errorf("failed to create HEAD request: %v", err)
    }

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return false, 0, fmt.Errorf("HEAD request failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return false, 0, fmt.Errorf("server returned status: %d", resp.StatusCode)
    }

    acceptRanges := strings.ToLower(resp.Header.Get("Accept-Ranges"))
    if acceptRanges == "" || acceptRanges == "none" {
        return false, int(resp.ContentLength), nil
    }
    return true, int(resp.ContentLength), nil
}