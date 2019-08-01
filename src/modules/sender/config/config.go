package config

import (
	"fmt"
	"sync"

	"github.com/toolkits/pkg/file"
)

type CfgYml struct {
	HTTP    HttpSection         `yaml:"http"`
	Logger  LoggerSection       `yaml:"logger"`
	Send    SendSection         `yaml:"send"`
	Worker  WorkerSection       `yaml:"worker"`
	Api     ApiSection          `yaml:"api"`
	Smtp    SmtpSection         `yaml:"smtp"`
	Queue   QueueSection        `yaml:"queue"`
	Redis   RedisSection        `yaml:"redis"`
	Auths   []string            `yaml:"auths"`
	AuthMap map[string]struct{} `yaml:"-"`
}

type RedisSection struct {
	Addr    string         `yaml:"addr"`
	Pass    string         `yaml:"pass"`
	Idle    int            `yaml:"idle"`
	Timeout TimeoutSection `yaml:"timeout"`
}

type TimeoutSection struct {
	Conn  int `yaml:"conn"`
	Read  int `yaml:"read"`
	Write int `yaml:"write"`
}

type QueueSection struct {
	IM    string `yaml:"im"`
	Sms   string `yaml:"sms"`
	Mail  string `yaml:"mail"`
	Voice string `yaml:"voice"`
}

type SmtpSection struct {
	FromName   string `yaml:"from_name"`
	FromMail   string `yaml:"from_mail"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	ServerHost string `yaml:"server_host"`
	ServerPort int    `yaml:"server_port"`
	UseSSL     bool   `yaml:"use_ssl"`
	StartTLS   bool   `yaml:"start_tls"`
}

type WorkerSection struct {
	IM    int `yaml:"im"`
	Sms   int `yaml:"sms"`
	Mail  int `yaml:"mail"`
	Voice int `yaml:"voice"`
}

type ApiSection struct {
	IM    string `yaml:"im"`
	Sms   string `yaml:"sms"`
	Mail  string `yaml:"mail"`
	Voice string `yaml:"voice"`
}

type SendSection struct {
	IM    string `yaml:"im"`
	Sms   string `yaml:"sms"`
	Mail  string `yaml:"mail"`
	Voice string `yaml:"voice"`
}

type LoggerSection struct {
	Dir       string `yaml:"dir"`
	Level     string `yaml:"level"`
	KeepHours uint   `yaml:"keepHours"`
}

type HttpSection struct {
	Listen string `yaml:"listen"`
}

var (
	cfgYml *CfgYml
	lock   = new(sync.RWMutex)
)

func GetCfgYml() *CfgYml {
	lock.RLock()
	defer lock.RUnlock()
	return cfgYml
}

func ParseCfg(ymlfile string) error {
	var c CfgYml
	err := file.ReadYaml(ymlfile, &c)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", ymlfile, err)
	}

	lock.Lock()
	defer lock.Unlock()

	c.AuthMap = make(map[string]struct{})
	for i := 0; i < len(c.Auths); i++ {
		c.AuthMap[c.Auths[i]] = struct{}{}
	}
	cfgYml = &c

	return nil
}
