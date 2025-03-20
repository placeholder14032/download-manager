package demo

import (
	"fmt"
	"time"

	"github.com/placeholder14032/download-manager/internal/controller"
	"github.com/placeholder14032/download-manager/internal/util"
)

var(
	DEFAULT_START_TIME = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	DEFAULT_END_TIME = time.Date(0, 1, 1, 23, 59, 59, 0, time.UTC)
)

func printRequestTypes() {
	for i := 0; i <= int(util.GetDownloads); i++ {
		fmt.Printf("[%d] %s\n", i, util.RequestType(i))
	}
}

func askRequestType() util.RequestType {
	fmt.Println("please enter your request:")
	printRequestTypes()
	fmt.Print("> ")
	var a int
	fmt.Scanf("%d", &a)
	return util.RequestType(a)
}

func askAddDL() util.Request {
	body := util.BodyAddDownload{}
	fmt.Print("please enter the url (no white space please ^w^): ")
	fmt.Scanf("%s", &body.URL)
	fmt.Print("please enter the queue id you want to add this to: ")
	fmt.Scanf("%d", &body.QueueID)
	return util.Request{
		Type: util.AddDownload,
		Body: body,
	}
}

func askModDL(t util.RequestType) util.Request {
	var id int64
	fmt.Print("please enter the download id: ")
	fmt.Scanf("%d", &id)
	return util.Request{
		Type: t,
		Body: util.BodyModDownload{ID: id},
	}
}

func askAddQueue() util.Request {
	body := util.QueueBody{}
	fmt.Print("please enter save directory for this queue: ")
	fmt.Scanf("%s", &body.Directory)
	return util.Request{
		Type: util.AddQueue,
		Body: body,
	}
}

func askDelQueue() util.Request {
	body := util.QueueBody{}
	fmt.Print("please enter queue id: ")
	fmt.Scanf("%d", &body.ID)
	return util.Request{
		Type: util.AddQueue,
		Body: body,
	}
}

func readTime(prompt string, def time.Time) time.Time {
	var tmps string
	fmt.Printf(prompt, time.TimeOnly)
	fmt.Scanf("%s", &tmps)
	start, err := time.Parse(time.TimeOnly, tmps)
	if err != nil {
		return def
	}
	return start
}

func askEditQueue() util.Request {
	body := util.QueueBody{}
	fmt.Print("please enter the queue id: ")
	fmt.Scanf("%d", &body.ID)
	fmt.Print("please enter the save directory: ")
	fmt.Scanf("%s", &body.Directory)
	fmt.Print("please enter the max number of retries: ")
	fmt.Scanf("%d", &body.MaxRetries)
	fmt.Print("please enter the max bandwidth (in bytes per second): ")
	fmt.Scanf("%d", &body.MaxBandWidth)
	fmt.Print("please enter the max number of concurrent downloads: ")
	fmt.Scanf("%d", &body.MaxSimul)
	//
	fmt.Print("does this queue have time constraint? [y/n] ")
	var tmps string
	fmt.Scanf("%s", &tmps)
	if tmps == "y" {
		body.HasTimeConstraint = true
		body.TimeRange.Start = readTime("please enter the start of active time range (format %s): ", DEFAULT_START_TIME)
		body.TimeRange.End = readTime("please enter the start of active time range (format %s): ", DEFAULT_END_TIME)

	} else {
		body.HasTimeConstraint = false
		body.TimeRange.Start = DEFAULT_START_TIME
		body.TimeRange.End = DEFAULT_END_TIME
	}
	return util.Request{
		Type: util.EditQueue,
		Body: body,
	}
}

func BasicClient() {
	for {
		t := askRequestType()
		var r util.Request
		switch t {
		case util.AddDownload:
			r = askAddDL()
		case util.PauseDownload,
			util.ResumeDownload,
			util.RetryDownload,
			util.StartDownload,
			util.CancelDownload,
			util.DeleteDownload:
			r = askModDL(t)
		//
		case util.AddQueue:
			r = askAddQueue()
		case util.DeleteQueue:
			r = askDelQueue()
		case util.EditQueue:
			r = askEditQueue()
		//
		case util.GetDownloads:
			r = util.Request{Type: util.GetDownloads}
		case util.GetQueues:
			r = util.Request{Type: util.GetQueues}
		default:
			fmt.Println("bad input. quitting")
			return
		}
		controller.SendAndPrint(r)

	}
}
