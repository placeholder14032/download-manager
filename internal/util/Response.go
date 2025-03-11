package util

type ResponseType int

const (
	OK ResponseType = iota
	FAIL
)

type Response struct {
	Type ResponseType
	Body any // similar to Requests body
}

// these structs are used when aswering some requests

type StaticQueueList struct {
	Queues []QueueBody
}

type StaticDownloadList struct {
	Downloads []DownloadBody
}

