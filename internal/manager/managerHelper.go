package manager

import (
	"fmt"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/util"
)

func (m *Manager) findQueueIndex(qID int64) int { // maybe can be used to clean up some dublicate code
	for i, q := range m.qs {
		if q.ID == qID {
			return i
		}
	}
	return -1
}

func (m *Manager) findDownloadQueueIndex(dlID int64) (int, int) {
	for i, q := range m.qs {
		for j, d := range q.DownloadLists {
			if d.ID == dlID {
				return i, j
}
		}
	}
	return -1, -1
}

func createDownload(dlID int64, url string, filePath string, maxRetry int64) download.Download {
	return download.Download {
		ID: dlID,
		URL: url,
		FilePath: filePath, // TODO: derive from something. url or queue
		MaxRetries: maxRetry,
		Status: download.Pending,
		RetryCount: 0,
	}
}

func (m *Manager) addDownload(qID int64, url string) (int64, error) {
	i := m.findQueueIndex(qID)
	if i == -1 {
		return -1, fmt.Errorf("Bad queue id: %d", qID)
	}
	dl := createDownload(m.lastUID, url, "", 0)
	m.lastUID++
	m.qs[i].DownloadLists = append(m.qs[i].DownloadLists, dl)
	return dl.ID, nil
}

func (m *Manager) pauseDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf("can't find download with id: %d", dlID)
	}
	if m.qs[i].DownloadLists[j].Status != download.Downloading {
		return fmt.Errorf("download with id %d is not running", dlID)
	}
	err := m.hs[dlID].Pause()
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) resumeDownload(queueId int64, dlId int64) error {
	// TODO
	return nil
}

func (m *Manager) retryDownload(dlID int64) error {
	// TODO
	return nil
}

func (m *Manager) answerBadRequest(msg string) {
	resp := util.Response {
		Type: util.FAIL,
		Body: util.FailureMessage{Message: msg},
	}
	m.resps <- resp
}

func (m *Manager) answerOKRequest() {
	resp := util.Response{Type: util.OK, Body: nil}
	m.resps <- resp
}

