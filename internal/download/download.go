package download

type Download struct {
	ID            int64
	URL           string
	File_path     string
	Status        State
	Progress int64
	Bytes_written int64
	Retry_count   int64
	Max_retries   int64
}
