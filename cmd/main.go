package main

import (
    "github.com/placeholder14032/download-manager/internal/download"
    "net/http"
    "fmt"
    "time"
)

func main() {
    bandwidthLimit := int64(500 * 1024)

    download := download.Download{
        URL:      "https://sample-videos.com/video321/mp4/720/big_buck_bunny_720p_1mb.mp4",
        FilePath: "/Users/nazaninsmac/Downloads/big_buck_bunny_1mb.mp4",
    }

    // Use an HTTP client with a timeout to avoid long network delays
    client := &http.Client{
        Timeout: 10 * time.Second,
        Transport: &http.Transport{
            ResponseHeaderTimeout: 5 * time.Second, // Timeout for reading response headers
        },
    }
    download.Handler = *download.NewDownloadHandler(client, bandwidthLimit)
    handler := &download.Handler

    // Periodically call DisplayProgress to check the speed
    go func() {
        for {
            handler.DisplayProgress()
            time.Sleep(1 * time.Second)
        }
    }()

    fmt.Printf("Starting download from: %s\n", download.URL)
    fmt.Printf("Saving to: %s\n", download.FilePath)

    downloadErr := make(chan error, 1)

    if err := handler.StartDownloading(); err != nil {
        downloadErr <- fmt.Errorf("initial download failed: %v", err)
        return
    }
}
// ------------------------------------------------------------------------------------------------------- Serializer
// package main

// import (
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"time"

// 	"github.com/placeholder14032/download-manager/internal/download"
// )
// const stateFile = "download_state.json" // File to store the serialized state

// func main() {
// 	// Configure HTTP client with a longer timeout
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			ResponseHeaderTimeout: 30 * time.Second,
// 		},
// 	}

// 	// Define the download
// 	dl := download.Download{
// 		URL:      "https://archive.apache.org/dist/httpd/httpd-2.4.58.tar.gz",
// 		FilePath: "/Users/nazaninsmac/Downloads/testfile.tar.gz",
// 	}

// 	// Check if a saved state exists
// 	var dh *download.DownloadHandler
// 	if _, err := os.Stat(stateFile); err == nil {
// 		// Load existing state
// 		data, err := os.ReadFile(stateFile)
// 		if err != nil {
// 			fmt.Printf("Failed to read state file: %v\n", err)
// 			return
// 		}
// 		dh, err = download.DeserializeHandler(data, client)
// 		if err != nil {
// 			fmt.Printf("Failed to deserialize state: %v\n", err)
// 			return
// 		}
// 		fmt.Printf("Resuming download from saved state: %s\n", dh.URL)
// 	} else {
// 		// Start a new download
// 		dh = dl.NewDownloadHandler(client, 512*1024, 4, 0)
// 		if dh == nil {
// 			fmt.Println("Failed to initialize DownloadHandler")
// 			return
// 		}
// 		fmt.Printf("Starting new download from: %s\n", dl.URL)
// 	}
// 	fmt.Printf("Saving to: %s\n", dh.FilePath)

// 	// Channels for coordination
// 	downloadDone := make(chan struct{})
// 	stopAll := make(chan struct{})

// 	// Download goroutine
// 	go func() {
// 		err := dh.StartDownloading()
// 		if err != nil {
// 			fmt.Printf("Download failed: %v\n", err)
// 		} else {
// 			fmt.Println("Download completed successfully")
// 			// Remove state file on successful completion
// 			if err := os.Remove(stateFile); err != nil {
// 				fmt.Printf("Warning: failed to remove state file: %v\n", err)
// 			}
// 		}
// 		close(downloadDone)
// 	}()

// 	// Progress reporting goroutine
// 	go func() {
// 		ticker := time.NewTicker(1 * time.Second)
// 		defer ticker.Stop()
// 		fmt.Println("Progress reporting started")
// 		for {
// 			select {
// 			case <-stopAll:
// 				fmt.Println("Progress reporting stopped")
// 				return
// 			case <-ticker.C:
// 				have := dh.State.CurrentByte
// 				want := dh.State.TotalBytes
// 				if want > 0 {
// 					percent := float64(have) / float64(want) * 100
// 					elapsed := time.Since(dh.Progress.StartTime).Seconds()
// 					speed := float64(have) / elapsed / 1024 // KB/s
// 					fmt.Printf("Progress: %.2f%% (%d/%d bytes), Speed: %.2f KB/s\n", percent, have, want, speed)
// 				}
// 			}
// 		}
// 	}()

// 	// Pause/resume test goroutine with serialization to file
// 	go func() {
// 		for i := 1; i <= 2; i++ {
// 			select {
// 			case <-stopAll:
// 				return
// 			case <-time.After(5 * time.Second):
// 				fmt.Printf("Pausing download #%d after 5 seconds...\n", i)
// 				dh.Pause()

// 				// Serialize state and write to file
// 				data, err := dh.Serialize()
// 				if err != nil {
// 					fmt.Printf("Serialization failed: %v\n", err)
// 					return
// 				}
// 				if err := os.WriteFile(stateFile, data, 0644); err != nil {
// 					fmt.Printf("Failed to write state to file: %v\n", err)
// 					return
// 				}
// 				fmt.Printf("Download state saved to %s\n", stateFile)

// 				select {
// 				case <-stopAll:
// 					return
// 				case <-time.After(3 * time.Second):
// 					// Load from file and deserialize
// 					data, err = os.ReadFile(stateFile)
// 					if err != nil {
// 						fmt.Printf("Failed to read state file: %v\n", err)
// 						return
// 					}
// 					newDh, err := download.DeserializeHandler(data, client)
// 					if err != nil {
// 						fmt.Printf("Deserialization failed: %v\n", err)
// 						return
// 					}
// 					dh = newDh // Replace the old handler
// 					fmt.Printf("Resuming download #%d after 3 seconds of pause...\n", i)
// 					if err := dh.Resume(); err != nil {
// 						fmt.Printf("Resume #%d failed: %v\n", i, err)
// 						return
// 					}
// 				}
// 			}
// 		}
// 	}()

// 	// Wait for download completion or timeout
// 	select {
// 	case <-downloadDone:
// 		close(stopAll)
// 		fmt.Println("Download finished, verifying file...")
// 		info, err := os.Stat(dh.FilePath)
// 		if err != nil {
// 			fmt.Printf("File check failed: %v\n", err)
// 		} else {
// 			fmt.Printf("File size: %d bytes (expected 9825177 bytes)\n", info.Size())
// 		}
// 	case <-time.After(60 * time.Second):
// 		fmt.Println("Test timeout reached")
// 		dh.Pause()
// 		// Serialize final state to file
// 		data, err := dh.Serialize()
// 		if err != nil {
// 			fmt.Printf("Serialization failed: %v\n", err)
// 		} else if err := os.WriteFile(stateFile, data, 0644); err != nil {
// 			fmt.Printf("Failed to write final state to file: %v\n", err)
// 		} else {
// 			fmt.Printf("Final state saved to %s; you can resume later\n", stateFile)
// 		}
// 		close(stopAll)
// 	}

// 	fmt.Println("Exiting program")
// }
// --------------------------------------------------------------------------------------------------------- progressTracker
// package main

// import (
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"time"

// 	"github.com/placeholder14032/download-manager/internal/download" // Adjust this import path
// )

// func main() {
// 	// Configure HTTP client with a longer timeout
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			ResponseHeaderTimeout: 30 * time.Second, // Increase to avoid timeouts
// 		},
// 	}

// 	// Define the download
// 	dl := &download.Download{
// 		URL:      "https://archive.apache.org/dist/httpd/httpd-2.4.58.tar.gz",
// 		FilePath: "/Users/nazaninsmac/Downloads/testfile.tar.gz",
// 	}

// 	// Initialize DownloadHandler
// 	dh := dl.NewDownloadHandler(client, 512*1024, 4, 0)
// 	fmt.Printf("Starting download from: %s\n", dl.URL)
// 	fmt.Printf("Saving to: %s\n", dl.FilePath)

// 	// Channels for coordination
// 	downloadDone := make(chan struct{})
// 	stopAll := make(chan struct{})

// 	// Download goroutine
// 	go func() {
// 		err := dh.StartDownloading()
// 		if err != nil {
// 			fmt.Printf("Download failed: %v\n", err)
// 		} else {
// 			fmt.Println("Download completed successfully")
// 		}
// 		close(downloadDone)
// 	}()

// 	// Progress reporting goroutine
// 	go func() {
// 		ticker := time.NewTicker(1 * time.Second)
// 		defer ticker.Stop()
// 		fmt.Println("Progress reporting started") // Debug
// 		for {
// 			select {
// 			case <-stopAll:
// 				fmt.Println("Progress reporting stopped") // Debug
// 				return
// 			case <-ticker.C:
// 				dh.UpdateProgress()
// 				dh.DisplayProgress()
// 			}
// 		}
// 	}()

// 	// Pause/resume test goroutine
// 	go func() {
// 		for i := 1; i <= 2; i++ {
// 			select {
// 			case <-stopAll:
// 				return
// 			case <-time.After(5 * time.Second):
// 				fmt.Printf("Pausing download #%d after 2 seconds...\n", i)
// 				dh.Pause()

// 				select {
// 				case <-stopAll:
// 					return
// 				case <-time.After(3 * time.Second):
// 					fmt.Printf("Resuming download #%d after 3 seconds of pause...\n", i)
// 					if err := dh.Resume(); err != nil {
// 						fmt.Printf("Resume #%d failed: %v\n", i, err)
// 						return
// 					}
// 				}
// 			}
// 		}
// 	}()

// 	// Wait for download completion or timeout
// 	select {
// 	case <-downloadDone:
// 		close(stopAll)
// 		fmt.Println("Download finished, verifying file...")
// 		info, err := os.Stat(dl.FilePath)
// 		if err != nil {
// 			fmt.Printf("File check failed: %v\n", err)
// 		} else {
// 			fmt.Printf("File size: %d bytes (expected ~9825177 bytes)\n", info.Size())
// 		}
// 	case <-time.After(60 * time.Second):
// 		fmt.Println("Test timeout reached")
// 		dh.Pause()
// 		close(stopAll)
// 	}

// 	fmt.Println("Exiting program")
// }
// --------------------------------------------------------------------------------------------------------------- pause/resume test
// package main

// import (
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"time"
// 	"github.com/placeholder14032/download-manager/internal/download"
// )

// func main() {
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			ResponseHeaderTimeout: 10000 * time.Millisecond,
// 		},
// 	}

// 	dl := &download.Download{
// 		URL:      "https://archive.apache.org/dist/httpd/httpd-2.4.58.tar.gz",
// 		FilePath: "/Users/nazaninsmac/Downloads/testfile.tar.gz",
// 	}

// 	dh := dl.NewDownloadHandler(client, 512*1024, 4, 0)
// 	fmt.Printf("Starting download from: %s\n", dl.URL)
// 	fmt.Printf("Saving to: %s\n", dl.FilePath)

// 	downloadDone := make(chan struct{})
// 	stopAll := make(chan struct{})

// 	// Download goroutine
// 	go func() {
// 		err := dh.StartDownloading()
// 		if err != nil {
// 			fmt.Printf("Download failed: %v\n", err)
// 		} else {
// 			fmt.Println("Download completed successfully")
// 		}
// 		close(downloadDone)
// 	}()

// 	// Status reporting goroutine
// 	go func() {
// 		startTime := time.Now()
// 		ticker := time.NewTicker(1 * time.Second)
// 		defer ticker.Stop()

// 		for {
// 			select {
// 			case <-ticker.C:
// 				elapsed := time.Since(startTime).Round(time.Second)
// 				fmt.Printf("Elapsed time: %s, Bytes downloaded: %d/%d\n",
// 					elapsed, dh.State.CurrentByte, dh.State.TotalBytes)
// 			case <-stopAll:
// 				ticker.Stop() // Explicitly stop the ticker
// 				return        // Exit the goroutine immediately
// 			}
// 		}
// 	}()

// 	// Pause/resume test goroutine
// 	go func() {
// 		select {
// 		case <-stopAll:
// 			return
// 		default:
// 			time.Sleep(1 * time.Second)
// 			fmt.Println("Pause 1 after 1 second...")
// 			dh.Pause()
// 			time.Sleep(2 * time.Second)
// 			fmt.Println("Resume 1 after 2 seconds of pause...")
// 			if err := dh.Resume(); err != nil {
// 				fmt.Printf("Resume 1 failed: %v\n", err)
// 				return
// 			}

// 			time.Sleep(3 * time.Second)
// 			fmt.Println("Pause 2 after 3 seconds...")
// 			dh.Pause()
// 			time.Sleep(2 * time.Second)
// 			fmt.Println("Resume 2 after 2 seconds of pause...")
// 			if err := dh.Resume(); err != nil {
// 				fmt.Printf("Resume 2 failed: %v\n", err)
// 				return
// 			}
// 		}
// 	}()

// 	// Wait for download completion or timeout
// 	select {
// 	case <-downloadDone:
// 		close(stopAll)
// 		fmt.Println("Download finished, verifying file...")
// 		info, err := os.Stat(dl.FilePath)
// 		if err != nil {
// 			fmt.Printf("File check failed: %v\n", err)
// 		} else {
// 			fmt.Printf("File size: %d bytes (expected 9825177 bytes)\n", info.Size())
// 		}
// 	case <-time.After(60 * time.Second):
// 		fmt.Println("Test timeout reached")
// 		close(stopAll)
// 	}

// 	// No need for extra sleep if goroutines exit cleanly
// 	fmt.Println("Exiting program")
// }
// ---------------------------------------------------------------------------------------------------- SIMPLE W/O PAUSE ...
// package main

// import(
// 	    "github.com/placeholder14032/download-manager/internal/download"
// 		"net/http"
// 		"fmt"
// 		"time"
// )

// func main() {
// 	 download := download.Download{
// 		URL:           "https://sample-videos.com/video321/mp4/720/big_buck_bunny_720p_1mb.mp4",
// 		FilePath:      "/Users/nazaninsmac/Downloads/big_buck_bunny_1mb.mp4",
// 	 }

// 	// download.Handler = *download.NewDownloadHandler(&http.Client{Timeout: 10 * time.Second}, 256 * 1024, 5, 0)
// 	download.Handler = *download.NewDownloadHandler(&http.Client{Timeout: 10 * time.Second})
//     handler := &download.Handler



// 	fmt.Printf("Starting download from: %s\n", download.URL)
//     fmt.Printf("Saving to: %s\n", download.FilePath)

// 	downloadErr := make(chan error, 1)


// 	if err := handler.StartDownloading(); err != nil {
// 		downloadErr <- fmt.Errorf("initial download failed: %v", err)
// 		return
// 	}
// }