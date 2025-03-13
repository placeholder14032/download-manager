package download

import (
	"encoding/json"
	"fmt"
)

type SavedDownloadState struct{
	URL string
	FilePath string
	CHUNK_SIZE int
	CompletedParts []bool
	CurrByte int64
	TotalBytes int64
	PartsCount int
	IsPaused bool
	IncompleteParts []int
}

func (h *DownloadHandler) Export(download *Download) (*SavedDownloadState, error){
		h.State.Mutex.Lock()
		defer h.State.Mutex.Unlock()

		savedState := &SavedDownloadState{
			URL: download.URL,
			FilePath: download.FilePath,
			CHUNK_SIZE: h.CHUNK_SIZE,
			CompletedParts: h.State.Completed,
			CurrByte: h.State.CurrentByte,
			TotalBytes: h.State.TotalBytes,
			PartsCount: h.PartsCount,
			IsPaused: h.State.IsPaused,
			IncompleteParts: make([]int, 0),
		}

		return savedState, nil
}

// func (h *DownloadHandler) Import(state *SavedDownloadState) (*Download, error){}

// same as save
func (h *DownloadHandler) Serialize(download *Download) ([]byte, error){
	savedState, err := h.Export(download)
	if err != nil{
		return nil, fmt.Errorf("failed to export download state: %v", err)
	}

	data , err := json.Marshal(savedState)
	if err != nil {
        return nil, fmt.Errorf("failed to serialize state: %v", err)
    }

	return data, nil
}
// same as load
// func (h *DownloadHandler) Deserialize(data []byte) (*Download, error){}