package manager

import (
	"fmt"
	"net/http"
	"time"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/util"
)

const (
	CHUNK_SIZE = 1024 * 1024 // 1mb chunks
	BAD_REQ_TYPE_NOT_BODY_ADD_DL = "bad request: type is add download but body is not of type BodyAddDownload"
	BAD_REQ_TYPE_NOT_BODY_MOD_DL = "bad request: type is mod download but body is not of type BodyModDownload"
)

func (m *Manager) answerAddDL(r util.Request) {
	body, ok := r.Body.(util.BodyAddDownload)
	if !ok {
		m.answerBadRequest(BAD_REQ_TYPE_NOT_BODY_ADD_DL)
		return
	}
	id, err := m.addDownload(body.QueueID, body.URL)
	if err != nil {
		m.answerBadRequest(err.Error())
		return
	}
	m.hs[id] = download.NewDownloadHandler(&http.Client{Timeout: 30 * time.Minute}, CHUNK_SIZE, 8) // arbitary worker count. TODO decide based on something else? a config file maybe
	m.answerOKRequest()
}

func (m *Manager) answerPauseDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(BAD_REQ_TYPE_NOT_BODY_MOD_DL)
		return
	}
	err := m.pauseDownload(body.ID)
	if err == nil {
		m.answerOKRequest()
	} else {
		m.answerBadRequest(err.Error())
	}
}

func (m *Manager) answerRequest(r util.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch r.Type {
	case util.AddDownload:
		m.answerAddDL(r)
	case util.AddQueue:
	case util.CancelDownload:
	case util.DeleteQueue:
	case util.EditQueue:
	case util.GetDownloads:
	case util.GetQueues:
	case util.PauseDownload:
		m.answerPauseDL(r)
	case util.ResumeDownload:
	case util.RetryDownload:
	default:
		panic(fmt.Sprintf("unexpected util.RequestType: %#v", r.Type))
	}
}


