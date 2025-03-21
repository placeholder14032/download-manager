package download

import(
	"sync"
	"time"
	"fmt"
)

type ProgressTracker struct {
    StartTime      time.Time
    LastUpdateTime time.Time
    LastBytes      int64
    CurrentSpeed   float64 // Speed over the last update interval (bytes/s)
    AvgSpeed       float64 // Overall average speed (bytes/s)
	SpeedSamples   []float64 // we will use this to make progress tracking more smooth (like the realetion we had in physics)
    Mutex          sync.Mutex
}


func (h *DownloadHandler) updateProgress() {
    h.Progress.Mutex.Lock()
    defer h.Progress.Mutex.Unlock()
    h.State.Mutex.Lock()
    defer h.State.Mutex.Unlock()

    now := time.Now()
    totalElapsed := now.Sub(h.Progress.StartTime).Seconds()
    intervalElapsed := now.Sub(h.Progress.LastUpdateTime).Seconds()

    // Current speed (over the last interval)
    if intervalElapsed > 0 {
        bytesSinceLast := h.State.CurrentByte - h.Progress.LastBytes
        h.Progress.CurrentSpeed = float64(bytesSinceLast) / intervalElapsed
    } else {
        h.Progress.CurrentSpeed = 0
    }

	// Update moving average for CurrentSpeed
    const maxSamples = 5 // Use last 5 samples for moving average
    h.Progress.SpeedSamples = append(h.Progress.SpeedSamples, h.Progress.CurrentSpeed)
        if len(h.Progress.SpeedSamples) > maxSamples {
        h.Progress.SpeedSamples = h.Progress.SpeedSamples[1:]
    }

	// Calculate currentSpeed using averaging samples we have 
    var speedSum float64
    for _, speed := range h.Progress.SpeedSamples {
        speedSum += speed
    }
    h.Progress.CurrentSpeed = speedSum / float64(len(h.Progress.SpeedSamples))

    // Average speed (from start to now)
    if totalElapsed > 0 {
        h.Progress.AvgSpeed = float64(h.State.CurrentByte) / totalElapsed
    } else {
        h.Progress.AvgSpeed = 0
    }

    h.Progress.LastUpdateTime = now
    h.Progress.LastBytes = h.State.CurrentByte
}

func (h *DownloadHandler) DisplayProgress() {
	h.Progress.Mutex.Lock()
	h.State.Mutex.Lock()
	defer h.Progress.Mutex.Unlock()
	defer h.State.Mutex.Unlock()

	percent := float64(h.State.CurrentByte) / float64(h.State.TotalBytes) * 100
	currentSpeedMBps := h.Progress.CurrentSpeed / (1024 * 1024)
	avgSpeedMBps := h.Progress.AvgSpeed / (1024 * 1024)
	fmt.Printf("Progress: %.2f%%, Current Speed: %.2f MB/s, Avg Speed: %.2f MB/s, Downloaded: %d/%d bytes\n",
		percent, currentSpeedMBps, avgSpeedMBps, h.State.CurrentByte, h.State.TotalBytes)
}