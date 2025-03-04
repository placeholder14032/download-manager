package download

type State int

const (
	Paused State = iota
	Cancelled
	Done
	Downloading
	Starting
	Retrying
	Failed
)
