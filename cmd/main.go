package main

import(
	    "github.com/placeholder14032/download-manager/internal/download"
		"net/http"
		"fmt"
		"time"
)

func main() {
	 download := download.Download{
		URL:     "https://releases.ubuntu.com/24.04.1/SHA256SUMS",
        FilePath: "/Users/nazaninsmac/Downloads/SHA256SUMS",
	 }

	 download.Handler = *download.NewDownloadHandler(&http.Client{Timeout: 10 * time.Second}, 1024*1024, 3, 0)
    handler := &download.Handler

	fmt.Printf("Starting download from: %s\n", download.URL)
    fmt.Printf("Saving to: %s\n", download.FilePath)

	downloadErr := make(chan error, 1)


	if err := handler.StartDownloading(); err != nil {
		downloadErr <- fmt.Errorf("initial download failed: %v", err)
		return
	}
}
