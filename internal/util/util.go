package util

type RequestType int

const (
	AddDownload RequestType = iota
	PauseDownload
	ResumeDownload
	CancelDownload
	RetryDownload
	// queue ones
	AddQueue
	DeleteQueue
	EditTimeRange
	EditDirectory
	EditMaxTry
)

type Request struct {
	Type RequestType
	body any // probably json or map[string]string
}

type Response struct {
	body any // similar to Requests body
}

