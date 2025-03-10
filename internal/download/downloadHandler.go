package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type DownloadHandler struct {
	Client        *http.Client
	CHUNK_SIZE    int
	WORKERS_COUNT int
	paused        bool
	lastByte      int // name?
    partsCount   int 
}

type chunk struct {
    start int
    end   int
}

// to do: 
// 3) add dd cleanup functionality to remove the temporary .part files after successful combination
// 4) implement pause/resume functionality (the `paused` flag is already there but needs to be properly handled)

func (h *DownloadHandler) isAcceptRangeSupported(download Download) (bool, int) {
	var url = download.URL
	req, _ := http.NewRequest("HEAD", url, nil)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		fmt.Println("Cannot continue download, status code " + fmt.Sprint(resp.StatusCode))
		return false, 0
	}

	acceptRanges := strings.ToLower(resp.Header.Get("Accept-Ranges"))
	if acceptRanges == "" || acceptRanges == "none" {
		return false, int(resp.ContentLength)
	}

	return true, int(resp.ContentLength)
}

func (h *DownloadHandler) StartDownloading(d Download) error {
    supportsRange, contentLength := h.isAcceptRangeSupported(d)

    if (!supportsRange) {
        return h.downloadWithoutRanges(d, contentLength)
    }

	// we have 3 channels one for errors one for jobs (used to send download chunks to worker goroutines) and one for done
    jobs := make(chan chunk, h.WORKERS_COUNT) // sends chunk information to workers
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool)

    h.partsCount = (contentLength + h.CHUNK_SIZE - 1) / h.CHUNK_SIZE
    fmt.Printf("Starting download with %d parts\n", h.partsCount)

    var wg sync.WaitGroup
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, &d, jobs, errChan, &wg)
    }

    // managing what work to give to workers
    go func() {
        currentByte := 0
        partsNum := 0
        for currentByte < contentLength {
            if h.paused {
                break
            }

            end := currentByte + h.CHUNK_SIZE
            if end > contentLength {
                end = contentLength
            }

            fmt.Printf("Scheduling part %d (bytes %d-%d)\n", partsNum, currentByte, end-1)
            jobs <- chunk{
                start: currentByte,
                end:   end - 1,
            }
            partsNum++
            currentByte = end
        }
        fmt.Printf("All %d parts scheduled for download\n", partsNum)
        close(jobs)
    }()

    // wiat for workers and handle errors
    go func() {
        wg.Wait()
        fmt.Println("All workers are done")
        close(errChan)
        done <- true
    }()

    // handle errors
    select {
    case err := <-errChan:
        if err != nil {
            return err
        }
        fmt.Println("Download completed, starting to combine parts...")
        if err := h.combineParts(&d, contentLength); err != nil {
            return fmt.Errorf("failed to combine parts: %v", err)
        }
        fmt.Println("Parts combined successfully")
        return nil
    case <-done:
        // combine parts if the doenload was successful
        fmt.Println("Download completed, starting to combine parts...")
        if err := h.combineParts(&d, contentLength); err != nil {
            return fmt.Errorf("failed to combine parts: %v", err)
        }
        fmt.Println("Parts combined successfully")
        return nil
    }

    return nil
}

func (h *DownloadHandler) worker(id int, d *Download, jobs <-chan chunk, errChan chan<- error, wg *sync.WaitGroup) {
    defer wg.Done()

    for chunk := range jobs {
        if h.paused {
            return
        }

        if err := h.downloadWithRanges(d, chunk.start, chunk.end); err != nil {
            errChan <- fmt.Errorf("worker %d failed: %v", id, err)
            return
        }
    }
}

func (h *DownloadHandler) downloadWithoutRanges(d Download, contentLength int) error {
    req, err := http.NewRequest("GET", d.URL, nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %v", err)
    }

    resp, err := h.Client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to execute request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("server returned error status: %d", resp.StatusCode)
    }

    file, err := os.Create(d.FilePath)
    if err != nil {
        return fmt.Errorf("failed to create file: %v", err)
    }
    defer file.Close()

    _, err = io.Copy(file, resp.Body)
    if err != nil {
        return fmt.Errorf("failed to download file: %v", err)
    }

    return nil
}

func (h *DownloadHandler) downloadWithRanges(download *Download, start int, end int) error {
	req, err := http.NewRequest("GET", download.URL, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

    // Set range header
    rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
    req.Header.Add("Range", rangeHeader)

	resp, err := h.Client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
    defer resp.Body.Close()

    // Create part file name
    partNumber := start / h.CHUNK_SIZE
    partFileName := fmt.Sprintf("%s.part%d", download.FilePath, partNumber)
    file, err := os.Create(partFileName)
    if err != nil {
        fmt.Println(err)
        return err
    }
    defer file.Close()

    // Copy the response body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (h *DownloadHandler) combineParts(download *Download, contentLength int) error {
    fmt.Printf("Starting to combine %d parts\n", h.partsCount)
    // the final file
    combinedFile, err := os.Create(download.FilePath)
    if err != nil {
        return fmt.Errorf("failed to create the final file: %v", err)
    }
    defer combinedFile.Close()

    combinedParts := make([]bool, h.partsCount)
    
    //combining parts in order
    buffer := make([]byte, 32*1024)
    for i := 0; i < h.partsCount; i++ {
        partFileName := fmt.Sprintf("%s.part%d", download.FilePath, i)
        
        partFile, err := os.Open(partFileName)
        if err != nil {
            return fmt.Errorf("failed to open part file %s: %v", partFileName, err)
        }


        written, err := io.CopyBuffer(combinedFile, partFile, buffer)
        partFile.Close() 

        if err != nil {
            return fmt.Errorf("failed to copy part %d: %v", i, err)
        }

        if written > 0 {
            combinedParts[i] = true
        }

        // delte part file after successful copy
        if err := os.Remove(partFileName); err != nil {
            fmt.Printf("Warning: failed to remove part file %s: %v\n", partFileName, err)
        }
    }

    // make sure all parts were combined
    for i, combined := range combinedParts {
        if !combined {
            return fmt.Errorf("part %d was not combined successfully", i)
        }
    }

    return nil
}