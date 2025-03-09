package download

import (
    "net/http"
    "time"
    "fmt"
    "strings"
)

type DownloadHandler struct {
    client       *http.Client
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