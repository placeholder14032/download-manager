package manager

import (
	"errors"
	"sync"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/queue"
	"github.com/placeholder14032/download-manager/internal/util"
)



type Event struct {
}

type Manager struct {
	mu sync.Mutex // used to protect the following fields
	qs []queue.Queue
	lastUID int64
	events chan Event
}

func (m *Manager) findQueueIndex(queueId int64) int { // maybe can be used to clean up some dublicate code
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < len(m.qs); i++ {
		if m.qs[i].ID == queueId {
			return i
		}
	}
	return -1
}

func (m *Manager) AddDownload(queueId int64, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	//
	for i := 0; i < len(m.qs); i++ {
		if m.qs[i].ID == queueId {
			m.qs[i].AddDownload(download.Download {
				ID: m.lastUID,
				URL: url,
			})
			m.lastUID++
			return nil
		}
	}
	return errors.New("bad queue id")
}

func (m *Manager) PauseDownload(queueId int64, dlId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock() // make sure to unlock after use
	//
	for i := 0; i < len(m.qs); i++ {
		if m.qs[i].ID == queueId {
			return m.qs[i].PauseDownload(dlId) // if it errors we return an error too
		}
	}
	return errors.New("bad queue id")
}

func (m *Manager) ResumeDownload(queueId int64, dlId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock() // make sure to unlock after use
	//
	for i := 0; i < len(m.qs); i++ {
		if m.qs[i].ID == queueId {
			return m.qs[i].ResumeDownload(dlId) // if it errors we return an error too
		}
	}
	return errors.New("bad queue id")
}

func (m *Manager) handleEvent(e Event) {
}

func (m *Manager) answerRequest(r util.Request) {
}

func (m *Manager) init() {
}

func (m *Manager) loadJson() {
	// TODO
}

func (m *Manager) Start(req chan util.Request, resps chan util.Response) {
	// start downloading unpaused downloads
	// Initialization
	m.init()
	// load json
	m.loadJson()
	// starting the main loop handling events and occasionally checking the whole state of things
	for {
		select {
		case e := <- m.events:
			m.handleEvent(e)
		case r := <- req:
			m.answerRequest(r)
		default:
		}
	}
}

