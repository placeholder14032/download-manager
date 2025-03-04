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
	DownloadLists []download.Download
	SaveDir string
	MaxConcurrent int64
	MaxBandwith int64
	HasTimeConstraint bool
	TimeRange TimeRange
	MaxRetries int64
}
