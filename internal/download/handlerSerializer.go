package download

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"context"
	"time"
)
type SavedDownloadState struct {
	URL             string
	FilePath        string
	CHUNK_SIZE      int64
	CompletedParts  []bool
	CurrByte        int64
	TotalBytes      int64
	PartsCount      int64 
	IsPaused        bool
	IncompleteParts []int64 
}

// Export: serializes the current state to SavedDownloadState
func (h *DownloadHandler) Export() (*SavedDownloadState, error) {
	h.State.Mutex.Lock()
	defer h.State.Mutex.Unlock()

	incompleteParts := make([]int64, 0, len(h.State.IncompleteParts))
	for _, chunk := range h.State.IncompleteParts {
		incompleteParts = append(incompleteParts, chunk.Start)
	}

	savedState := &SavedDownloadState{
		URL:             h.URL,
		FilePath:        h.FilePath,
		CHUNK_SIZE:      h.CHUNK_SIZE,
		CompletedParts:  h.State.Completed,
		CurrByte:        h.State.CurrentByte,
		TotalBytes:      h.State.TotalBytes,
		PartsCount:      h.PartsCount,
		IsPaused:        h.State.IsPaused,
		IncompleteParts: incompleteParts,
	}

	return savedState, nil
}

// Import: creates a DownloadHandler from a SavedDownloadState
func Import(state *SavedDownloadState, client *http.Client) (*DownloadHandler, error) {
	if state == nil {
		return nil, fmt.Errorf("invalid state: nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	handler := &DownloadHandler{
		Client:        client,
		CHUNK_SIZE:    state.CHUNK_SIZE,
		WORKERS_COUNT: 4, // default value for now -> later we can calculate it
		PartsCount:    state.PartsCount,
		URL:           state.URL,
		FilePath:      state.FilePath,
		PauseChan:     make(chan struct{}),
		ResumeChan:    make(chan struct{}),
		ctx:           ctx,
		cancel:        cancel,
		Progress: &ProgressTracker{
			StartTime: time.Now(),
		},
	}

	incompleteParts := make([]chunk, 0, len(state.IncompleteParts))
	for _, start := range state.IncompleteParts {
		end := start + state.CHUNK_SIZE - 1
		if end >= state.TotalBytes {
			end = state.TotalBytes - 1
		}
		incompleteParts = append(incompleteParts, chunk{Start: start, End: end})
	}

	handler.State = &DownloadState{
		Completed:       state.CompletedParts,
		IncompleteParts: incompleteParts,
		CurrentByte:     state.CurrByte,
		TotalBytes:      state.TotalBytes,
		Mutex:           sync.Mutex{},
		IsPaused:        state.IsPaused,
	}

	return handler, nil
}

// Serialize: converts the DownloadHandler state to JSON bytes
func (h *DownloadHandler) Serialize() ([]byte, error) {
	savedState, err := h.Export()
	if err != nil {
		return nil, fmt.Errorf("failed to export download state: %v", err)
	}

	data, err := json.Marshal(savedState)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize state: %v", err)
	}

	return data, nil
}

// DeserializeHandler: creates a DownloadHandler from JSON bytes
func DeserializeHandler(data []byte, client *http.Client) (*DownloadHandler, error) {
	var state SavedDownloadState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to deserialize state: %v", err)
	}

	return Import(&state, client)
}