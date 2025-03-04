package download

type Download struct {
	ID           int64
	URL          string
	FileRath     string
	Status       State
	Progress     int64
	BytesWritten int64
	RetryCount   int64
	MaxRetries   int64
}
