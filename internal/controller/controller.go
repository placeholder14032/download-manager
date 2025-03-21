package controller

import (
	"fmt"

	"github.com/placeholder14032/download-manager/internal/util"
)

var Req chan util.Request
var Resp chan util.Response

func SetChannels(req chan util.Request, resp chan util.Response) {
	Req = req
	Resp = resp
}

func SendReq(r util.Request) util.Response {
	Req <- r
	return <- Resp
}

func SendAndPrint(r util.Request) {
	resp := SendReq(r)
	fmt.Println()
	fmt.Printf("%+v\n", resp)
}

func Close(){

}

