package util

import "strconv"

type RequestType int

const (
	// download ones: these all have empty body and only OK/FAIL status
	AddDownload RequestType = iota
	StartDownload
	PauseDownload
	ResumeDownload
	CancelDownload
	RetryDownload
	DeleteDownload
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

var typeNames = []string{
	"Add Download",
	"Start Download",
	"Pause Download",
	"Resume Download",
	"Cancel Download",
	"Retry Download",
	"Delete Download",
	"Add Queue",
	"Delete Queue",
	"Edit Queue",
	"Get Queues",
	"Get Downlaods",
}

func (r RequestType) String() string{
	if 0 <= r && r <= GetDownloads {
		return typeNames[r]
	}
	return strconv.Itoa(int(r))
}

type Request struct {
	Type RequestType
	Body any // probably json or map[string]string
}

// specifying a bunch of types for bodies of request types so we can send their
// data over a channel and receive them

type BodyAddDownload struct {
	URL string
	QueueID int64
	FileName string // can be empty and I dunno maybe get it from the url
}

type BodyModDownload struct {
	// can be used for all of pause, resume, cancel, retry
	ID int64 // download id
}

