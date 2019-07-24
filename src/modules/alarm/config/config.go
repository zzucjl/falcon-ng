package config

import (
	"fmt"
	"sync"

	"github.com/toolkits/pkg/file"
)

const (
	RECOVERY = "recovery"
	ALERT    = "alert"
)

var (
	EventTypeMap = map[string]string{RECOVERY: "恢复", ALERT: "报警"}
)

type CfgYml struct {
	Interval int                 `yaml:"interval"`
	HTTP     HttpSection         `yaml:"http"`
	Logger   LoggerSection       `yaml:"logger"`
	Redis    RedisSection        `yaml:"redis"`
	Queue    QueueSection        `yaml:"queue"`
	Notify   map[string][]string `yaml:"notify"`
	API      APISection          `yaml:"api"`
	Link     LinkSection         `yaml:"link"`
	Cleaner  CleanerSection      `yaml:"cleaner"`
	Merge    MergeSection        `yaml:"merge"`
}

type MergeSection struct {
	Hash     string `yaml:"hash"`
	Max      int    `yaml:"max"`
	Interval int    `yaml:"interval"`
}

type CleanerSection struct {
	Days  int `yaml:"days"`
	Batch int `yaml:"batch"`
}

type LinkSection struct {
	Stra  string `yaml:"stra"`
	Event string `yaml:"event"`
	Claim string `yaml:"claim"`
}

type APISection struct {
	Sender ServerSection `yaml:"sender"`
	Portal ServerSection `yaml:"portal"`
}

type ServerSection struct {
	Server []string `yaml:"server"`
	Auth   string   `yaml:"auth"`
}

type QueueSection struct {
	High     []interface{} `yaml:"high"`
	Low      []interface{} `yaml:"low"`
	Callback string        `yaml:"callback"`
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
	cfgYml = &c

	return nil
}
