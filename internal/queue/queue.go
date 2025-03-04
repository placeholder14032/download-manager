package queue

import (
	"time"
)

type TimeRange struct {
	Start time.Time
	End time.Time
}

type Queue struct{
	ID int64
	DownloadLists []Download
	SaveDir string
	MaxConcurrent int64
	MaxBandwith int64
	HasTimeConstraint bool
	TimeRange TimeRange
	MaxRetries int64
}