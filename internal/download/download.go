package download


type Download struct {
	ID           int64
	URL          string
	FilePath     string
	Status       State
	Progress     int64
	RetryCount   int64
	MaxRetries   int64


	Handler		DownloadHandler
}

