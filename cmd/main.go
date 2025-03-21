package main

import (
	"github.com/placeholder14032/download-manager/internal/manager"
	"github.com/placeholder14032/download-manager/internal/util"
	"github.com/placeholder14032/download-manager/ui"
)

func main() {
	var reqs = make(chan util.Request)
	var resps = make(chan util.Response)
	var manager = manager.Manager{}
	go manager.Start(reqs, resps)
	ui.Main()
}
