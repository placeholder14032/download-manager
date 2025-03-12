package util

type RequestType int

const (
	// download ones: these all have empty body and only OK/FAIL status
	AddDownload RequestType = iota
	PauseDownload
	ResumeDownload
	CancelDownload
	RetryDownload
	// queue ones: these also have empty body
	AddQueue
	DeleteQueue
	EditQueue // will take parameters inside the body
	// possible parameters are
	// TimeRange
	// Directory
	// MaxTry
	// MaxSimul
	// MaxBandWidth
	//
	GetQueues // pass all of the queues with their downloads
	GetDownloads // pass all of the downloads
)

type Request struct {
	Type RequestType
	Body any // probably json or map[string]string
}

// specifying a bunch of types for bodies of request types so we can send their
// data over a channel and receive them

type BodyAddDownload struct {
	URL string
	QueueID int64
	FilePath string // can be empty and I dunno maybe get it from the url
}

type BodyModDownload struct {
	// can be used for all of pause, resume, cancel, retry
	ID int64 // download id
}

