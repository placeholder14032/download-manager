package manager

import (
	"errors"
	"sync"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/queue"
)

type Manager struct {
	mu sync.Mutex // used to protect the following fields
	qs []queue.Queue
	lastUID int64
}

func loadJson() {
	// TODO
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


func (m *Manager) Start() {
	// load json
	// start downloading unpaused downloads
	loadJson()
}

