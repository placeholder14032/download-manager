package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

type chunk struct {
    Start int
    End   int
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

func (h *DownloadHandler) StartDownloading(d Download) error {
    supportsRange, contentLength := h.isAcceptRangeSupported(d)
    if (!supportsRange) {
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
            IsPaused:        false,
            isCombined:      false,
        }
    } else {
        h.State.mutex.Lock()

        h.State.Completed = make([]bool, h.PartsCount)
        h.State.IncompleteParts = make([]chunk, 0)
        h.State.CurrentByte = 0
        h.State.TotalBytes = int64(contentLength)
        h.State.IsPaused = false
        h.State.isCombined = false
        
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
            if h.State.IsPaused {
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
        if !h.State.IsPaused {
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
        if (!h.State.IsPaused) {
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
    if info, err := os.Stat(download.FilePath); err == nil {
        if info.Size() == int64(contentLength) {
            return nil 
        }
    }

    partFiles, err := filepath.Glob(fmt.Sprintf("%s.part*", download.FilePath))
    if err != nil {
        return fmt.Errorf("failed to check part files: %v", err)
    }
    
    if len(partFiles) == 0 {
        if info, err := os.Stat(download.FilePath); err == nil && info.Size() == int64(contentLength) {
            return nil
        }
        return fmt.Errorf("no part files found to combine")
    }

    partsMap := make(map[int]string)
    for _, partFile := range partFiles {
        var partNum int
        _, err := fmt.Sscanf(filepath.Base(partFile), filepath.Base(download.FilePath)+".part%d", &partNum)
        if err != nil {
            continue
        }
        partsMap[partNum] = partFile
    }

    for i := 0; i < h.PartsCount; i++ {
        if _, exists := partsMap[i]; !exists {
            return fmt.Errorf("missing part file %d", i)
        }
    }

    fmt.Printf("Starting to combine %d parts\n", h.PartsCount)
    
    combinedFile, err := os.Create(download.FilePath)
    if err != nil {
        return fmt.Errorf("failed to create final file: %v", err)
    }
    defer combinedFile.Close()

    buffer := make([]byte, 32*1024)
    completedParts := make([]string, 0, len(partFiles))
    
    for i := 0; i < h.PartsCount; i++ {
        partFile, err := os.Open(partsMap[i])
        if err != nil {
            return fmt.Errorf("failed to open part %d: %v", i, err)
        }

        _, err = io.CopyBuffer(combinedFile, partFile, buffer)
        partFile.Close()

        if err != nil {
            return fmt.Errorf("failed to copy part %d: %v", i, err)
        }
        completedParts = append(completedParts, partsMap[i])
    }

    info, err := os.Stat(download.FilePath)
    if err != nil {
        return fmt.Errorf("failed to verify final file: %v", err)
    }
    if info.Size() != int64(contentLength) {
        return fmt.Errorf("final file size mismatch: got %d, want %d", info.Size(), contentLength)
    }

    // Single cleanup of ALL part files after successful combination
    cleanupFiles, _ := filepath.Glob(fmt.Sprintf("%s.part*", download.FilePath))
    for _, partFile := range cleanupFiles {
        if err := os.Remove(partFile); err != nil {
            fmt.Printf("Warning: failed to remove part file %s: %v\n", partFile, err)
        }
    }

    fmt.Println("All parts combined successfully")
    return nil
}

type DownloadState struct {
    IncompleteParts []chunk
    Completed       []bool
    CurrentByte     int64
    TotalBytes      int64
    mutex           sync.Mutex
    IsPaused        bool
    isCombined      bool   
}

func (h *DownloadHandler) Pause() error {
    if h.State == nil {
        return fmt.Errorf("download not initialized")
    }

    h.State.mutex.Lock()
    defer h.State.mutex.Unlock()

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

    h.State.mutex.Lock()
    if !h.State.IsPaused {
        h.State.mutex.Unlock()
        return fmt.Errorf("download is not paused")
    }

    h.State.IsPaused = false

    // Create new pause channel
    h.PauseChan = make(chan struct{})

    // Get incomplete parts and clear them from state
    incompleteParts := make([]chunk, len(h.State.IncompleteParts))
    copy(incompleteParts, h.State.IncompleteParts)
    h.State.IncompleteParts = make([]chunk, 0)
    h.State.mutex.Unlock()

    return h.resumingDownload(d, incompleteParts)
}

func (h *DownloadHandler) resumingDownload(d Download, incompleteParts []chunk) error {
    //  check if it's already completed
    completedCount := 0
    h.State.mutex.Lock()
    for _, completed := range h.State.Completed {
        if completed {
            completedCount++
        }
    }
    h.State.mutex.Unlock()

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
            
            h.State.mutex.Lock()
            isCompleted := h.State.Completed[part.Start/h.CHUNK_SIZE]
            h.State.mutex.Unlock()
            
            if !isCompleted {
                select {
                case <-h.PauseChan:
                    h.State.mutex.Lock()
                    h.State.IncompleteParts = append(h.State.IncompleteParts, part)
                    h.State.mutex.Unlock()
                    return
                case jobs <- part:
                    fmt.Printf("Resuming part %d (bytes %d-%d)\n", 
                        part.Start/h.CHUNK_SIZE, part.Start, part.End)
                }
            }
        }

        // then we will handle the remaining parts

        h.State.mutex.Lock()
        currentByte := h.State.CurrentByte
        h.State.mutex.Unlock()

        for currentByte < h.State.TotalBytes {
            partIndex := int(currentByte) / h.CHUNK_SIZE
            
            h.State.mutex.Lock()
            isCompleted := h.State.Completed[partIndex]
            h.State.mutex.Unlock()

            if !isCompleted {
                end := currentByte + int64(h.CHUNK_SIZE)
                if end > h.State.TotalBytes {
                    end = h.State.TotalBytes
                }

                chunk := chunk{Start: int(currentByte), End: int(end - 1)}
                select {
                case <-h.PauseChan:
                    h.State.mutex.Lock()
                    h.State.CurrentByte = currentByte
                    h.State.IncompleteParts = append(h.State.IncompleteParts, chunk)
                    h.State.mutex.Unlock()
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
            mutex:           sync.Mutex{},
            isCombined:      false,
        },
    }
}