package manager

import (
	"sync"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/queue"
	"github.com/placeholder14032/download-manager/internal/util"
)

type Manager struct {
	mu      sync.Mutex // used to protect the following fields
	qs      []queue.Queue
	hs map[int64]*download.DownloadHandler
	lastUID int64
	events  chan util.Event
	req chan util.Request
	resps chan util.Response
}

func (m *Manager) handleEvent(e util.Event) {
}

func (m *Manager) sendAnswer() {
}

func (m *Manager) answerAddDownload(r util.Request) {
}

func (m *Manager) init() {
	m.qs = make([]queue.Queue, 0)
	m.hs = make(map[int64]*download.DownloadHandler)
	m.lastUID = 1
	m.events = make(chan util.Event)
}

func (m *Manager) loadJson() {
	// TODO
}

func (m *Manager) Start(req chan util.Request, resps chan util.Response) {
	// Initialization
	m.init()
	m.req = req
	m.resps = resps
	// start downloading unpaused downloads
	// load json
	m.loadJson()
	// starting the main loop handling events and occasionally checking the whole state of things
	for {
		select {
		case e := <- m.events:
			m.handleEvent(e)
		case r := <- req:
			m.answerRequest(r)
		}
	}
}
