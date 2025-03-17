package manager

import (
	"fmt"
	"os"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/util"
)

func (m *Manager) handleFailed(dl *download.Download) {
	dl.Status = download.Failed
	// TODO call clean up
}

func (m *Manager) handleEvent(e util.Event) {
	i, j := m.findDownloadQueueIndex(e.DownloadID)
	if i == -1 || j == -1 {
		fmt.Fprintf(os.Stderr, "Can't find download with id %d", e.DownloadID)
	}
	dl := &m.qs[i].DownloadLists[j] // not a copy but a pointer to the real one
	switch e.Type {
	case util.Starting:
		dl.Status = download.Downloading
	case util.Pausing:
		dl.Status = download.Paused
	case util.Resuming:
		dl.Status = download.Downloading
	case util.Failed:
		m.handleFailed(dl) // we have to clean up after failure
	case util.Finished:
		dl.Status = download.Done
	default:
		panic(fmt.Sprintf("unexpected util.EventType: %#v", e.Type))
	}
}

