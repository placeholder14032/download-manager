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

	delta  time.Duration // the delta we want to calculate current speed with (eg. 0.5-1 sec)
	downloadedBytesInDelta     int64         
	timeStampStart     time.Time // we are saving last time we are calculating downloadedBytesInDelta from
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

func (pt *ProgressTracker) UpdateBytesDone(lastUpdate int64) {
	pt.Mutex.Lock()
	defer pt.Mutex.Unlock()

	now := time.Now()
	pt.LastUpdate = now
	pt.LastBytesDone = lastUpdate

	timePassed := now.Sub(pt.timeStampStart)


	if timePassed >= pt.delta {
		pt.downloadedBytesInDelta = lastUpdate - pt.BytesDone // Bytes since last paty
		pt.timeStampStart = now
	} else {
		pt.downloadedBytesInDelta += lastUpdate - pt.BytesDone 
	}

	pt.BytesDone = lastUpdate
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

	sec := pt.delta.Seconds()
	return float64(pt.downloadedBytesInDelta)/ sec
}

func (pt *ProgressTracker) GetProgress() float64{
	pt.Mutex.Lock()
	defer pt.Mutex.Unlock()

	if pt.TotalBytes == 0 {
		return 0
	}
	return float64(pt.BytesDone) / float64(pt.TotalBytes) * 100
}