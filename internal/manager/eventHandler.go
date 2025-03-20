package manager

import (
	"fmt"
	"os"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/util"
)

func (m *Manager) tryRun(cand *download.Download) bool {
	if cand.Status == download.Pending {
		m.startDownload(cand.ID)
		return true
	}
	if cand.Status == download.Paused {
		m.resumeDownload(cand.ID)
		return true
	}
	return false
}

func (m *Manager) runNext(i, j int) {
	ln := len(m.qs[i].DownloadLists)
	for k := j+1; k < ln; k++ {
		cand := m.qs[i].DownloadLists[k]
		if m.tryRun(&cand){
			return
		}
	}
	for k := 0; k < j; k++ {
		cand := m.qs[i].DownloadLists[k]
		if m.tryRun(&cand){
			return
		}
	}
}

func (m *Manager) handleFailed(dl *download.Download, i, j int) {
	cleanUp(dl.FilePath)
	if dl.RetryCount < dl.MaxRetries {
		dl.RetryCount++
		m.cancelDownload(dl.ID) // making sure everybody is dead
		m.retryDownload(dl.ID) // should work after cancel
	} else {
		dl.Status = download.Failed
		if m.qs[i].IsSafeToRunDL() {
			m.runNext(i, j)
		}
	}
}

func (m *Manager) handleFinished(dl *download.Download, i, j int) {
	dl.Status = download.Done
	if m.qs[i].IsSafeToRunDL() {
		m.runNext(i, j)
	}
}

func (m *Manager) handleEvent(e util.Event) {
	i, j := m.findDownloadQueueIndex(e.DownloadID)
	if i == -1 || j == -1 {
		fmt.Fprintf(os.Stderr, "Can't find download with id %d", e.DownloadID)
	}
	dl := &m.qs[i].DownloadLists[j] // not a copy but a pointer to the real one
	switch e.Type {
	case util.Pausing:
		dl.Status = download.Paused
	case util.Resuming:
		dl.Status = download.Downloading
	case util.Failed:
		m.handleFailed(dl, i, j) // we have to clean up after failure
	case util.Finished:
		m.handleFinished(dl, i, j)
	default:
		panic(fmt.Sprintf("unexpected util.EventType: %#v", e.Type))
	}
}

