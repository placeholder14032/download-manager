package manager

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/queue"
	"github.com/placeholder14032/download-manager/internal/util"
)

const (
	CANT_FIND_DL_ERROR = "can't find download with id: %d"
	DOWNLOAD_IS_NOT_IN_STATE = "download with id %d is not in state: %s"
	DOWNLOAD_IS_RUNNING = "download with id %d is still running"
	DOWNLOADS_ARE_RUNNING = "downloads are running in queueu: %d: can not modify"
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

func determineFilePath(directory string, url string) string {
	// joins the directory with the filename
	// if the directory doesn't have the last slash (/) it will usse the parent
	// because it is seen as a file in that case
	return path.Join(directory, path.Base(url))
	// changed from Path.Dir(Directory) because it might cause problems with omitting the last folder
}

func createDownload(dlID int64, url string, filePath string, maxRetry int64) download.Download {
	return download.Download {
		ID: dlID,
		URL: url,
		FilePath: filePath,
		MaxRetries: maxRetry,
		Status: download.Pending,
		RetryCount: 0,
	}
}

func createDefaultHandler(d *download.Download) {
	d.Handler = *d.NewDownloadHandler(&http.Client{Timeout: 0}, CHUNK_SIZE, 8, 0)
	// TODO check bandwidth limit because its buggy
}

// this takes in a pointer just so we dont have dangling copies of everything
// that needs to be cleaned up but there is really no actual need
func convertToStaticQueue(q *queue.Queue) util.QueueBody {
	return util.QueueBody{
		ID: q.ID,
		Directory: q.SaveDir,
		MaxSimul: q.MaxConcurrent,
		MaxBandWidth: q.MaxBandwidth,
		MaxRetries: q.MaxRetries,
		TimeRange: q.TimeRange,
	}
}

func convertToStaticDownload(d *download.Download) util.DownloadBody {
	return util.DownloadBody{
		ID: d.ID,
		URL: d.URL,
		FilePath: d.FilePath,
		Status: d.Status,
		Progress: d.GetProgress(),
		Speed: d.GetSpeed(),
	}
}

func checkRunningDL(d download.Download) bool {
	return d.Status == download.Downloading || d.Status == download.Paused || d.Status == download.Retrying
}

func checkRunningDLsInQueue(q queue.Queue) bool {
	// checks if any downloads are running to stop queue modification
	for _, dl := range q.DownloadLists {
		if checkRunningDL(dl) {
			return false
		}
	}
	return true
}

func getDownloadStarted(dl *download.Download, echan chan util.Event) {
	dl.Status = download.Downloading
	err := dl.Handler.StartDownloading()
	if err == nil {
		echan <- util.Event{Type: util.Finished, DownloadID: dl.ID}
	} else {
		echan <- util.Event{Type: util.Failed, DownloadID: dl.ID}
	}
	// this writing to channel will block the current goroutine
	// but it's okay because the handler is running in the parent one
	// and is not blocked
}

func cleanUp(path string) {
	// works like the part combiner
	partFiles, err := filepath.Glob(fmt.Sprintf("%s.part*", path))
	if err != nil {
		fmt.Println("cant find parts to clean up for ", path)
		return
	}
	for _, file := range partFiles {
		if err := os.Remove(file); err != nil {
			fmt.Printf("Warning: failed to remove part file %s: %v\n", file, err)
		}
	}
}

func (m *Manager) addDownload(qID int64, url string) error {
	i := m.findQueueIndex(qID)
	if i == -1 {
		return fmt.Errorf("Bad queue id: %d", qID)
	}
	dl := createDownload(m.lastUID, url, determineFilePath(m.qs[i].SaveDir, url), 0)
	createDefaultHandler(&dl)
	m.lastUID++
	m.qs[i].DownloadLists = append(m.qs[i].DownloadLists, dl)
	return nil
}

func (m *Manager) startDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf(CANT_FIND_DL_ERROR, dlID)
	}
	dl := &m.qs[i].DownloadLists[j] // pointer to the real download
	if dl.Status != download.Pending {
		return fmt.Errorf(DOWNLOAD_IS_NOT_IN_STATE, dlID, "Pending")
	}
	go getDownloadStarted(dl, m.events)
	return nil
}

func (m *Manager) pauseDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf(CANT_FIND_DL_ERROR, dlID)
	}
	dl := &m.qs[i].DownloadLists[j] // pointer to the real download
	if dl.Status != download.Downloading {
		return fmt.Errorf(DOWNLOAD_IS_NOT_IN_STATE, dlID, "Downloading")
	}
	err := dl.Handler.Pause()
	if err == nil {
		dl.Status = download.Paused
	} else {
		fmt.Println("cant pause", err)
	}
	return err
}

func (m *Manager) resumeDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf(CANT_FIND_DL_ERROR, dlID)
	}
	dl := &m.qs[i].DownloadLists[j] // pointer to the real download
	if dl.Status != download.Paused {
		return fmt.Errorf(DOWNLOAD_IS_NOT_IN_STATE, dlID, "Downloading")
	}
	err := dl.Handler.Resume(*dl) // returns error or nil
	if err == nil {
		dl.Status = download.Downloading
	} else {
		fmt.Println("cant resume", err)
		dl.Status = download.Failed
	}
	return err
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
	dl.Handler.Pause() // effectively this should kill all the workers because. also if there are non just ignore the returned error
	cleanUp(dl.FilePath) // cleans residual part files
	createDefaultHandler(dl)
	go getDownloadStarted(dl, m.events)
	return nil
}

func (m *Manager) cancelDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf(CANT_FIND_DL_ERROR, dlID)
	}
	dl := &m.qs[i].DownloadLists[j] // not a copy
	if dl.Status != download.Downloading {
		return fmt.Errorf(DOWNLOAD_IS_NOT_IN_STATE, dlID, "Downloading")
	}
	dl.Handler.Pause()
	cleanUp(dl.FilePath)
	createDefaultHandler(dl)
	dl.Status = download.Cancelled
	return nil
}

func (m *Manager) deleteDownload(dlID int64) error {
	i, j := m.findDownloadQueueIndex(dlID)
	if i == -1 || j == -1 {
		return fmt.Errorf(CANT_FIND_DL_ERROR, dlID)
	}
	dl := &m.qs[i].DownloadLists[j] // not a copy
	if checkRunningDL(*dl) {
		return fmt.Errorf(DOWNLOAD_IS_RUNNING, dl.ID)
	}
	m.qs[i].DownloadLists = util.Remove(m.qs[i].DownloadLists, j)
	return nil
}

// gets the settings from a body
// the id will be ignored so it should probably be -1
func (m *Manager) addQueue(body util.QueueBody) error {
	q := queue.Queue{
		ID: m.lastQID,
		Name: fmt.Sprintf("queue %d", m.lastQID),
		DownloadLists: make([]download.Download, 0),
		SaveDir: body.Directory,
		MaxConcurrent: body.MaxSimul,
		MaxBandwidth: body.MaxBandWidth,
		MaxRetries: body.MaxRetries,
		TimeRange: body.TimeRange,
	}
	m.lastQID++
	m.qs = append(m.qs, q)
	return nil
}

func (m *Manager) editQueue(body util.QueueBody) error {
	qid := body.ID;
	i := m.findQueueIndex(qid)
	if checkRunningDLsInQueue(m.qs[i]) {
		return fmt.Errorf(DOWNLOADS_ARE_RUNNING, qid)
	}
	m.qs[i].SaveDir = body.Directory
	m.qs[i].MaxConcurrent = body.MaxSimul
	m.qs[i].MaxBandwidth = body.MaxBandWidth
	m.qs[i].MaxRetries = body.MaxRetries
	m.qs[i].TimeRange = body.TimeRange
	return nil
}

func (m *Manager) delQueue(body util.QueueBody) error {
	qid := body.ID;
	i := m.findQueueIndex(qid)
	if checkRunningDLsInQueue(m.qs[i]) {
		return fmt.Errorf(DOWNLOADS_ARE_RUNNING, qid)
	}
	m.qs = util.Remove(m.qs, i)
	return nil
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

