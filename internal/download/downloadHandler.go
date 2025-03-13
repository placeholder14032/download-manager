package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

type DownloadHandler struct {
    Client        *http.Client
    CHUNK_SIZE    int
    WORKERS_COUNT int
    PartsCount    int
    State         *DownloadState
    PauseChan     chan struct{} 
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


func (download *Download) NewDownloadHandler(client *http.Client, chunkSize int, workersCount int) *DownloadHandler {
    return &DownloadHandler{
        Client:        client,
        CHUNK_SIZE:    chunkSize,
        WORKERS_COUNT: workersCount,
        PauseChan:     make(chan struct{}),
        State:         &DownloadState{
            Completed:       make([]bool, 0),
            IncompleteParts: make([]chunk, 0),
            CurrentByte:     0,
            TotalBytes:      0,
            Mutex:           sync.Mutex{},
            isCombined:      false,
        },
    }
}

func (h *DownloadHandler) initializeState(contentLength int) {
    h.State = &DownloadState{
        Completed:       make([]bool, h.PartsCount),
        IncompleteParts: make([]chunk, 0),
        CurrentByte:     0,
        TotalBytes:      int64(contentLength),
        Mutex:           sync.Mutex{},
        IsPaused:        false,
        isCombined:      false,
    }
    h.PauseChan = make(chan struct{})
}


func (h *DownloadHandler) StartDownloading(d Download) error {
    supportsRange, contentLength, err := h.IsAcceptRangeSupported(d)
    if err != nil {
        return err
    }
    if !supportsRange {
        return h.downloadWithoutRanges(d)
    }

    h.PartsCount = (contentLength + h.CHUNK_SIZE - 1) / h.CHUNK_SIZE
    h.initializeState(contentLength)

    // jobs: it's a channel used to send chunks to worker "task to download a specific piece (or "chunk")""
    jobs := make(chan chunk, h.WORKERS_COUNT)    // sends chunk information to workers
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool, 1) // used to notify if it's done or not
    pauseAck := make(chan bool, h.WORKERS_COUNT) // channel to acknowledge worker pause completion


    var wg sync.WaitGroup
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, &d, jobs, errChan, pauseAck, &wg)
    }

    h.startWorkers(&d, &wg, jobs, errChan, pauseAck)
    go h.distributeJobs(jobs, contentLength)
    go h.waitForCompletion(&wg, errChan, done)

    return h.handleDownloadCompletion(&d, contentLength, errChan, done)
}

func (h *DownloadHandler) downloadWithoutRanges(d Download) error {
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
    partFileName := fmt.Sprintf("%s.part%d", download.FilePath, partNumber)
    file, err := os.Create(partFileName)
    if (err != nil) {
        fmt.Println(err)
        return err
    }
    defer file.Close()

    // copy the response body to file
    _, err = io.Copy(file, resp.Body)
    if (err != nil) {
        fmt.Println(err)
        return err
    }
    return nil
}
