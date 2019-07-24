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

type Sms struct {
	Tos     []string `json:"tos"`
	Content string   `json:"content"`
}

func ConsumeSms() {
	for {
		smsList := redi.Pop(1, "sms")
		if len(smsList) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendSmsList(smsList)
	}
}

func SendSmsList(smsList []*dataobj.Notify) {
	for _, sms := range smsList {
		SmsWorkerChan <- 1
		go SendSms(sms)
	}
}

func SendSms(sms *dataobj.Notify) {
	defer func() {
		<-SmsWorkerChan
	}()

	cfg := config.GetCfgYml()

	switch cfg.Send.Sms {
	case "api":
		sendSmsByApi(sms)
	case "shell":
		sendSmsByShell(sms)
	default:
		logger.Errorf("not support %s to send sms, sms: %v", cfg.Send.Mail, *sms)
	}

}

func sendSmsByApi(sms *dataobj.Notify) {
	cfg := config.GetCfgYml()
	url := cfg.Api.Sms

	_, err := httplib.PostJSON(url, 5, Sms{
		Tos:     sms.Tos,
		Content: sms.Content,
	}, map[string]string{})
	logger.Infof("sendSms use api, tos: %v, content: %v, err: %v", sms.Tos, sms.Content, err)
}

func sendSmsByShell(sms *dataobj.Notify) {
	sms_shell := path.Join(file.SelfDir(), "script", "send_sms")
	if !file.IsExist(sms_shell) {
		logger.Errorf("%s not found", sms_shell)
		return
	}
	output, err, isTimeout := sys.CmdRunT(time.Second*10, sms_shell, strings.Join(sms.Tos, ","), sms.Content)
	logger.Infof("sendSms use shell, tos: %v, content: %v, output:%v, err: %v, isTimeout: %v", sms.Tos, sms.Content, output, err, isTimeout)
}
