package main

import (
	"github.com/placeholder14032/download-manager/internal/manager"
	"github.com/placeholder14032/download-manager/internal/util"
	"github.com/placeholder14032/download-manager/ui"
)
var reqs = make(chan util.Request)

func main() {
	var reqs = make(chan util.Request)
	var resps = make(chan util.Response)
	var manager = manager.Manager{}
	go manager.Start()
	go ui.Main(reqs, resps)
}
