package controller

import(
	"github.com/placeholder14032/download-manager/internal/download"
	"github.com/placeholder14032/download-manager/internal/queue"
)
// make sure you check state in each condition for error ...
func AddDownload(queue Queue, url string) error {

}

func PauseDownload(queue Queue, id int64) error {

}

func ResumeDownload(queue Queue, id int64) error {

}

func CancelDownload(queue Queue, id int64) error {

}

func RetryDownload(queue Queue, id int64) error {

}