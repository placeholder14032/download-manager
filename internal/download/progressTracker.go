package download

import (
	"time"
	"sync"
)

type ProgressTracker struct {
	TotalBytes    int64
	BytesDone     int64
	StartTime     time.Time
	LastUpdate    time.Time
	LastBytesDone int64
	Mutex         sync.Mutex
}

func NewProgressTracker(totalBytes int64) *ProgressTracker {
	now := time.Now()
	return &ProgressTracker{
		TotalBytes: totalBytes,
		BytesDone: 0,
		StartTime: now,
		LastUpdate: now,
	}

}

func (pt *ProgressTracker) UpdateBytesDone(lastUpdate int64) {
	pt.Mutex.Lock()
	defer pt.Mutex.Unlock()

	now := time.Now()
	pt.LastUpdate = now
	pt.LastBytesDone = lastUpdate
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

func (pt *ProgressTracker) CurrentSPeed() float64{
	pt.Mutex.Lock()
	defer pt.Mutex.Unlock()
}

func (pt *ProgressTracker) GetProgress() float64{}