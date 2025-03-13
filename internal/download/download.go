package download

import(
)

type Download struct {
	ID           int64
	URL          string
	FilePath     string
	Status       State
	Progress     int64
	BytesWritten int64
	RetryCount   int64
	MaxRetries   int64
}
