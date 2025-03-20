package manager

import (
	"sync"
	"time"

	"github.com/placeholder14032/download-manager/internal/queue"
	"github.com/placeholder14032/download-manager/internal/util"
)

type Manager struct {
	mu      sync.Mutex // used to protect the following fields
	qs      []queue.Queue
	lastUID int64
	lastQID int64
	events  chan util.Event
	req chan util.Request
	resps chan util.Response
}

func (m *Manager) init() {
	m.qs = make([]queue.Queue, 0)
	m.lastUID = 1
	m.lastQID = 1
	m.events = make(chan util.Event, 10) // making buffer size bigger just to be safe
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
	// creating a timer to check stuff on a frequent basis
	minTimer := time.NewTicker(time.Minute) // ticks every minute
	// starting the main loop handling events and occasionally checking the whole state of things
	for {
		select {
		case e := <- m.events:
			m.handleEvent(e)
		case r := <- req:
			m.answerRequest(r)
		case <- minTimer.C:
			m.checkQueueTimes()
		}
	}
}
