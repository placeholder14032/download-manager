package download

import(
	"sync"
)

type Download struct {
	ID           int64
	URL          string
	FilePath     string
	Status       State
	Progress     int64
	RetryCount   int64
	MaxRetries   int64

	State         *DownloadState
	Handler		DownloadHandler
}

func (download *Download) Init (){
	download.State = &DownloadState{
		Completed:       make([]bool, 0),
		IncompleteParts: make([]chunk, 0),
		CurrentByte:     0,
		TotalBytes:      0,
		Mutex:           sync.Mutex{},
		isCombined:      false,
	}
	download.Progress = 0
	download.RetryCount = 0
}

