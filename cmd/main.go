package main

import(
	    "github.com/placeholder14032/download-manager/internal/download"
		"net/http"
		"fmt"
		"time"
)

func main() {
	 download := download.Download{
		URL:           "https://sample-videos.com/video321/mp4/720/big_buck_bunny_720p_1mb.mp4",
		FilePath:      "/Users/nazaninsmac/Downloads/big_buck_bunny_1mb.mp4",
	 }

	 download.Handler = *download.NewDownloadHandler(&http.Client{Timeout: 10 * time.Second}, 256 * 1024, 5, 0)
    handler := &download.Handler

	fmt.Printf("Starting download from: %s\n", download.URL)
    fmt.Printf("Saving to: %s\n", download.FilePath)

	downloadErr := make(chan error, 1)


	if err := handler.StartDownloading(); err != nil {
		downloadErr <- fmt.Errorf("initial download failed: %v", err)
		return
	}
}
