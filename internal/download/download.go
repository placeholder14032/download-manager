package download

import (
	"encoding/json"
	"net/http"
)


const(
	HANDLER_NAME = "Handler"
	CHUNK_SIZE = 1024 * 1024 // 1mb chunks
	WORKER_COUNT = 8
)

type Download struct {
	ID           int64
	URL          string
	FilePath     string
	Status       State
	RetryCount   int64
	MaxRetries   int64


	Handler		DownloadHandler `json:"-"`
}

type DownloadAlias Download

type downloadRepresentation struct {
	*DownloadAlias

	SavedState SavedDownloadState
}

func (d *Download) GetProgress() float64 {
	return d.Handler.Progress.GetProgress()
}

func (d *Download) GetSpeed() string {
	// formatted speed
	return d.Handler.Progress.CurrentSpeedFormatted()
}

func CreateDefaultHandler(d *Download) {
	d.Handler = *d.NewDownloadHandler(&http.Client{Timeout: 0}, CHUNK_SIZE, WORKER_COUNT, 0)
	// TODO check bandwidth limit because its buggy
}

func (d Download) MarshalJSON() ([]byte, error) {
	dlst, _ := d.Handler.Export()
	rep := downloadRepresentation{
		DownloadAlias: (*DownloadAlias)(&d),
		SavedState: *dlst,
	}
	return json.Marshal(rep)
}

func (d *Download) UnmarshalJson(bts []byte) (error) {
	var rep = downloadRepresentation{
		DownloadAlias: (*DownloadAlias)(d),
	}
	if err := json.Unmarshal(bts, &rep); err != nil {
		return err
	}
	hd, err := Import(&rep.SavedState, &http.Client{Timeout: 0})
	if err != nil {
		return err
	}
	d.Handler = *hd
	return nil
}

