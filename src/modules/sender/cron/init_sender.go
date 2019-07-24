package cron

import "github.com/open-falcon/falcon-ng/src/modules/sender/config"

var (
	IMWorkerChan    chan int
	SmsWorkerChan   chan int
	VoiceWorkerChan chan int
	MailWorkerChan  chan int
)

func InitSenderWorker() {
	cfg := config.GetCfgYml()
	IMWorkerChan = make(chan int, cfg.Worker.IM)
	SmsWorkerChan = make(chan int, cfg.Worker.Sms)
	VoiceWorkerChan = make(chan int, cfg.Worker.Voice)
	MailWorkerChan = make(chan int, cfg.Worker.Mail)
}
