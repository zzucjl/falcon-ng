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

type Voice struct {
	Tos     []string `json:"tos"`
	Content string   `json:"content"`
}

func ConsumeVoice() {
	for {
		voiceList := redi.Pop(1, "voice")
		if len(voiceList) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendVoiceList(voiceList)
	}
}

func SendVoiceList(voiceList []*dataobj.Notify) {
	for _, voice := range voiceList {
		VoiceWorkerChan <- 1
		go SendVoice(voice)
	}
}

func SendVoice(voice *dataobj.Notify) {
	defer func() {
		<-VoiceWorkerChan
	}()

	cfg := config.GetCfgYml()

	switch cfg.Send.Voice {
	case "api":
		sendVoiceByApi(voice)
	case "shell":
		sendVoiceByShell(voice)
	default:
		logger.Errorf("not support %s to send voice, voice: %v", cfg.Send.Mail, *voice)
	}

}

func sendVoiceByApi(voice *dataobj.Notify) {
	cfg := config.GetCfgYml()
	url := cfg.Api.Voice

	_, err := httplib.PostJSON(url, 5, Voice{
		Tos:     voice.Tos,
		Content: voice.Content,
	}, map[string]string{})
	logger.Infof("sendVoice use api, tos: %v, content: %v, err: %v", voice.Tos, voice.Content, err)
}

func sendVoiceByShell(voice *dataobj.Notify) {
	voice_shell := path.Join(file.SelfDir(), "script", "send_voice")
	if !file.IsExist(voice_shell) {
		logger.Errorf("%s not found", voice_shell)
		return
	}
	output, err, isTimeout := sys.CmdRunT(time.Second*10, voice_shell, strings.Join(voice.Tos, ","), voice.Content)
	logger.Infof("sendVoice use shell, tos: %v, content: %v, output:%v, err: %v, isTimeout: %v", voice.Tos, voice.Content, output, err, isTimeout)
}
