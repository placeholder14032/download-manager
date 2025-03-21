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
    Percent        float64
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
    h.Progress.Percent = float64(h.State.CurrentByte) / float64(h.State.TotalBytes) * 100

    // Average speed (from start to now)
    if totalElapsed > 0 {
        h.Progress.AvgSpeed = float64(h.State.CurrentByte) / totalElapsed
    } else {
        h.Progress.AvgSpeed = 0
    }

    h.Progress.LastUpdateTime = now
    h.Progress.LastBytes = h.State.CurrentByte
}

// func (h *DownloadHandler) DisplayProgress() {
// 	h.Progress.Mutex.Lock()
// 	h.State.Mutex.Lock()
// 	defer h.Progress.Mutex.Unlock()
// 	defer h.State.Mutex.Unlock()

// 	percent := float64(h.State.CurrentByte) / float64(h.State.TotalBytes) * 100
// 	currentSpeedMBps := h.Progress.CurrentSpeed / (1024 * 1024)
// 	avgSpeedMBps := h.Progress.AvgSpeed / (1024 * 1024)
// 	fmt.Printf("Progress: %.2f%%, Current Speed: %.2f MB/s, Avg Speed: %.2f MB/s, Downloaded: %d/%d bytes\n",
// 		percent, currentSpeedMBps, avgSpeedMBps, h.State.CurrentByte, h.State.TotalBytes)
// }

func (pt *ProgressTracker) GetCurrentSpeed() string {
	return formatSpeed(pt.CurrentSpeed)
}

func (pt *ProgressTracker) GetOverallSpeed() string {
	return formatSpeed(pt.AvgSpeed)
}

func (pt  *ProgressTracker) GetProgress() float64 {
    return pt.Percent
}
func formatSpeed(bytesPerSec float64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytesPerSec >= GB:
		return fmt.Sprintf("%.2f GB/s", bytesPerSec/float64(GB))
	case bytesPerSec >= MB:
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/float64(MB))
	case bytesPerSec >= KB:
		return fmt.Sprintf("%.2f KB/s", bytesPerSec/float64(KB))
	default:
		return fmt.Sprintf("%.2f B/s", bytesPerSec)
	}
}