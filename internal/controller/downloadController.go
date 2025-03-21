package controller

import (
	"github.com/placeholder14032/download-manager/internal/util"
)

func AddDownload(url string, qid int64, fileName string) error {
	req := util.Request{
		Type: util.AddDownload,
		Body: util.BodyAddDownload{
			URL: url,
			QueueID: qid,
			FileName: fileName,
		},
	}
	resp := SendReq(req)
	return returnResp(resp)
}

func ModDownload(t util.RequestType, id int64) error {
	req := util.Request{
		Type: t,
		Body: util.BodyModDownload{
			ID: id,
		},
	}
	resp := SendReq(req)
	return returnResp(resp)
}

func GetAllDownloads() []util.DownloadBody {
	req := util.Request{
		Type: util.GetDownloads,
	}
	resp := SendReq(req)
	body, ok := resp.Body.(util.StaticDownloadList)
	if !ok {
		return []util.DownloadBody{}
	}
	return body.Downloads
}

