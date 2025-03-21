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
	elapsed := time.Since(lr.startTime).Seconds()

	// Effective rate so far
    effectiveRate := float64(lr.bytesRead) / elapsed
    if effectiveRate > float64(lr.limit) {
        // We've read too fast; calculate how long we *should* have taken
        expectedDuration := float64(lr.bytesRead) / float64(lr.limit)
        sleepDuration := time.Duration((expectedDuration - elapsed) * float64(time.Second))
        if sleepDuration > 0 {
            fmt.Printf("LimitedReader: Read %d bytes, total %d, elapsed %.2f s, rate %.2f B/s, sleeping for %.2f ms\n",
                n, lr.bytesRead, elapsed, effectiveRate, float64(sleepDuration)/float64(time.Millisecond))
            time.Sleep(sleepDuration)
        }
    } else {
        fmt.Printf("LimitedReader: Read %d bytes, total %d, elapsed %.2f s, rate %.2f B/s (within limit)\n",
            n, lr.bytesRead, elapsed, effectiveRate)
    }

    // Reset every second
    if elapsed >= 1 {
        fmt.Printf("LimitedReader: Resetting counters after %.2f s\n", elapsed)
        lr.bytesRead = 0
        lr.startTime = time.Now()
    }
}
    return n, err
}