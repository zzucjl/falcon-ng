package cron

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"
	email "github.com/toolkits/pkg/mail"
	"github.com/toolkits/pkg/net/httplib"
	"github.com/toolkits/pkg/sys"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/sender/config"
	"github.com/open-falcon/falcon-ng/src/modules/sender/redi"
)

type Mail struct {
	Tos     []string `json:"tos"`
	Subject string   `json:"subject"`
	Content string   `json:"content"`
}

func ConsumeMail() {
	for {
		mailList := redi.Pop(1, "mail")
		if len(mailList) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendMailList(mailList)
	}
}

func SendMailList(mailList []*dataobj.Notify) {
	for _, mail := range mailList {
		MailWorkerChan <- 1
		go SendMail(mail)
	}
}

func SendMail(mail *dataobj.Notify) {
	defer func() {
		<-MailWorkerChan
	}()

	cfg := config.GetCfgYml()

	switch cfg.Send.Mail {
	case "api":
		sendMailByApi(mail)
	case "smtp":
		sendMailBySmtp(mail)
	case "shell":
		sendMailByShell(mail)
	default:
		logger.Errorf("not support %s to send mail, mail: %v", cfg.Send.Mail, *mail)
	}
}

func sendMailByApi(mail *dataobj.Notify) {
	cfg := config.GetCfgYml()
	_, err := httplib.PostJSON(cfg.Api.Mail, 5, Mail{
		Tos:     mail.Tos,
		Subject: mail.Subject,
		Content: mail.Content,
	}, map[string]string{})
	logger.Infof("sendMail use api, tos: %v, subject: %v, content: %v, err: %v", mail.Tos, mail.Subject, mail.Content, err)
}

func sendMailBySmtp(mail *dataobj.Notify) {
	cfg := config.GetCfgYml()
	smtp := email.NewSMTP(
		cfg.Smtp.FromMail,
		cfg.Smtp.FromName,
		cfg.Smtp.Username,
		cfg.Smtp.Password,
		cfg.Smtp.ServerHost,
		cfg.Smtp.ServerPort,
		cfg.Smtp.UseSSL,
	)

	err := smtp.Send(email.Mail{
		Tos:     mail.Tos,
		Subject: mail.Subject,
		Content: mail.Content,
	})
	logger.Infof("sendMail use smtp, tos: %v, subject: %v, content: %v, err: %v", mail.Tos, mail.Subject, mail.Content, err)
}

func sendMailByShell(mail *dataobj.Notify) {
	mail_shell := path.Join(file.SelfDir(), "script", "send_mail")
	if !file.IsExist(mail_shell) {
		logger.Errorf("%s not found", mail_shell)
		return
	}

	fp := fmt.Sprintf("/tmp/n9e.mail.content.%d", time.Now().UnixNano())
	_, err := file.WriteString(fp, mail.Content)
	if err != nil {
		logger.Errorf("cannot write string to %s", fp)
		return
	}

	output, err, isTimeout := sys.CmdRunT(time.Second*10, mail_shell, strings.Join(mail.Tos, ","), mail.Subject, fp)
	logger.Infof("sendMail use shell, tos: %v, subject: %v, content: %v, output:%v, err: %v, isTimeout: %v", mail.Tos, mail.Subject, mail.Content, output, err, isTimeout)

	file.Unlink(fp)
}
