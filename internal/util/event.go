package util

type EventType int

const (
	Starting EventType = iota
	Pausing
	Resuming
	Finished
	Failed
)

type Event struct {
	Type EventType
	DownloadID int64
}

