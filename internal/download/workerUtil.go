package download

import (
	"fmt"
    "sync"
)

func (h *DownloadHandler) worker(id int, jobs <-chan chunk, errChan chan<- error, pauseAck chan<- bool, wg *sync.WaitGroup) {
    defer wg.Done()

    // we will iterae on jobs/chunks on channel
    for chunk := range jobs {
        // // Check pause state
        // h.State.Mutex.Lock()
        // if h.State.IsPaused {
        //     h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
        //     h.State.Mutex.Unlock()
        //     fmt.Printf("Worker %d stopped due to pause state\n", id)
        //     pauseAck <- true
        //     return
        // }
        // h.State.Mutex.Unlock()

        select {
        // case <-h.PauseChan:
        //     h.State.Mutex.Lock()
        //     h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
        //     h.State.Mutex.Unlock()
        //     fmt.Printf("Worker %d paused at chunk %d-%d\n", id, chunk.Start, chunk.End)
        //     pauseAck <- true
        //     return
        default: // we should start downloading assigned chunk

            partIndex := chunk.Start / h.CHUNK_SIZE // later we will use it for path and stuff
            // try downloadWithrange starting from chunks start till end of the chunk
            if err := h.downloadWithRanges(chunk.Start, chunk.End); err != nil {
                fmt.Printf("Worker %d: Failed chunk %d-%d: %v\n", id, chunk.Start, chunk.End, err)
                h.State.Mutex.Lock()
                h.State.IncompleteParts = append(h.State.IncompleteParts, chunk) // Requeue failed chunk
                // Ensure the part is not marked as completed
                if int(partIndex) < len(h.State.Completed) {
                    h.State.Completed[partIndex] = false
                }
                h.State.Mutex.Unlock()
                errChan <- fmt.Errorf("worker %d failed: %v", id, err)
                continue
            }

            fmt.Printf("Worker %d: Successfully downloaded chunk %d-%d\n", id, chunk.Start, chunk.End)
            h.State.Mutex.Lock()
            if int(partIndex) < len(h.State.Completed) {
                h.State.Completed[partIndex] = true
                chunkSize := int64(chunk.End - chunk.Start + 1)
                h.State.CurrentByte += chunkSize
                // h.Progress.UpdateBytesDone(h.State.CurrentByte)
            }
            h.State.Mutex.Unlock()
        }
    }
	fmt.Printf("Worker %d: Finished task\n", id)
    // pauseAck <- true
}

func (h *DownloadHandler) distributeJobs(jobs chan<- chunk, contentLength int) {
    currentByte := h.State.CurrentByte
    for currentByte < int64(contentLength) {
        // // Check pause state first
        // h.State.Mutex.Lock()
        // if h.State.IsPaused {
        //     h.State.CurrentByte = currentByte
        //     h.State.Mutex.Unlock()
        //     return
        // }
        // h.State.Mutex.Unlock()

        // we will add up current byte with chunk_size to get the end
        end := currentByte + int64(h.CHUNK_SIZE)
        if end > int64(contentLength) {
            end = int64(contentLength)
        }

        // defining the chunk
        chunk := chunk{Start: currentByte, End: end - 1}

		jobs <- chunk
		fmt.Printf("Dispatched chunk %d-%d\n", chunk.Start, chunk.End)
		currentByte = end
        
    }
}