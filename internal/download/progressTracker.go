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

}

func (pt *ProgressTracker) Update() {

}

func (pt *ProgressTracker) OverallSpeed() float64 {}

func (pt *ProgressTracker) CurrentSPeed() float64{}

func (pt *ProgressTracker) GetProgress() float64{}