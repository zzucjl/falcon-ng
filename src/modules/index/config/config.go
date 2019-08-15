package config

import (
	"fmt"
	"sync"

	"github.com/toolkits/pkg/file"
)

type ConfYaml struct {
	Debug           bool          `yaml:"debug"`
	CacheDuration   int           `yaml:"cacheDuration"`
	CleanInterval   int           `yaml:"cleanInterval"`
	PersistInterval int           `yaml:"persistInterval"`
	RebuildWorker   int           `yaml:"rebuildWorker"`
	BuildWorker     int           `yaml:"buildWorker"`
	DefaultStep     int           `yaml:"defaultStep"`
	Logger          LoggerSection `yaml:"logger"`
	HTTP            HTTPSection   `yaml:"http"`
	RPC             RPCSection    `yaml:"rpc"`
	Limit           LimitSection  `yaml:"limit"`
	NSQ             NSQSection    `yaml:"nsq"`
	Tree            TreeSection   `yaml:"tree"`
}

type TreeSection struct {
	Timeout int      `yaml:"timeout"`
	Addrs   []string `yaml:"addrs"`
}

type NSQSection struct {
	Enabled   bool     `yaml:"enabled"`
	Lookupds  []string `yaml:"lookupds"`
	FullTopic string   `yaml:"fullTopic"`
	IncrTopic string   `yaml:"incrTopic"`
	Chan      string   `yaml:"chan"`
	Worker    int      `yaml:"worker"`
}

type LimitSection struct {
	UI                  int `yaml:"ui"`
	Match               int `yaml:"match"`
	CludeLogCounter     int `yaml:"cludeLogCounter"`
	Clude               int `yaml:"clude"`
	FullmatchLogCounter int `yaml:"fullmatchLogCounter"`
}

type LoggerSection struct {
	Level     string `yaml:"level"`
	Dir       string `yaml:"dir"`
	Rotatenum int    `yaml:"rotatenum"`
	Rotatemb  uint64 `yaml:"rotatemb"`
}

type HTTPSection struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
	Access  string `yaml:"access"`
}

type RPCSection struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

var (
	Config *ConfYaml
	lock   = new(sync.RWMutex)
)

func GetCfgYml() *ConfYaml {
	lock.RLock()
	defer lock.RUnlock()
	return Config
}

func Parse(conf string) error {
	var c ConfYaml
	err := file.ReadYaml(conf, &c)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", conf, err)
	}

	DEFAULT_STEP = c.DefaultStep

	lock.Lock()
	defer lock.Unlock()
	Config = &c

	return err
}
