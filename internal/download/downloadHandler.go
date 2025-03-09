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
	client        *http.Client
	CHUNK_SIZE    int
	WORKERS_COUNT int
	paused        bool
	lastByte      int // name?
}

type chunk struct {
    start int
    end   int
}

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

func (h *DownloadHandler) startDownloading(d Download) error {
    supportsRange, contentLength := h.isAcceptRangeSupported(d)

    if (!supportsRange) {
        return h.downloadWithoutRanges(d, contentLength)
    }

	// we have 3 channels one for errors one for jobs (used to send download chunks to worker goroutines) and one for done
    jobs := make(chan chunk, h.WORKERS_COUNT) // sends chunk information to workers
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool)


    var wg sync.WaitGroup
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, &d, jobs, errChan, &wg)
    }

    // managing what work to give to workers
    go func() {
        currentByte := 0
        for currentByte < contentLength {
            if h.paused {
                break
            }

            end := currentByte + h.CHUNK_SIZE
            if end > contentLength {
                end = contentLength
            }

            jobs <- chunk{
                start: currentByte,
                end:   end - 1,
            }

            currentByte = end
        }
        close(jobs)
    }()

    // wiat for workers and handle errors
    go func() {
        wg.Wait()
        close(errChan)
        done <- true
    }()

    // handle errors
    select {
    case err := <-errChan:
        if err != nil {
            return err
        }
    case <-done:
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
	panic("unimplemented")
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

	resp, err := h.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
    defer resp.Body.Close()

    // Create part file name
    partFileName := fmt.Sprintf("%s.part%d", download.FilePath, start)
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
