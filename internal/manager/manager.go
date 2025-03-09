package manager

import "github.com/placeholder14032/download-manager/internal/queue"

type Manager struct {
	qs []queue.Queue
}

func loadJson() {
	// TODO
}

func (m *Manager) Start() {
	// load json
	// start downloading unpaused downloads
	loadJson()
}

