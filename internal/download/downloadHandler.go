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
    Paused        bool
    PartsCount    int
    State         *DownloadState
    PauseChan     chan struct{} 
}

type chunk struct {
    Start int
    End   int
}

// to do:
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
    if !supportsRange {
        return h.downloadWithoutRanges(d, contentLength)
    }

    // we have 3 channels one for errors one for jobs (used to send download chunks to worker goroutines) and one for done and for error handling
    jobs := make(chan chunk, h.WORKERS_COUNT)    // sends chunk information to workers
    errChan := make(chan error, h.WORKERS_COUNT)
    done := make(chan bool)
    pauseAck := make(chan bool, h.WORKERS_COUNT) // Added to acknowledge worker pause completion

    h.PartsCount = (contentLength + h.CHUNK_SIZE - 1) / h.CHUNK_SIZE
    fmt.Printf("Starting download with %d parts\n", h.PartsCount)

    // initializing state
    if h.State == nil {
        h.State = &DownloadState{
            Completed:       make([]bool, h.PartsCount),
            IncompleteParts: make([]chunk, 0),
            CurrentByte:     0,
            TotalBytes:      int64(contentLength),
            mutex:           sync.Mutex{},
            isPaused:        false,
        }
    } else {
        h.State.mutex.Lock()
        h.State.Completed = make([]bool, h.PartsCount)
        h.State.IncompleteParts = make([]chunk, 0)
        h.State.CurrentByte = 0
        h.State.TotalBytes = int64(contentLength)
        h.State.isPaused = false
        h.State.mutex.Unlock()
    }

    if h.PauseChan == nil {
        h.PauseChan = make(chan struct{})
    }

    var wg sync.WaitGroup
    for i := 0; i < h.WORKERS_COUNT; i++ {
        wg.Add(1)
        go h.worker(i, &d, jobs, errChan, pauseAck, &wg)
    }

    // managing what work to give to workers
    go func() {
        currentByte := h.State.CurrentByte
        for currentByte < int64(contentLength) {
            h.State.mutex.Lock()
            if h.State.isPaused {
                h.State.CurrentByte = currentByte
                h.State.mutex.Unlock()
                return
            }
            h.State.mutex.Unlock()

            end := currentByte + int64(h.CHUNK_SIZE)
            if end > int64(contentLength) {
                end = int64(contentLength)
            }

            select {
            case <-h.PauseChan:
                // we should pause the download so first we need to save the state -> mutex for avoiding race condition
                h.State.mutex.Lock()
                h.State.CurrentByte = currentByte
                h.State.mutex.Unlock()
                return
            case jobs <- chunk{Start: int(currentByte), End: int(end - 1)}:
                currentByte = end
            }
        }
        fmt.Printf("All %d parts scheduled for download\n", h.PartsCount)
        close(jobs)
    }()

    // wait for workers and handle errors
    go func() {
        wg.Wait()
        if !h.State.isPaused {
            close(errChan)
            done <- true
        }
    }()

    // handle errors
    select {
    case err := <-errChan:
        if err != nil {
            return err
        }
        if !h.State.isPaused {
            fmt.Println("Download completed, starting to combine parts...")
            if err := h.combineParts(&d, contentLength); err != nil {
                return fmt.Errorf("failed to combine parts: %v", err)
            }
            fmt.Println("Parts combined successfully")
            return nil
        }
        return nil
    case <-done:
        // combine parts if the download was successful
        fmt.Println("Download completed, starting to combine parts...")
        if err := h.combineParts(&d, contentLength); err != nil {
            return fmt.Errorf("failed to combine parts: %v", err)
        }
        fmt.Println("Parts combined successfully")
        return nil
    }
    return nil
}

func (h *DownloadHandler) worker(id int, d *Download, jobs <-chan chunk, errChan chan<- error, pauseAck chan<- bool, wg *sync.WaitGroup) {
    defer wg.Done()

    for chunk := range jobs {
        select {
        case <-h.PauseChan:
            // we should append the incomplete part to the state, we need mutex to avoid race condition
            h.State.mutex.Lock()
            h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
            h.State.mutex.Unlock()
            fmt.Printf("Worker %d paused, saving incomplete chunk %d-%d\n", id, chunk.Start, chunk.End)
            pauseAck <- true
            return
        default:
            if err := h.downloadWithRanges(d, chunk.Start, chunk.End); err != nil {
                errChan <- fmt.Errorf("worker %d failed: %v", id, err)
                return
            }
            h.State.mutex.Lock()
            h.State.Completed[chunk.Start/h.CHUNK_SIZE] = true
            h.State.mutex.Unlock()
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
    fmt.Printf("Starting to combine %d parts\n", h.PartsCount)
    // the final file
    combinedFile, err := os.Create(download.FilePath)
    if err != nil {
        return fmt.Errorf("failed to create the final file: %v", err)
    }
    defer combinedFile.Close()

    combinedParts := make([]bool, h.PartsCount)

    // combining parts in order
    buffer := make([]byte, 32*1024)
    for i := 0; i < h.PartsCount; i++ {
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

        // delete part file after successful copy
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

type DownloadState struct {
    IncompleteParts []chunk
    Completed       []bool
    CurrentByte     int64
    TotalBytes      int64
    mutex           sync.Mutex
    isPaused        bool
}

func (h *DownloadHandler) Pause() error {
    if h.State == nil {
        return fmt.Errorf("download not initialized")
    }

    h.State.mutex.Lock()
    defer h.State.mutex.Unlock()

    if h.Paused {
        return fmt.Errorf("download is already paused")
    }

    h.Paused = true
    h.State.isPaused = true
    
    // closing a channel signals all goroutines, here workers that are listening on it that they should stop
    if h.PauseChan != nil {
        close(h.PauseChan)
        fmt.Println("Download pause signal sent")
    }

    return nil
}

func (h *DownloadHandler) Resume() {}

func NewDownloadHandler(client *http.Client, chunkSize int, workersCount int) *DownloadHandler {
    return &DownloadHandler{
        Client:        client,
        CHUNK_SIZE:    chunkSize,
        WORKERS_COUNT: workersCount,
        Paused:        false,
        PauseChan:     make(chan struct{}),
        State:         &DownloadState{
            Completed:       make([]bool, 0),
            IncompleteParts: make([]chunk, 0),
            CurrentByte:     0,
            TotalBytes:      0,
            mutex:           sync.Mutex{},
            isPaused:        false,
        },
    }
}