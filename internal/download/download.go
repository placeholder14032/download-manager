package download

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

func (d *Download) GetProgress() int64 {
	// TODO
	return 0
}

func (d *Download) GetSpeed() int64 {
	// TODO
	return 0
}

