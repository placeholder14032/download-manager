// package main

// import (
//     "fmt"
//     "net/http"
//     "os"
//     "time"
//     "github.com/placeholder14032/download-manager/internal/download"
// )

// func main() {
//     client := &http.Client{
//         Transport: &http.Transport{
//             ResponseHeaderTimeout: 10000 * time.Millisecond, // Slow down for mid-progress pause
//         },
//     }
//     dl := &download.Download{
//         URL:      "https://archive.apache.org/dist/httpd/httpd-2.4.58.tar.gz",
//         FilePath: "/Users/nazaninsmac/Downloads/testfile.tar.gz",
//     }
//     dh := dl.NewDownloadHandler(client, 64*1024, 4, 0)

//     fmt.Printf("Starting download from: %s\n", dl.URL)
//     fmt.Printf("Saving to: %s\n", dl.FilePath)

//     downloadDone := make(chan struct{})
//     go func() {
//         err := dh.StartDownloading()
//         if err != nil {
//             fmt.Printf("Download failed: %v\n", err)
//             return
//         }
//         fmt.Println("Download completed successfully")
//         close(downloadDone)
//     }()

//     ticker := time.NewTicker(1 * time.Second)
//     defer ticker.Stop()
//     go func() {
//         startTime := time.Now()
//         for range ticker.C {
//             elapsed := time.Since(startTime).Round(time.Second)
//             fmt.Printf("Elapsed time: %s, Bytes downloaded: %d/%d\n", elapsed, dh.State.CurrentByte, dh.State.TotalBytes)
//         }
//     }()

//     go func() {
//         time.Sleep(1 * time.Second)
//         fmt.Println("Pause 1 after 1 second...")
//         dh.Pause()

//         time.Sleep(2 * time.Second)
//         fmt.Println("Resume 1 after 2 seconds of pause...")
//         if err := dh.Resume(); err != nil {
//             fmt.Printf("Resume 1 failed: %v\n", err)
//             return
//         }

//         time.Sleep(3 * time.Second)
//         fmt.Println("Pause 2 after 3 seconds...")
//         dh.Pause()

//         time.Sleep(2 * time.Second)
//         fmt.Println("Resume 2 after 2 seconds of pause...")
//         if err := dh.Resume(); err != nil {
//             fmt.Printf("Resume 2 failed: %v\n", err)
//             return
//         }
//     }()

//     select {
//     case <-downloadDone:
//         fmt.Println("Download finished, verifying file...")
//         info, err := os.Stat(dl.FilePath)
//         if err != nil {
//             fmt.Printf("File check failed: %v\n", err)
//         } else {
//             fmt.Printf("File size: %d bytes (expected 9825177 bytes)\n", info.Size())
//         }
//     case <-time.After(60 * time.Second):
//         fmt.Println("Test timeout reached")
//     }
// }

// ---------------------------------------------------------------------------------------------------- SIMPLE W/O PAUSE ...
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