package util

import (
	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/queue"
)

// this is a static representation of a queue
type QueueBody struct {
	// can be used when adding a queue or editing it
	// notice that you have to pass all of them when editing
	// which means you should first query them
	//
	// you know what? fuck it. use this for deleting queues as well
	// just pass the id and someone will handle it
	// case closed.
	//
	// on third thought this can be used when passing back the list of queues as well
	// I'm just gonna be passing this struct everywhere man
	//
	ID int64 // optional. can be -1 meaning to add a new one. otherwise do something with it
	Directory string
	Name string // queues name? might be optional
	MaxSimul int64
	MaxBandWidth int64
	MaxRetries int64
	HasTimeConstraint bool
	TimeRange queue.TimeRange
}

// similar thing for a download
type DownloadBody struct {
	ID int64
	URL string
	FilePath string
	Status download.State
	Progress float64 // percentage
	Speed string // formatted string for speed
}

// this is a function used to remove an element from a slice
// which is probably used in many places.
// the basic idea is using append to splice together two parts
// from before and after the index we want to remove

func Remove[T any](slice []T, idx int) []T {
	return append(slice[:idx], slice[idx+1:]...)
}

