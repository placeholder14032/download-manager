package util

type EventType int

const (
	Pausing EventType = iota
	Resuming
	Finished
	Failed
)

type Event struct {
	Type EventType
	DownloadID int64
}

