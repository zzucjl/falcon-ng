package cron

import (
	"path"
	"strings"
	"time"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/net/httplib"
	"github.com/toolkits/pkg/sys"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/sender/config"
	"github.com/open-falcon/falcon-ng/src/modules/sender/redi"
)

type IM struct {
	Tos     []string `json:"tos"`
	Content string   `json:"content"`
}

func ConsumeIM() {
	for {
		imList := redi.Pop(1, "im")
		if len(imList) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendIMList(imList)
	}
}

func SendIMList(imList []*dataobj.Notify) {
	for _, im := range imList {
		IMWorkerChan <- 1
		go SendIM(im)
	}
}

func SendIM(im *dataobj.Notify) {
	defer func() {
		<-IMWorkerChan
	}()

	cfg := config.GetCfgYml()
	switch cfg.Send.IM {
	case "api":
		sendIMByApi(im)
	case "shell":
		sendIMByShell(im)
	default:
		logger.Errorf("not support %s to send im, im: %v", cfg.Send.Mail, *im)
	}
}

func sendIMByApi(im *dataobj.Notify) {
	cfg := config.GetCfgYml()
	url := cfg.Api.IM
	_, err := httplib.PostJSON(url, 5, IM{
		Tos:     im.Tos,
		Content: im.Content,
	}, map[string]string{})
	logger.Infof("sendIM use api, tos: %v, content: %v, err: %v", im.Tos, im.Content, err)
	return
}

func sendIMByShell(im *dataobj.Notify) {
	im_shell := path.Join(file.SelfDir(), "script", "send_im")
	if !file.IsExist(im_shell) {
		logger.Errorf("%s not found", im_shell)
		return
	}

	output, err, isTimeout := sys.CmdRunT(time.Second*10, im_shell, strings.Join(im.Tos, ","), im.Content)
	logger.Infof("sendIM use shell, tos: %v, content: %v, output: %v, err: %v, isTimeout: %v", im.Tos, im.Content, output, err, isTimeout)
}
