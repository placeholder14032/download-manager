package download

import (
	"fmt"
	"net/http"
	"strings"
    "io"
    "context"
    "time"
)

func (h *DownloadHandler) IsAcceptRangeSupported() (bool, int64, error) {
    req, err := http.NewRequest("HEAD", h.URL, nil)
    if err != nil {
        return false, 0, fmt.Errorf("failed to create HEAD request: %v", err)
    }
    req.Header.Add("User-Agent", "Go-Download-Client/1.0")

    resp, err := h.Client.Do(req)
    if err != nil {
        return false, 0, fmt.Errorf("HEAD request failed: %v", err)
    }
    defer resp.Body.Close()

    fmt.Println("Response Status Code:", resp.StatusCode)
    fmt.Println("Response Headers:", resp.Header)


    if resp.StatusCode >= 400 {
        return false, 0, fmt.Errorf("server returned status: %d", resp.StatusCode)
    }

    acceptRanges := strings.ToLower(resp.Header.Get("Accept-Ranges"))
    fmt.Println("Accept-Ranges:", acceptRanges)

    if acceptRanges == "bytes" {
        fmt.Println("Range is supported")
        return true, resp.ContentLength, nil
    } else {
        fmt.Println("Range not supported")
    }
    
    return false, resp.ContentLength, nil
}

// Custom reader to ensure we are reading bytes properly
type countingReader struct {
    reader io.Reader
    count  *int64
    handler *DownloadHandler // for more accuracy, updatinf current byte stuff
}

func (r *countingReader) Read(p []byte) (n int, err error) {
    n, err = r.reader.Read(p)
    *r.count += int64(n)
    // Update the DownloadHandler's CurrentByte
    r.handler.State.Mutex.Lock()
    r.handler.State.CurrentByte += int64(n)
    r.handler.State.Mutex.Unlock()

    // Update progress after each read
    r.handler.updateProgress()

    return n, err
}

func (h *DownloadHandler) combineParts( contentLength int64) error {
    c :=  NewPartsCombiner(contentLength,int(h.PartsCount),h.CHUNK_SIZE)
    return c.CombineParts(h.FilePath, contentLength, int(h.PartsCount))
}

// old initializer we might need it later
func (download *Download) NewDlHandler(client *http.Client, chunkSize int64, workersCount int, bandsWidth int64) *DownloadHandler {
	ctx, cancel := context.WithCancel(context.Background())

	// we might need this to avoid NaN we got for speed:
	resp, err := client.Head(download.URL)
	if err != nil {
		fmt.Printf("Failed to get content length: %v\n", err)
	}
	defer resp.Body.Close()
	cl := resp.ContentLength

    dh := &DownloadHandler{
        Client:            client,
        CHUNK_SIZE:        chunkSize,
        WORKERS_COUNT:     workersCount,
        URL:              download.URL,
        FilePath:         download.FilePath,
        State:            &DownloadState{
			TotalBytes: cl,
		}, 

		PauseChan:     make(chan struct{}),
        ResumeChan:    make(chan struct{}),
		ctx:           ctx,
        cancel:        cancel,

		Progress: &ProgressTracker{
            StartTime: time.Now(),
        },
    }
    return dh
}