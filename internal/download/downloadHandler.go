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

func NewDownloadHandler(client *http.Client, chunkSize int, workersCount int) *DownloadHandler {
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
// start downloading function parts
func (h *DownloadHandler) startWorkers(d *Download, wg *sync.WaitGroup, jobs <-chan chunk, errChan chan<- error, pauseAck chan<- bool) {
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, d, jobs, errChan, pauseAck, wg)
    }
}

func (h *DownloadHandler) distributeJobs(jobs chan<- chunk, contentLength int) {
    defer close(jobs)
    currentByte := h.State.CurrentByte

    for currentByte < int64(contentLength) {
        // Check pause state first
        h.State.Mutex.Lock()
        if h.State.IsPaused {
            h.State.CurrentByte = currentByte
            h.State.Mutex.Unlock()
            return
        }
        h.State.Mutex.Unlock()

        end := currentByte + int64(h.CHUNK_SIZE)
        if end > int64(contentLength) {
            end = int64(contentLength)
        }

        chunk := chunk{Start: int(currentByte), End: int(end - 1)}
        
        select {
        case <-h.PauseChan:
            h.State.Mutex.Lock()
            h.State.CurrentByte = currentByte
            h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
            h.State.Mutex.Unlock()
            fmt.Printf("Distributor paused at byte %d\n", currentByte)
            return
        case jobs <- chunk:
            currentByte = end
            fmt.Printf("Dispatched chunk %d-%d\n", chunk.Start, chunk.End)
        }
    }
}

func (h *DownloadHandler) waitForCompletion(wg *sync.WaitGroup, errChan chan<- error, done chan<- bool) {
    wg.Wait()
    if !h.State.IsPaused {
        close(errChan)
        done <- true
    }
}

func (h *DownloadHandler) handleDownloadCompletion(d *Download, contentLength int, errChan <-chan error, done <-chan bool) error {
    select {
    case err := <-errChan:
        if err != nil {
            return err
        }
        if !h.State.IsPaused {
            return h.combineParts(d, contentLength)
        }
        return nil
    case <-done:
        return h.combineParts(d, contentLength)
    }
}
// worker
func (h *DownloadHandler) worker(id int, d *Download, jobs <-chan chunk, errChan chan<- error, pauseAck chan<- bool, wg *sync.WaitGroup) {
    defer wg.Done()

    for chunk := range jobs {
        // Check pause state before starting new chunk
        h.State.Mutex.Lock()
        if h.State.IsPaused {
            h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
            h.State.Mutex.Unlock()
            fmt.Printf("Worker %d stopped due to pause state\n", id)
            pauseAck <- true
            return
        }
        h.State.Mutex.Unlock()

        select {
        case <-h.PauseChan:
            h.State.Mutex.Lock()
            h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
            h.State.Mutex.Unlock()
            fmt.Printf("Worker %d paused at chunk %d-%d\n", id, chunk.Start, chunk.End)
            pauseAck <- true
            return
        default:
            if err := h.downloadWithRanges(d, chunk.Start, chunk.End); err != nil {
                errChan <- fmt.Errorf("worker %d failed: %v", id, err)
                return
            }
            h.State.Mutex.Lock()
            h.State.Completed[chunk.Start/h.CHUNK_SIZE] = true
            h.State.Mutex.Unlock()
        }
    }
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

func (h *DownloadHandler) combineParts(download *Download, contentLength int) error {
    c :=  NewPartsCombiner()
    return c.CombineParts(download.FilePath, contentLength, h.PartsCount)
}

func (h *DownloadHandler) IsAcceptRangeSupported(download Download) (bool, int, error) {
    req, err := http.NewRequest("HEAD", download.URL, nil)
    if err != nil {
        return false, 0, fmt.Errorf("failed to create HEAD request: %v", err)
    }

    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return false, 0, fmt.Errorf("HEAD request failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return false, 0, fmt.Errorf("server returned status: %d", resp.StatusCode)
    }

    acceptRanges := strings.ToLower(resp.Header.Get("Accept-Ranges"))
    if acceptRanges == "" || acceptRanges == "none" {
        return false, int(resp.ContentLength), nil
    }
    return true, int(resp.ContentLength), nil
}


func (h *DownloadHandler) Pause() error {
    if h.State == nil {
        return fmt.Errorf("download not initialized")
    }

    h.State.Mutex.Lock()
    defer h.State.Mutex.Unlock()

    if h.State.IsPaused {
        return fmt.Errorf("download is already paused")
    }

    h.State.IsPaused = true
    
    if h.PauseChan != nil {
        close(h.PauseChan)
        fmt.Println("Download pause signal sent")
    }

    return nil
}

func (h *DownloadHandler) Resume(d Download) error {
    if h.State == nil {
        return fmt.Errorf("download not initialized")
    }

    h.State.Mutex.Lock()
    if !h.State.IsPaused {
        h.State.Mutex.Unlock()
        return fmt.Errorf("download is not paused")
    }

    h.State.IsPaused = false

    // Create new pause channel
    h.PauseChan = make(chan struct{})

    // Get incomplete parts and clear them from state
    incompleteParts := make([]chunk, len(h.State.IncompleteParts))
    copy(incompleteParts, h.State.IncompleteParts)
    h.State.IncompleteParts = make([]chunk, 0)
    h.State.Mutex.Unlock()

    return h.resumingDownload(d, incompleteParts)
}

func (h *DownloadHandler) resumingDownload(d Download, incompleteParts []chunk) error {
    //  check if it's already completed
    completedCount := 0
    h.State.Mutex.Lock()
    for _, completed := range h.State.Completed {
        if completed {
            completedCount++
        }
    }
    h.State.Mutex.Unlock()

    // if already completed -> so we should just combine parts
    if completedCount == h.PartsCount {
        return h.combineParts(&d, int(h.State.TotalBytes))
    }

    // else we should resume now we will do the same as in StartDownloading we are defining and alligning jobs
    // then we will start workers and distribute jobs
    jobs := make(chan chunk, h.WORKERS_COUNT)
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool)
    pauseAck := make(chan bool, h.WORKERS_COUNT)
    var wg sync.WaitGroup

    // start workers
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, &d, jobs, errChan, pauseAck, &wg)
    }

    // job distribution
    go func() {
        defer close(jobs)
        
        //  handle incomplete parts first
        for _, part := range incompleteParts {
            
            h.State.Mutex.Lock()
            isCompleted := h.State.Completed[part.Start/h.CHUNK_SIZE]
            h.State.Mutex.Unlock()
            
            if !isCompleted {
                select {
                case <-h.PauseChan:
                    h.State.Mutex.Lock()
                    h.State.IncompleteParts = append(h.State.IncompleteParts, part)
                    h.State.Mutex.Unlock()
                    return
                case jobs <- part:
                    fmt.Printf("Resuming part %d (bytes %d-%d)\n", 
                        part.Start/h.CHUNK_SIZE, part.Start, part.End)
                }
            }
        }

        // then we will handle the remaining parts

        h.State.Mutex.Lock()
        currentByte := h.State.CurrentByte
        h.State.Mutex.Unlock()

        for currentByte < h.State.TotalBytes {
            partIndex := int(currentByte) / h.CHUNK_SIZE
            
            h.State.Mutex.Lock()
            isCompleted := h.State.Completed[partIndex]
            h.State.Mutex.Unlock()

            if !isCompleted {
                end := currentByte + int64(h.CHUNK_SIZE)
                if end > h.State.TotalBytes {
                    end = h.State.TotalBytes
                }

                chunk := chunk{Start: int(currentByte), End: int(end - 1)}
                select {
                case <-h.PauseChan:
                    h.State.Mutex.Lock()
                    h.State.CurrentByte = currentByte
                    h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
                    h.State.Mutex.Unlock()
                    return
                case jobs <- chunk:
                    fmt.Printf("Downloading part %d (bytes %d-%d)\n", 
                        partIndex, currentByte, end-1)
                }
            }
            currentByte += int64(h.CHUNK_SIZE)
        }
    }()

    // Wait and handle completion
    go func() {
        wg.Wait()
        if !h.State.IsPaused {
            close(errChan)
            done <- true
        }
    }()

    select {
    case err := <-errChan:
        if err != nil {
            return err
        }
        if !h.State.IsPaused {
            return h.combineParts(&d, int(h.State.TotalBytes))
        }
    case <-done:
        return h.combineParts(&d, int(h.State.TotalBytes))
    }

    return nil
}

