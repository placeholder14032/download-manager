package controller

import (
	"time"

	"github.com/placeholder14032/download-manager/internal/queue"
	"github.com/placeholder14032/download-manager/internal/util"
)

var(
	DEFAULT_START_TIME = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	DEFAULT_END_TIME = time.Date(0, 1, 1, 23, 59, 59, 0, time.UTC)
)

func createBody(id int64, dir, name string, maxSim, maxBandWidth, maxRetry int64,
	hasTimeConstraint bool, start, end string) util.QueueBody {
	var sttime, edtime time.Time
	if hasTimeConstraint {
		sttime, _ = time.Parse(time.TimeOnly, start)
		edtime, _ = time.Parse(time.TimeOnly, end)
	} else {
		sttime = DEFAULT_START_TIME
		edtime = DEFAULT_END_TIME
	}
	//
	return 	util.QueueBody{
			ID: id,
			Directory: dir,
			Name: name,
			MaxSimul: maxSim,
			MaxBandWidth: maxBandWidth,
			MaxRetries: maxRetry,
			HasTimeConstraint: hasTimeConstraint,
			TimeRange: queue.TimeRange{Start: sttime, End: edtime},
		}
}

func AddQueue(dir, name string, maxSim, maxBandWidth, maxRetry int64,
	hasTimeConstraint bool, start, end string) error {
	req := util.Request{
		Type: util.AddQueue,
		Body: createBody(-1, dir, name, maxSim, maxBandWidth, maxRetry, hasTimeConstraint, start, end),
	}
	resp := SendReq(req)
	return returnResp(resp)
}

func DeleteQueue(qid int64) error {
	req := util.Request{
		Type: util.DeleteQueue,
		Body: util.QueueBody{
			ID: qid,
		},
	}
	resp := SendReq(req)
	return returnResp(resp)
}

func EditQueue(id int64, dir, name string, maxSim, maxBandWidth, maxRetry int64,
	hasTimeConstraint bool, start, end string) error {
	req := util.Request{
		Type: util.EditQueue,
		Body: createBody(id, dir, name, maxSim, maxBandWidth, maxRetry, hasTimeConstraint, start, end),
	}
	resp := SendReq(req)
	return returnResp(resp)
}

func GetQueues() []util.QueueBody {
	req := util.Request{
		Type: util.GetQueues,
	}
	resp := SendReq(req)
	body, ok := resp.Body.(util.StaticQueueList)
	if !ok {
		return []util.QueueBody{}
	}
	return body.Queues
}


