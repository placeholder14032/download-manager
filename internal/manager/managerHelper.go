package manager

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/queue"
	"github.com/placeholder14032/download-manager/internal/util"
)

const (
	CANT_FIND_DL_ERROR = "can't find download with id: %d"
	DOWNLOAD_IS_NOT_IN_STATE = "download with id %d is not in state: %s"
	DOWNLOAD_IS_RUNNING = "download with id %d is still running"
	DOWNLOADS_ARE_RUNNING = "downloads are running in queueu: %d: can not modify"
	QUEUE_IS_FULL = "Queue with id %d is full and cant run anymore download until the others are finished"
	DIRECTORY_DOESNT_EXIST = "directory `%s` doesn't exist choose another one"
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

func checkDirExists(dir string) bool {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func checkParDirExists(filePath string) bool {
	return checkDirExists(path.Dir(filePath))
}

func determineFilePath(directory string, url string) string {
	// joins the directory with the filename
	// if the directory doesn't have the last slash (/) it will usse the parent
	// because it is seen as a file in that case
	return path.Join(directory, path.Base(url))
	// changed from Path.Dir(Directory) because it might cause problems with omitting the last folder
}

// returns the name if it's not empty. otherwise creates an unempty name for it
func chooseQueueName(name string, id int64) string {
	if name != "" {
		return name
	}
	return fmt.Sprintf("queue [%d]", id)
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

// this takes in a pointer just so we dont have dangling copies of everything
// that needs to be cleaned up but there is really no actual need
func convertToStaticQueue(q *queue.Queue) util.QueueBody {
	return util.QueueBody{
		ID: q.ID,
		Directory: q.SaveDir,
		MaxSimul: q.MaxConcurrent,
		MaxBandWidth: q.MaxBandwidth,
		MaxRetries: q.MaxRetries,
		HasTimeConstraint: q.HasTimeConstraint,
		TimeRange: q.TimeRange,
	}
}

func convertToStaticDownload(d *download.Download, q_name string) util.DownloadBody {
	return util.DownloadBody{
		ID: d.ID,
		URL: d.URL,
		FilePath: d.FilePath,
		Status: d.Status,
		Progress: d.GetProgress(),
		Speed: d.GetSpeed(),
		QueueName: q_name,
	}
}

func checkRunningDL(d download.Download) bool {
	return d.Status == download.Downloading || d.Status == download.Paused || d.Status == download.Retrying
}

func checkRunningDLsInQueue(q queue.Queue) bool {
	// checks if any downloads are running to stop queue modification
	for _, dl := range q.DownloadLists {
		if checkRunningDL(dl) {
			return true 
		}
	}
	return false
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

func checkTimeInRange(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func checkNowInRange(start, end time.Time) bool {
	now, _ := time.Parse(time.TimeOnly, time.Now().Format(time.TimeOnly))
	return checkTimeInRange(start, end, now)
}

func (m *Manager) addDownload(qID int64, url string) error {
	i := m.findQueueIndex(qID)
	if i == -1 {
		return fmt.Errorf("Bad queue id: %d", qID)
	}
	dl := createDownload(m.lastUID, url, determineFilePath(m.qs[i].SaveDir, url), m.qs[i].MaxRetries)
	download.CreateDefaultHandler(&dl)
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
	if !m.qs[i].IsSafeToRunDL() {
		return fmt.Errorf(QUEUE_IS_FULL, m.qs[i].ID)
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
	dl.Handler.Pause()
	dl.Status = download.Paused
	return nil
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
	if !m.qs[i].IsSafeToRunDL() {
		return fmt.Errorf(QUEUE_IS_FULL, m.qs[i].ID)
	}
	err := dl.Handler.Resume() // returns error or nil
	if err == nil {
		dl.Status = download.Downloading
	} else {
		fmt.Println("cant resume", err)
		dl.Status = download.Failed
	}
	return err
	return nil
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
	if !m.qs[i].IsSafeToRunDL() {
		return fmt.Errorf(QUEUE_IS_FULL, m.qs[i].ID)
	}
	dl.Status = download.Retrying // temporary status to stop other threads from meddling with this one even though there might not be any other threads probably
	dl.Handler.Pause() // effectively this should kill all the workers because. also if there are non just ignore the returned error
	cleanUp(dl.FilePath) // cleans residual part files
	download.CreateDefaultHandler(dl)
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
	download.CreateDefaultHandler(dl)
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
	if !checkDirExists(body.Directory) {
		return fmt.Errorf(DIRECTORY_DOESNT_EXIST, body.Directory)
	}
	q := queue.Queue{
		ID: m.lastQID,
		Name: chooseQueueName(body.Name, m.lastQID),
		DownloadLists: make([]download.Download, 0),
		SaveDir: body.Directory,
		MaxConcurrent: body.MaxSimul,
		MaxBandwidth: body.MaxBandWidth,
		MaxRetries: body.MaxRetries,
		HasTimeConstraint: body.HasTimeConstraint,
		TimeRange: body.TimeRange,
		Disabled: false,
	}
	m.lastQID++
	m.qs = append(m.qs, q)
	return nil
}

func (m *Manager) editQueue(body util.QueueBody) error {
	if !checkDirExists(body.Directory) {
		return fmt.Errorf(DIRECTORY_DOESNT_EXIST, body.Directory)
	}
	qid := body.ID;
	i := m.findQueueIndex(qid)
	if checkRunningDLsInQueue(m.qs[i]) {
		return fmt.Errorf(DOWNLOADS_ARE_RUNNING, qid)
	}
	m.qs[i].SaveDir = body.Directory
	m.qs[i].Name = body.Name
	m.qs[i].MaxConcurrent = body.MaxSimul
	m.qs[i].MaxBandwidth = body.MaxBandWidth
	m.qs[i].MaxRetries = body.MaxRetries
	m.qs[i].HasTimeConstraint = body.HasTimeConstraint
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

func (m *Manager) disableQueue(idx int) {
	m.qs[idx].Disabled = true
	fmt.Println("disabling queue", m.qs[idx].ID, time.Now())
	for _, dl := range m.qs[idx].DownloadLists {
		m.pauseDownload(dl.ID) // this is O(n^2) but at this point I dont really care
	}
}

func (m *Manager) enableQueue(idx int) {
	m.qs[idx].Disabled = false
	q := &m.qs[idx]
	fmt.Println("enabling queue", q.ID, time.Now())
	ln := len(q.DownloadLists)
	mx := q.MaxConcurrent
	if mx == 0 {
		mx = int64(ln)
	}
	ecnt := int64(0)
	for i := 0; ecnt < mx && i < ln; i++ {
		dl := &q.DownloadLists[i]
		if dl.Status == download.Pending {
			m.startDownload(dl.ID)
		}
		if dl.Status == download.Paused {
			m.resumeDownload(dl.ID)
		}
	}
}

func (m *Manager) checkQueueTimes() {
	for i := range m.qs {
		q := &m.qs[i]
		if !q.HasTimeConstraint {
			continue
		}
		inRange := checkNowInRange(q.TimeRange.Start, q.TimeRange.End)
		//
		if inRange && q.Disabled {
			m.enableQueue(i)
		} else if !inRange && !q.Disabled {
			m.disableQueue(i)
		}
	}
}

