package manager

import (
	"fmt"
	"net/http"
	"time"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/util"
)

const (
	CANT_FIND_DL_ERROR = "can't find download with id: %d"
	DOWNLOAD_IS_NOT_IN_STATE = "download with id %d is not in state: %s"
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

func (m *Manager) getHandler(dlID int64) *download.DownloadHandler {
	if m.hs[dlID] == nil {
		m.hs[dlID] = createDefaultHandler()
	}
	return m.hs[dlID]
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

func createDefaultHandler() *download.DownloadHandler {
	return download.NewDownloadHandler(
		&http.Client{Timeout: 30 * time.Minute},
		CHUNK_SIZE,
		8,
		) // arbitary worker count. TODO decide based on something else? a config file maybe
}

func (m *Manager) addDownload(qID int64, url string) error {
	i := m.findQueueIndex(qID)
	if i == -1 {
		return fmt.Errorf("Bad queue id: %d", qID)
	}
	dl := createDownload(m.lastUID, url, "", 0)
	m.lastUID++
	m.qs[i].DownloadLists = append(m.qs[i].DownloadLists, dl)
	m.hs[dl.ID] = createDefaultHandler()
	return nil
}

func (m *Manager) startDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf(CANT_FIND_DL_ERROR, dlID)
	}
	dl := m.qs[i].DownloadLists[j] // just a copy of the real download
	if dl.Status != download.Pending {
		return fmt.Errorf(DOWNLOAD_IS_NOT_IN_STATE, dlID, "Pending")
	}
	return m.getHandler(dlID).StartDownloading(dl) // returns error or nil
}

func (m *Manager) pauseDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf(CANT_FIND_DL_ERROR, dlID)
	}
	if m.qs[i].DownloadLists[j].Status != download.Downloading {
		return fmt.Errorf(DOWNLOAD_IS_NOT_IN_STATE, dlID, "Downloading")
	}
	return m.getHandler(dlID).Pause() // returns error or nil
}

func (m *Manager) resumeDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf(CANT_FIND_DL_ERROR, dlID)
	}
	dl := m.qs[i].DownloadLists[j] // this is a copy of struct, chaning anything wont affect the real download
	if dl.Status != download.Paused {
		return fmt.Errorf(DOWNLOAD_IS_NOT_IN_STATE, dlID, "Paused")
	}
	return m.getHandler(dlID).Resume(dl)
}

func (m *Manager) retryDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf(CANT_FIND_DL_ERROR, dlID)
	}
	dl := &m.qs[i].DownloadLists[j] // not a copy
	if dl.Status != download.Cancelled && dl.Status != download.Failed {
		return fmt.Errorf(DOWNLOAD_IS_NOT_IN_STATE, dlID, "Cancelled or Failed")
	}
	dl.Status = download.Retrying // temporary status to stop other threads from meddling with this one even though there might not be any other threads probably
	m.getHandler(dlID).Pause() // effectively this should kill all the workers because
	m.hs[dlID] = createDefaultHandler()
	return m.hs[dlID].StartDownloading(m.qs[i].DownloadLists[j]) // will return error or nil
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

