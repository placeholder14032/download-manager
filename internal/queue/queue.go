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
