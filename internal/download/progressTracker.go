package download

import (
	"fmt"
	"sync"
	"time"
)

type ProgressTracker struct {
	TotalBytes    int64
	BytesDone     int64
	StartTime     time.Time
	LastUpdate    time.Time
	Mutex         sync.Mutex

	delta                 time.Duration // the delta we want to calculate current speed with (eg. 0.5-1 sec)
	downloadedBytesInDelta int64        // bytes downloaded in current delta window
	timeStampStart        time.Time     // start time of current delta window
}

func NewProgressTracker(totalBytes int64, delta time.Duration) *ProgressTracker {
	now := time.Now()
	return &ProgressTracker{
		TotalBytes: totalBytes,
		BytesDone: 0,
		StartTime: now,
		LastUpdate: now,
		delta: delta,
		timeStampStart: now,
		downloadedBytesInDelta: 0,
	}

}

func (pt *ProgressTracker) UpdateBytesDone(currentBytes int64) {
    pt.Mutex.Lock()
    defer pt.Mutex.Unlock()

    if currentBytes > pt.TotalBytes {
        currentBytes = pt.TotalBytes
    }

    now := time.Now()
    bytesDelta := currentBytes - pt.BytesDone
    
    // Update speed calculation
    timePassed := now.Sub(pt.timeStampStart)
    if timePassed >= pt.delta {
        pt.downloadedBytesInDelta = bytesDelta
        pt.timeStampStart = now
    } else {
        pt.downloadedBytesInDelta += bytesDelta
    }

    pt.BytesDone = currentBytes
    pt.LastUpdate = now
}

func (pt *ProgressTracker) OverallSpeed() float64 {
	pt.Mutex.Lock()
	defer pt.Mutex.Unlock()

	elapsed := pt.LastUpdate.Sub(pt.StartTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(pt.BytesDone) / elapsed
}
func (pt *ProgressTracker) OverallSpeedFormated() string {
	return formatSpeed(pt.OverallSpeed())
}



func (pt *ProgressTracker) CurrentSpeed() float64 {
	pt.Mutex.Lock()
	defer pt.Mutex.Unlock()

	// Calculate actual time passed since last update
	timePassed := time.Since(pt.timeStampStart).Seconds()

	// Avoid division by zero
	if timePassed == 0 {
		return 0
	}

	// Calculate current speed using actual bytes downloaded in the delta period
	// This gives us a more accurate "instantaneous" speed measurement
	return float64(pt.downloadedBytesInDelta) / timePassed
}

func (pt *ProgressTracker) CurrentSpeedFormatted() string {
	return formatSpeed(pt.CurrentSpeed())
}

func (pt *ProgressTracker) GetProgress() float64{
	pt.Mutex.Lock()
	defer pt.Mutex.Unlock()

	if pt.TotalBytes == 0 {
		return 0
	}
	return float64(pt.BytesDone) / float64(pt.TotalBytes) * 100
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