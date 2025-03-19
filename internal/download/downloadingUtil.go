package download

import (
	"fmt"
	"net/http"
	"strings"
)

func (h *DownloadHandler) IsAcceptRangeSupported() (bool, int64, error) {
    fmt.Println("Before HTTP request")
    req, err := http.NewRequest("HEAD", h.URL, nil)
    if err != nil {
        return false, 0, fmt.Errorf("failed to create HEAD request: %v", err)
    }
    fmt.Println("After HTTP request creation")

    resp, err := h.Client.Do(req)
    if err != nil {
        return false, 0, fmt.Errorf("HEAD request failed: %v", err)
    }
    fmt.Println("After executing the HEAD request")
    defer resp.Body.Close()

    fmt.Println("Response Headers:", resp.Header)


    if resp.StatusCode >= 400 {
        return false, 0, fmt.Errorf("server returned status: %d", resp.StatusCode)
    }

    acceptRanges := strings.ToLower(resp.Header.Get("Accept-Ranges"))
    if acceptRanges == "bytes" {
        return true, resp.ContentLength, nil
    }
    return false, resp.ContentLength, nil
}
