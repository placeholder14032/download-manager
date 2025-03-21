package manager

import (
	"fmt"

	"github.com/placeholder14032/download-manager/internal/util"
)

const (
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
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Start Download", "BodyModDownload"))
		return
	}
	err := m.startDownload(body.ID)
	m.answerERR(err)
}

func (m *Manager) answerPauseDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Pause Download", "BodyModDownload"))
		return
	}
	err := m.pauseDownload(body.ID)
	m.answerERR(err)
}

func (m *Manager) answerResumeDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Resume Download", "BodyModDownload"))
		return
	}
	err := m.resumeDownload(body.ID)
	m.answerERR(err)
}

func (m *Manager) answerRetryDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Retry Download", "BodyModDownload"))
		return
	}
	err := m.retryDownload(body.ID)
	m.answerERR(err)
}

func (m *Manager) answerCancelDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Cancel Download", "BodyModDownload"))
		return
	}
	err := m.cancelDownload(body.ID)
	m.answerERR(err)
}

func (m *Manager) answerDeleteDL(r util.Request) {
	body, ok := r.Body.(util.BodyModDownload)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "Delete Download", "BodyModDownload"))
		return
	}
	err := m.deleteDownload(body.ID)
	m.answerERR(err)
}

// creates a static queue list and returns it in response
func (m *Manager) answerGetQueues(r util.Request) {
	body := util.StaticQueueList{Queues: make([]util.QueueBody, len(m.qs))}
	for i, q := range m.qs {
		body.Queues[i] = convertToStaticQueue(&q)
	}
	resp := util.Response{
		Type: util.OK,
		Body: body,
	}
	m.resps <- resp
}

// returns a list of all downloads and their respective states.
// maybe should sort them by state? dunno TODO
func (m *Manager) answerGetDLS(r util.Request) {
	count := 0 // first we count the number of total downloads to create a slice. I dont really like using the damn append on them
	for _, q := range m.qs {
		count += len(q.DownloadLists)
	}
	// now that we have the count we create a similar slice and add all download representives to it.
	body := util.StaticDownloadList{
		Downloads: make([]util.DownloadBody, count),
	}
	i := 0
	for _, q := range m.qs {
		for _, d := range q.DownloadLists {
			body.Downloads[i] = convertToStaticDownload(&d, q.Name)
			i++
		}
	}
	resp := util.Response {
		Type: util.OK,
		Body: body,
	}
	m.resps <- resp
}

func (m *Manager) answerAddQ(r util.Request) {
	body, ok := r.Body.(util.QueueBody)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "AddQueue", "QueueBody"))
		return
	}
	err := m.addQueue(body)
	m.answerERR(err)
}

func (m *Manager) answerEditQ(r util.Request) {
	body, ok := r.Body.(util.QueueBody)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "ModQueue", "QueueBody"))
		return
	}
	err := m.editQueue(body)
	m.answerERR(err)
}

func (m *Manager) answerDelQ(r util.Request) {
	body, ok := r.Body.(util.QueueBody)
	if !ok {
		m.answerBadRequest(fmt.Sprintf(BAD_REQ_BODY_TYPE, "DelQueue", "QueueBody"))
		return
	}
	err := m.delQueue(body)
	m.answerERR(err)
}

func (m *Manager) answerRequest(r util.Request) {
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
	case util.CancelDownload:
		m.answerCancelDL(r)
	case  util.DeleteDownload:
		m.answerDeleteDL(r)
	//
	case util.AddQueue:
		m.answerAddQ(r)
	case util.EditQueue:
		m.answerEditQ(r)
	case util.DeleteQueue:
		m.answerDelQ(r)
	//
	case util.GetDownloads:
		m.answerGetDLS(r)
	case util.GetQueues:
		m.answerGetQueues(r)
	default:
		panic(fmt.Sprintf("unexpected util.RequestType: %#v", r.Type))
	}
}

