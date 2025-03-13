package download

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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

func Import(state *SavedDownloadState, client *http.Client)  (*DownloadHandler, error)  {
	if state == nil {
        return nil, fmt.Errorf("invalid state: nil")
    }

	// new handler to return
	handler := &DownloadHandler{
        Client:        client,
        CHUNK_SIZE:    state.CHUNK_SIZE,
        WORKERS_COUNT: 4, // Default value or could be added to SavedDownloadState
        PartsCount:    state.PartsCount,
        PauseChan:     make(chan struct{}),
    }

	// before creating and returning state we need to 
	// create chunk slices for incomplete chuncks instead of int slice we had
	incompleteParts := make([]chunk,0)
	for _, start := range state.IncompleteParts {
        end := start + state.CHUNK_SIZE - 1
        if end > int(state.TotalBytes) {
            end = int(state.TotalBytes)
        }
        incompleteParts = append(incompleteParts, chunk{Start: start, End: end})
    }

	handler.State = &DownloadState{
        Completed:       state.CompletedParts,
        IncompleteParts: incompleteParts,
        CurrentByte:     state.CurrByte,
        TotalBytes:      state.TotalBytes,
        Mutex:          sync.Mutex{},
        IsPaused:       state.IsPaused,
        isCombined:     false,
    }
	
	return handler, nil
}

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


func SerializeHandler(handler *DownloadHandler, download *Download) ([]byte, error) {
	// exporting handler -> same as creating a state re port
    savedState, err := handler.Export(download)
    if err != nil {
        return nil, fmt.Errorf("failed to export handler state: %v", err)
    }

    // convert state to json -> write a json file out of the state we exported from handler
    data, err := json.Marshal(savedState)
    if err != nil {
        return nil, fmt.Errorf("failed to serialize state: %v", err)
    }

    return data, nil
}

func DeserializeHandler(data []byte, client *http.Client) (*DownloadHandler, error) {
    // first we need to read saved state fom json file
    var state SavedDownloadState
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, fmt.Errorf("failed to deserialize state: %v", err)
    }

    // creating handler out of that saved state we exported from json file
    return Import(&state, client)
}