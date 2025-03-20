package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"context"
	"time"
)

type DownloadHandler struct {
    Client        *http.Client
    CHUNK_SIZE    int64
    WORKERS_COUNT int
    PartsCount    int64

    URL           string
    FilePath      string
    State         *DownloadState

	ctx           context.Context    // Add context
    cancel        context.CancelFunc // Add cancel function
    PauseChan     chan struct{}      // Keep this for signaling pause
    ResumeChan    chan struct{}      // Add channel for resume
}

type DownloadState struct {
    IncompleteParts []chunk
    Completed       []bool
    CurrentByte     int64
    TotalBytes      int64
    Mutex           sync.Mutex
    IsPaused        bool
}

type chunk struct {
    Start int64
    End   int64
}

// Initializing 
func (download *Download) NewDownloadHandler(client *http.Client, chunkSize int64, workersCount int, bandsWidth int64) *DownloadHandler {
	ctx, cancel := context.WithCancel(context.Background())
    dh := &DownloadHandler{
        Client:            client,
        CHUNK_SIZE:        chunkSize,
        WORKERS_COUNT:     workersCount,
        URL:              download.URL,
        FilePath:         download.FilePath,
        State:            &DownloadState{}, 

		PauseChan:     make(chan struct{}),
        ResumeChan:    make(chan struct{}),
		ctx:           ctx,
        cancel:        cancel,
    }
    return dh
}


func (h *DownloadHandler) StartDownloading() error {
	// First, we will check if the server supports range requests or not -> using our IsAcceptRangeSupported() method
    supportsRange, contentLength, err := h.IsAcceptRangeSupported()
    if err != nil {
        return err
    }

	// If the server does not support range requests, we will download the file without using range requests
    if (!supportsRange) {
        return h.downloadWithoutRanges()
    }

    h.PartsCount = (contentLength + h.CHUNK_SIZE - 1) / h.CHUNK_SIZE
    h.State.Completed = make([]bool, h.PartsCount)
    h.State.TotalBytes = int64(contentLength)

    // jobs: it's a channel used to send chunks to worker "task to download a specific piece (or "chunk")""
    jobs := make(chan chunk, h.WORKERS_COUNT)    // sends chunk information to workers
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool, 1) // used to notify if it's done or not
    pauseAck := make(chan bool, h.WORKERS_COUNT) // channel to acknowledge worker pause completion


    var wg sync.WaitGroup
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, jobs, errChan, pauseAck, &wg)
    }

	go func() {
		h.distributeJobs(jobs, int(contentLength))
		// close(jobs) // close jobs channel after all tasks are sent 
	}()

	// waiting for workers to be done
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
        fmt.Println("Calling combineParts")
        return h.combineParts(contentLength)
    case err := <-errChan:
		close(jobs)
        return err
    }
}

func (h *DownloadHandler) downloadWithoutRanges() error {
    req, err := http.NewRequest("GET",h.URL, nil)
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

    file, err := os.Create(h.FilePath)
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

func (h *DownloadHandler) downloadWithRanges(start int64, end int64) error {
	// defining expected sixe and stuff we will use later 
	expectedSize := end - start + 1
    partNumber := start / h.CHUNK_SIZE
    partFileName := fmt.Sprintf("%s.part%d", h.FilePath, partNumber)
    fmt.Printf("Worker starting download for chunk %d-%d at %s\n", start, end, time.Now().Format(time.RFC3339))


	// creating request for server
	req, err := http.NewRequest("GET", h.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	// executing the request
    resp, err := h.Client.Do(req)
    if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
    }
    defer resp.Body.Close()

	// validating response:
	// server status
    if resp.StatusCode != http.StatusPartialContent {
        return fmt.Errorf("server returned unexpected status: %d", resp.StatusCode)
    }
	// making sure server is returning expectedSize
    if resp.ContentLength != expectedSize {
        return fmt.Errorf("server returned wrong content length: got %d, want %d", resp.ContentLength, expectedSize)
    }

    // creating file we will write the chunk on
    file, err := os.OpenFile(partFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
    if err != nil {
        return fmt.Errorf("failed to create part file %s: %v", partFileName, err)
    }
	defer file.Close()

    // now we will use a custom reader to count actual bytes read from the response -> make sure we are reading and writing all bytes
	var totalRead int64
    reader := &countingReader{reader: resp.Body, count: &totalRead}
    written, err := io.Copy(file, reader)
    if err != nil {
        return fmt.Errorf("failed to write chunk: %v", err)
    }
    if totalRead != expectedSize {
        return fmt.Errorf("read %d bytes from server, expected %d bytes", totalRead, expectedSize)
    }

    // ennsuring file is properly written
    if err := file.Sync(); err != nil {
        return fmt.Errorf("failed to sync part file %s: %v", partFileName, err)
    }

    // verifying file size after closing
    if info, err := os.Stat(partFileName); err != nil {
        return fmt.Errorf("failed to stat part file %s: %v", partFileName, err)
    } else if info.Size() != expectedSize {
        return fmt.Errorf("file size mismatch after close: got %d, want %d", info.Size(), expectedSize)
    }

    fmt.Printf("Completed part %d, wrote %d bytes (read from server: %d bytes, verified on disk: %d bytes)\n",
        partNumber, written, totalRead, expectedSize)

    return nil
}