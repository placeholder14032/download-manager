package download

import (
    "io"
    "time"
)

type LimitedReader struct {
    reader     io.Reader
    limit      int64 // Bytes per second
	bytesRead      int64
    startTime      time.Time
}

func NewLimitedReader(r io.Reader, limit int64) *LimitedReader {
    return &LimitedReader{
		reader:         r,
        limit: 			limit,
        startTime:      time.Now(),
    }
}

func (lr *LimitedReader) Read(p []byte) (n int, err error) {
n, err = lr.reader.Read(p)
if n > 0 { // n=0 means no limit
	lr.bytesRead += int64(n)

	// Calculate how long we *should* have taken to read this many bytes
	expectedDuration := float64(lr.bytesRead) / float64(lr.limit) // in seconds
	elapsed := time.Since(lr.startTime).Seconds()

	// If we've read too fast, sleep to throttle the speed
	if elapsed < expectedDuration {
		sleepDuration := time.Duration((expectedDuration - elapsed) * float64(time.Second))
		time.Sleep(sleepDuration)
	}

	// Reset the start time and bytes read if we've waited for at least a second
	if elapsed >= 1 {
		lr.bytesRead = 0
		lr.startTime = time.Now()
	}
}
return n, err
}