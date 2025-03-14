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
    req, err := http.NewRequest("GET",h.URL, nil)
    if (err != nil) {
        fmt.Println(err)
        return err
    }

    // settigng range header
    rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
    req.Header.Add("Range", rangeHeader)

    resp, err := h.Client.Do(req)
    if (err != nil) {
        fmt.Println(err)
        return err
    }
    defer resp.Body.Close()

    // create part files, we are writing in parts and combine it later
    partNumber := start / h.CHUNK_SIZE
    partFileName := fmt.Sprintf("%s.part%d",h.FilePath, partNumber)
    file, err := os.Create(partFileName)
    if (err != nil) {
        fmt.Println(err)
        return err
    }
    defer file.Close()

    // Bandwidth-limited download
    const bufferSize = 8192
    buf := make([]byte, bufferSize)
    startTime := time.Now()
    var bytesDownloaded int64

    for {
        select {
        case <-h.PauseChan:
            return nil
        default:
            n, err := resp.Body.Read(buf)
            if n > 0 {
                if _, err := file.Write(buf[:n]); err != nil {
                    return fmt.Errorf("failed to write chunk: %v", err)
                }
                
                bytesDownloaded += int64(n)
                h.State.Mutex.Lock()
                h.State.CurrentByte += int64(n)
                currentTotal := h.State.CurrentByte
                if currentTotal > h.State.TotalBytes {
                    currentTotal = h.State.TotalBytes
                }
                h.State.Mutex.Unlock()
                
                h.Progress.UpdateBytesDone(currentTotal)

                // Bandwidth control
                if h.IsBandWidthLimited {
                    elapsed := time.Since(startTime).Seconds()
                    expectedBytes := int64(elapsed * float64(h.BandWidth))
                    // if we downloaded more than bandswidth we need to sleep and wait for next tick
                    if bytesDownloaded > expectedBytes {
                        sleepTime := time.Duration(float64(bytesDownloaded-expectedBytes)/float64(h.BandWidth)*1000) * time.Millisecond
                        time.Sleep(sleepTime)
                    }
                }
            }
            if err == io.EOF {
                // update state as we failed
                h.State.Mutex.Lock()
                h.State.CurrentByte += int64(h.CHUNK_SIZE)
                h.Progress.UpdateBytesDone(h.State.CurrentByte)
                h.State.Mutex.Unlock()
                return nil
            }
            if err != nil {
                return fmt.Errorf("failed to read chunk: %v", err)
            }
        }
    }
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