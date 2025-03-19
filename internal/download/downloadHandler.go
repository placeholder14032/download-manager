package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type DownloadHandler struct {
    Client        *http.Client
    CHUNK_SIZE    int
    WORKERS_COUNT int
    PartsCount    int
    PauseChan     chan struct{} 

    DELTA         time.Duration
    Progress      *ProgressTracker

    URL           string
    FilePath      string
    State         *DownloadState
    BandWidth     int64
    IsBandWidthLimited bool
}

type DownloadState struct {
    IncompleteParts []chunk
    Completed       []bool
    CurrentByte     int64
    TotalBytes      int64
    Mutex           sync.Mutex
    IsPaused        bool
    isCombined      bool   
}

type chunk struct {
    Start int
    End   int
}


func (download *Download) NewDownloadHandler(client *http.Client, chunkSize int, workersCount int, bandsWidth int64) *DownloadHandler {
    dh := &DownloadHandler{
        Client:            client,
        CHUNK_SIZE:        chunkSize,
        WORKERS_COUNT:     workersCount,
        PauseChan:        make(chan struct{}),
        URL:              download.URL,
        FilePath:         download.FilePath,
        DELTA:            time.Second,
        Progress:         NewProgressTracker(0, time.Second),
        State:            &DownloadState{}, // Initialize State
        BandWidth:        bandsWidth,
        IsBandWidthLimited: bandsWidth > 0,
    }
    return dh
}

func (h *DownloadHandler) initializeState(contentLength int) {
    h.PauseChan = make(chan struct{})
}


func (h *DownloadHandler) StartDownloading() error {
    supportsRange, contentLength, err := h.IsAcceptRangeSupported()
    if err != nil {
        return err
    }
    h.Progress.TotalBytes = int64(contentLength)

    if (!supportsRange) {
        return h.downloadWithoutRanges()
    }

    h.PartsCount = (contentLength + h.CHUNK_SIZE - 1) / h.CHUNK_SIZE
   h.State.Completed = make([]bool, h.PartsCount)
   h.State.TotalBytes = int64(contentLength)
    h.initializeState(contentLength)

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

    h.startWorkers( &wg, jobs, errChan, pauseAck)
    go h.distributeJobs(jobs, contentLength)
    go h.waitForCompletion(&wg, errChan, done)

    return h.handleDownloadCompletion( contentLength, errChan, done)
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

func (h *DownloadHandler) downloadWithRanges(start int, end int) error {
    const maxRetries = 3
    var lastErr error

    for attempt := 1; attempt <= maxRetries; attempt++ {
        req, err := http.NewRequest("GET", h.URL, nil)
        if err != nil {
            return fmt.Errorf("failed to create request: %v", err)
        }
        rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
        req.Header.Add("Range", rangeHeader)

        resp, err := h.Client.Do(req)
        if err != nil {
            lastErr = fmt.Errorf("attempt %d: failed to execute request: %v", attempt, err)
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusPartialContent {
            lastErr = fmt.Errorf("attempt %d: server returned unexpected status: %d", attempt, resp.StatusCode)
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }

        expectedSize := end - start + 1
        if resp.ContentLength != int64(expectedSize) {
            lastErr = fmt.Errorf("attempt %d: server returned wrong content length: got %d, want %d", attempt, resp.ContentLength, expectedSize)
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }

        partNumber := start / h.CHUNK_SIZE
        partFileName := fmt.Sprintf("%s.part%d", h.FilePath, partNumber)
        
        file, err := os.OpenFile(partFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
        if err != nil {
            return fmt.Errorf("failed to create part file %s: %v", partFileName, err)
        }

        // Use a custom reader to count actual bytes read from the response
        var totalRead int64
        reader := &countingReader{reader: resp.Body, count: &totalRead}
        written, err := io.Copy(file, reader)
        if err != nil {
            file.Close()
            lastErr = fmt.Errorf("attempt %d: failed to write chunk: %v", attempt, err)
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }

        if written != int64(expectedSize) {
            file.Close()
            lastErr = fmt.Errorf("attempt %d: incomplete write: wrote %d bytes, expected %d bytes", attempt, written, expectedSize)
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }

        if totalRead != int64(expectedSize) {
            file.Close()
            lastErr = fmt.Errorf("attempt %d: read %d bytes from server, expected %d bytes", attempt, totalRead, expectedSize)
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }

        if err := file.Sync(); err != nil {
            file.Close()
            return fmt.Errorf("failed to sync part file %s: %v", partFileName, err)
        }

        if err := file.Close(); err != nil {
            return fmt.Errorf("failed to close part file %s: %v", partFileName, err)
        }

        // First size check
        info, err := os.Stat(partFileName)
        if err != nil {
            return fmt.Errorf("failed to stat part file %s after close: %v", partFileName, err)
        }
        if info.Size() != int64(expectedSize) {
            lastErr = fmt.Errorf("attempt %d: file size mismatch after close: got %d, want %d", attempt, info.Size(), expectedSize)
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }

        // Wait briefly and re-check to catch file system inconsistencies
        time.Sleep(100 * time.Millisecond)
        info, err = os.Stat(partFileName)
        if err != nil {
            return fmt.Errorf("failed to re-stat part file %s after delay: %v", partFileName, err)
        }
        if info.Size() != int64(expectedSize) {
            lastErr = fmt.Errorf("attempt %d: file size changed after delay: got %d, want %d", attempt, info.Size(), expectedSize)
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }

        fmt.Printf("downloadWithRanges: Completed part %d, wrote %d bytes (read from server: %d bytes, verified on disk: %d bytes)\n", 
            partNumber, written, totalRead, info.Size())
        return nil
    }

    return fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

// countingReader wraps an io.Reader to count the total bytes read
type countingReader struct {
    reader io.Reader
    count  *int64
}

func (r *countingReader) Read(p []byte) (n int, err error) {
    n, err = r.reader.Read(p)
    *r.count += int64(n)
    return n, err
}

func (h *DownloadHandler) MonitorProgress(done chan bool) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-done:
            return
        case <-ticker.C:
           h.State.Mutex.Lock()
            currentProgress := h.Progress.GetProgress()
            if !h.State.IsPaused {
                fmt.Printf("\rProgress: %.2f%%", currentProgress)
            }
           h.State.Mutex.Unlock()
        }
    }
}