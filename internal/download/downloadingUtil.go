package download

import (
	"fmt"
	"net/http"
	"strings"
    "io"
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
}

func (r *countingReader) Read(p []byte) (n int, err error) {
    n, err = r.reader.Read(p)
    *r.count += int64(n)
    return n, err
}

func (h *DownloadHandler) handleDownloadCompletion(contentLength int64, errChan <-chan error, done <-chan bool) error {
    
    // fmt.Println("Waiting for done signal...")
    <-done  // block until `done` is closed
    // fmt.Println("Done signal received, proceeding with combineParts")

    for err := range errChan {
        if err != nil {
            fmt.Println("Error in handleDownloadCompletion:", err)
            return err
        }
    }
    if h.State.CurrentByte < h.State.TotalBytes {
        return fmt.Errorf("download incomplete: got %d/%d bytes", h.State.CurrentByte, h.State.TotalBytes)
    }

    fmt.Println("Calling combineParts")
    return h.combineParts(contentLength)
}

func (h *DownloadHandler) combineParts( contentLength int64) error {
    c :=  NewPartsCombiner(contentLength,int(h.PartsCount),h.CHUNK_SIZE)
    return c.CombineParts(h.FilePath, contentLength, int(h.PartsCount))
}