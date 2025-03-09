package download

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type DownloadHandler struct {
	client        *http.Client
	CHUNK_SIZE    int
	WORKERS_COUNT int
	paused        bool
	lastByte      int // name?
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

func (h *DownloadHandler) startDownloading(d Download) {
	supportsRange, contentLength := h.isAcceptRangeSupported(d)

	if supportsRange {
		// spliting to chunks
		for !h.paused {
			for i := 0; i < h.WORKERS_COUNT; i++ {
				h.downloadWithRanges(&d, h.lastByte, h.lastByte+h.CHUNK_SIZE)
				h.lastByte += h.CHUNK_SIZE
			}
		}

	}
	h.downloadWithoutRanges(d, contentLength)
}

func (h *DownloadHandler) downloadWithoutRanges(d Download, contentLength int) {
	panic("unimplemented")
}

func (h *DownloadHandler) downloadWithRanges(download *Download, start int, end int) {
}
