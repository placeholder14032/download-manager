package download

import(
	"net/http"
	"time"
	"sync"
	"fmt"
	"os"
	"io"
)

type DownloadHandler struct {
    Client        *http.Client
    CHUNK_SIZE    int
    WORKERS_COUNT int
    PartsCount    int
    PauseChan     chan struct{} 

    DELTA         time.Duration
    // Progress      *ProgressTracker

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

// Initializing 
func (download *Download) NewDownloadHandler(client *http.Client, chunkSize int, workersCount int, bandsWidth int64) *DownloadHandler {
    dh := &DownloadHandler{
        Client:            client,
        CHUNK_SIZE:        chunkSize,
        WORKERS_COUNT:     workersCount,
        PauseChan:        make(chan struct{}),
        URL:              download.URL,
        FilePath:         download.FilePath,
        DELTA:            time.Second,
        // Progress:         NewProgressTracker(0, time.Second),
        State:            &DownloadState{}, 
        // BandWidth:        bandsWidth,
        // IsBandWidthLimited: bandsWidth > 0,
    }
    return dh
}


func (h *DownloadHandler) StartDownloading() error {
	// First, we will check if the server supports range requests or not -> using our IsAcceptRangeSupported() method
	
    supportsRange, contentLength, err := h.IsAcceptRangeSupported()
		fmt.Print(supportsRange,contentLength)

    if err != nil {
        return err
    }
    // h.Progress.TotalBytes = int64(contentLength)

	// If the server does not support range requests, we will download the file without using range requests
    if (!supportsRange) {
        return h.downloadWithoutRanges()
    }

    // h.PartsCount = (contentLength + h.CHUNK_SIZE - 1) / h.CHUNK_SIZE
    // h.State.Completed = make([]bool, h.PartsCount)
    // h.State.TotalBytes = int64(contentLength)



    // // jobs: it's a channel used to send chunks to worker "task to download a specific piece (or "chunk")""
    // jobs := make(chan chunk, h.WORKERS_COUNT)    // sends chunk information to workers
    // errChan := make(chan error, h.WORKERS_COUNT)
    // done := make(chan bool, 1) // used to notify if it's done or not
    // pauseAck := make(chan bool, h.WORKERS_COUNT) // channel to acknowledge worker pause completion


    // var wg sync.WaitGroup
    // for i := 0; i < h.WORKERS_COUNT; i++ {
    //     wg.Add(1)
    //     go h.worker(i, jobs, errChan, pauseAck, &wg)
    // }

    // h.startWorkers( &wg, jobs, errChan, pauseAck)
    // go h.distributeJobs(jobs, contentLength)
    // go h.waitForCompletion(&wg, errChan, done)


    // select {
    // case <-done:
    //     fmt.Print("case done in select startibg download")
    //     return h.handleDownloadCompletion(contentLength, errChan, done)
    // case err := <-errChan:
    //     return err
    // }

	return nil
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