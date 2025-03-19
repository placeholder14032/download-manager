package main

import(
	    "github.com/placeholder14032/download-manager/internal/download"
		"net/http"
		"fmt"
)

func main() {
	 download := download.Download{
		URL:      "http://ovh.net/files/1Mio.dat",
        FilePath: "/Users/nazaninsmac/Downloads/1Mio.dat",
	 }

	 download.Handler = *download.NewDownloadHandler(&http.Client{Timeout: 0}, 1024*1024, 8, 0)
    handler := &download.Handler

	fmt.Printf("Starting download from: %s\n", download.URL)
    fmt.Printf("Saving to: %s\n", download.FilePath)

	downloadErr := make(chan error, 1)


	if err := handler.StartDownloading(); err != nil {
		downloadErr <- fmt.Errorf("initial download failed: %v", err)
		return
	}
}
