package download

type State int

const (
	Pending State = iota
	Starting
	Downloading
	Paused
	Cancelled
	Failed
	Retrying
	Done
)
