package download

import (
    "io"
    "time"
	"fmt"
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
	fmt.Printf("LimitedReader: Read %d bytes, total %d, elapsed %.2f s, expected %.2f s\n", n, lr.bytesRead, elapsed, expectedDuration)

	// If we've read too fast, sleep to throttle the speed
	if elapsed < expectedDuration {
		sleepDuration := time.Duration((expectedDuration - elapsed) * float64(time.Second))
		fmt.Printf("LimitedReader: Read %d bytes, elapsed %.2f s, expected %.2f s, sleeping for %.2f ms\n",
			n, elapsed, expectedDuration, float64(sleepDuration)/float64(time.Millisecond))
		time.Sleep(sleepDuration)
	}

	// Reset counters every second to avoid problem stuff we faced
	elapsedSinceStart := time.Since(lr.startTime).Seconds()
	if elapsedSinceStart >= 1 {
		fmt.Printf("LimitedReader: Resetting counters after %.2f s\n", elapsed)
		lr.bytesRead = 0
		lr.startTime = time.Now()
	}
}
return n, err
}