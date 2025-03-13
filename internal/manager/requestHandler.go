package manager

import (
	"fmt"

	"github.com/placeholder14032/download-manager/internal/util"
)

const (
	CHUNK_SIZE = 1024 * 1024 // 1mb chunks
	BAD_REQ_BODY_TYPE = "bad request: type is %s but body is not of type %s"
)

func (m *Manager) answerERR(err error) {
	if err == nil {
		m.answerOKRequest()
	} else {
		m.answerBadRequest(err.Error())
	}
}

func (m *Manager) answerAddDL(r util.Request) {
	body, ok := r.Body.(util.BodyAddDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Add Download", "BodyAddDownload"))
		return
	}
	err := m.addDownload(body.QueueID, body.URL)
	m.answerERR(err)
}

func (m *Manager) answerStartDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Start Download", "BodyAddDownload"))
		return
	}
	err := m.startDownload(body.ID)
	m.answerERR(err)
}

func (m *Manager) answerPauseDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Pause Download", "BodyAddDownload"))
		return
	}
	err := m.pauseDownload(body.ID)
	m.answerERR(err)
}

func (m *Manager) answerResumeDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Resume Download", "BodyAddDownload"))
		return
	}
	err := m.resumeDownload(body.ID)
	m.answerERR(err)
}

func (m *Manager) answerRetryDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Retry Download", "BodyAddDownload"))
		return
	}
	err := m.retryDownload(body.ID)
	m.answerERR(err)
}

func (m *Manager) answerGetQueues(r util.Request) {
}

func (m *Manager) answerRequest(r util.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch r.Type {
	case util.AddDownload:
		m.answerAddDL(r)
	case util.StartDownload:
		m.answerStartDL(r)
	case util.PauseDownload:
		m.answerPauseDL(r)
	case util.ResumeDownload:
		m.answerResumeDL(r)
	case util.RetryDownload:
		m.answerRetryDL(r)
	//
	case util.AddQueue:
	case util.CancelDownload:
	case util.DeleteQueue:
	case util.EditQueue:
	//
	case util.GetDownloads:
	case util.GetQueues:
	default:
		panic(fmt.Sprintf("unexpected util.RequestType: %#v", r.Type))
	}
}

