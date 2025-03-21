package download

type State int

const (
	Pending State = iota
	Downloading
	Paused
	Cancelled
	Failed
	Retrying
	Done
)
