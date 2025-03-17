package queue

import (
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

