package manager

import (
	//"encoding/json"

	"encoding/json"
	"os"

	"github.com/placeholder14032/download-manager/internal/queue"
)

const SAVE_FILE = "save.json"

type saveFile struct {
	LastDLID int64
	LastQID int64
	Queues []queue.Queue
}

func (m *Manager) WriteJson() {
	file, err := os.Create(SAVE_FILE)
	if err != nil {
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	data := saveFile{
		LastDLID: m.lastUID,
		LastQID: m.lastQID,
		Queues: m.qs,
	}
	encoder.Encode(data)
}

func (m *Manager) LoadJson() {
	file, err := os.Open(SAVE_FILE)
	if err != nil {
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var data saveFile
	decoder.Decode(&data)
	m.lastQID = data.LastQID
	m.lastUID = data.LastDLID
	m.qs = data.Queues
}

