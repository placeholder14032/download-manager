package download


type Download struct {
	ID           int64
	URL          string
	FilePath     string
	Status       State
	RetryCount   int64
	MaxRetries   int64


	Handler		DownloadHandler
}

func (d *Download) GetProgress() float64 {
	return d.Handler.Progress.GetProgress()
}

func (d *Download) GetSpeed() string {
	// formatted speed
	return d.Handler.Progress.CurrentSpeedFormatted()
}

