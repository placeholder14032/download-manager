package queue

import (
	"errors"
	"fmt"
	"time"

	"github.com/placeholder14032/download-manager/internal/download"
)

type TimeRange struct {
	Start time.Time
	End time.Time
}

type Queue struct{
	ID int64
	Name string
	DownloadLists []download.Download
	SaveDir string
	MaxConcurrent int64
	MaxBandwidth int64
	HasTimeConstraint bool
	TimeRange TimeRange
	MaxRetries int64
}

func (q *Queue) Init(ID int64) {
	q.ID = ID
	q.DownloadLists = make([]download.Download, 0)
	q.SaveDir = "~/Downloads/" // default. TODO change to sane defaults
	q.MaxConcurrent = 1
	q.HasTimeConstraint = false
	q.TimeRange = TimeRange{time.Time{}, time.Time{}}
	q.MaxRetries = 1
}

func (q *Queue) AddDownload(dl download.Download) { // dl passed by value
	q.DownloadLists = append(q.DownloadLists, dl)
}

func (q *Queue) PauseDownload(dlId int64) error {
	for i := 0; i < len(q.DownloadLists); i++ {
		if q.DownloadLists[i].ID == dlId {
			if q.DownloadLists[i].Status != download.Downloading {
				return errors.New(fmt.Sprintf("download with id %d is not currently downloading to be paused", dlId))
			}
			q.DownloadLists[i].Status = download.Paused
			return nil
		}
	}
	return errors.New("bad download id")
}

func (q *Queue) ResumeDownload(dlId int64) error {
	for i := 0; i < len(q.DownloadLists); i++ {
		if q.DownloadLists[i].ID == dlId {
			if q.DownloadLists[i].Status != download.Paused {
				return errors.New(fmt.Sprintf("download with id %d is not currently paused to be resumed", dlId))
			}
			q.DownloadLists[i].Status = download.Downloading
			return nil
		}
	}
	return errors.New("bad download id")
}

