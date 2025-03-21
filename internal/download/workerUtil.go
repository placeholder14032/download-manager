package download

import (
	"fmt"
    "sync"
)

func (h *DownloadHandler) worker(id int, jobs <-chan chunk, errChan chan<- error, pauseAck chan<- bool, wg *sync.WaitGroup) {
    defer wg.Done()

    // we will iterae on jobs/chunks on channel
    for chunk := range jobs {
        select {
			case <-h.ctx.Done(): // Handle cancellation/pause
				h.State.Mutex.Lock()
            	h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
            	h.State.Mutex.Unlock()
				fmt.Printf("Worker %d paused at chunk %d-%d\n", id, chunk.Start, chunk.End)
				pauseAck <- true
				return // Exit immediately on cancel
			case <-h.PauseChan:
				h.State.Mutex.Lock()
				h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
				h.State.Mutex.Unlock()
				fmt.Printf("Worker %d paused at chunk %d-%d\n", id, chunk.Start, chunk.End)
				pauseAck <- true
				<-h.ResumeChan // Wait for resume signal
				continue       // Reprocess this chunk after resume
        default:    // Process the chunk normally
			// we should start downloading assigned chunk
            partIndex := chunk.Start / h.CHUNK_SIZE // later we will use it for path and stuff
            // try downloadWithrange starting from chunks start till end of the chunk

			h.State.Mutex.Lock()
			if int(partIndex) < len(h.State.Completed) && h.State.Completed[partIndex] {
				h.State.Mutex.Unlock()
				continue // Skip already completed chunk
			}
			h.State.Mutex.Unlock()
			
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
                return // exit on error
            }

            fmt.Printf("Worker %d: Successfully downloaded chunk %d-%d\n", id, chunk.Start, chunk.End)
            h.State.Mutex.Lock()
            if int(partIndex) < len(h.State.Completed) && !h.State.Completed[partIndex] { // Only increment if not completed
                h.State.Completed[partIndex] = true
                chunkSize := int64(chunk.End - chunk.Start + 1)
                h.State.CurrentByte += chunkSize
            }
            h.State.Mutex.Unlock()
        }
    }
	fmt.Printf("Worker %d: Finished task\n", id)
}

func (h *DownloadHandler) distributeJobs(jobs chan<- chunk, contentLength int) {
    currentByte := h.State.CurrentByte
    for currentByte < int64(contentLength) {
        select {
        case <-h.ctx.Done():
            return // Exit on pause without closing jobs
		default:
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
	if currentByte >= int64(contentLength) {
        close(jobs)
    }
}

