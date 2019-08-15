package config

import (
	"fmt"
	"sync"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"
)

type loggerSection struct {
	Level     string `yaml:"level"`
	Dir       string `yaml:"dir"`
	Rotatenum int    `yaml:"rotatenum"`
	Rotatemb  uint64 `yaml:"rotatemb"`
}

type httpSection struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
}

type strategySection struct {
	SyncCollect    bool     `yaml:"sync_collect"`
	ConfigAddrs    []string `yaml:"config_addrs"`
	Timeout        int      `yaml:"timeout"`
	UpdateDuration int      `yaml:"update_duration"`
	DefaultDegree  int      `yaml:"default_degree"`
	FilePath       string   `yaml:"file_path"`
}

type workerSection struct {
	WorkerNum    int    `yaml:"worker_num"`
	QueueSize    int    `yaml:"queue_size"`
	PushInterval int    `yaml:"push_interval"`
	PushURL      string `yaml:"push_url"`
	WaitPush     int    `yaml:"wait_push"`
}

type hostnameSection struct {
	Specify string `yaml:"specify"`
	Shell   string `yaml:"shell"`
}

type ConfYaml struct {
	Logger     loggerSection   `yaml:"logger"`
	Http       httpSection     `yaml:"http"`
	Strategy   strategySection `yaml:"strategy"`
	Worker     workerSection   `yaml:"worker"`
	Hostname   hostnameSection `yaml:"hostname"`
	MaxCPURate float64         `yaml:"max_cpu_rate"`
	MaxMemRate float64         `yaml:"max_mem_rate"`
}

var Config *ConfYaml
var Hostname string
var lock = new(sync.RWMutex)

func Parse(conf string) error {
	var c ConfYaml
	err := file.ReadYaml(conf, &c)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", conf, err)
	}

	lock.Lock()
	defer lock.Unlock()
	Config = &c

	logger.Infof("config file content : %v", Config)

	return err
}
