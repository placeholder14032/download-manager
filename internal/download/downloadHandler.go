package download

import (
	"fmt"
	"net/http"
	"strings"
	"time"
    "os"
    "io"
	"sync"
)

type DownloadHandler struct {
	client        *http.Client
	CHUNK_SIZE    int
	WORKERS_COUNT int
	paused        bool
	lastByte      int // name?
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

	if !supportsRange {
		h.downloadWithoutRanges(d, contentLength)
		return nil
	}
	
	var wg sync.WaitGroup
	errChan := make(chan error, h.WORKERS_COUNT)
	
	// Calculate chunk sizes
	chunksPerWorker := contentLength / (h.CHUNK_SIZE * h.WORKERS_COUNT)
	
	for i := 0; i < h.WORKERS_COUNT; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			startByte := workerID * chunksPerWorker * h.CHUNK_SIZE
			endByte := startByte + (chunksPerWorker * h.CHUNK_SIZE) - 1
			
			if workerID == h.WORKERS_COUNT-1 {
				endByte = contentLength - 1
			}
			
			for currentByte := startByte; currentByte <= endByte; currentByte += h.CHUNK_SIZE {
				if h.paused {
					return
				}
				
				end := currentByte + h.CHUNK_SIZE - 1
				if end > endByte {
					end = endByte
				}
				
				if err := h.downloadWithRanges(&d, currentByte, end); err != nil {
					errChan <- err
					return
				}
			}
		}(i)
	}

	// Wait for all workers and collect errors
	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	
	return nil
}

func (h *DownloadHandler) downloadWithoutRanges(d Download, contentLength int) {
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
