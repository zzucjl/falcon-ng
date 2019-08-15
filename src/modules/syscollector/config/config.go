package config

import (
	"fmt"
	"sync"

	"github.com/toolkits/pkg/file"
)

type ConfYaml struct {
	Debug            bool                `yaml:"debug"`
	NtpServers       []string            `yaml:"ntpServers"`
	PortPath         string              `yaml:"portPath"`
	ProcPath         string              `yaml:"procPath"`
	Plugin           string              `yaml:"plugin"`
	CollectAddr      string              `yaml:"collectAddr"`
	Endpoint         EndpointSection     `yaml:"endpoint"`
	Logger           LoggerSection       `yaml:"logger"`
	Transfer         TransferSection     `yaml:"transfer"`
	HTTP             HTTPSection         `yaml:"http"`
	Collector        CollectorSection    `yaml:"collector"`
	IgnoreMetricsLst []string            `yaml:"ignoreMetrics"`
	IgnoreMetricsMap map[string]struct{} `yaml:"-"`
}

type EndpointSection struct {
	Specify string `yaml:"specify"`
	Shell   string `yaml:"shell"`
}

type TransferSection struct {
	Enabled  string   `yaml:"enabled"`
	Addr     []string `yaml:"addr"`
	Interval int      `yaml:"interval"`
	Timeout  int      `yaml:"timeout"`
}

type LoggerSection struct {
	Level     string `yaml:"level"`
	Dir       string `yaml:"dir"`
	Rotatenum int    `yaml:"rotatenum"`
	Rotatemb  uint64 `yaml:"rotatemb"`
}

type HTTPSection struct {
	Listen string `yaml:"listen"`
}

type CollectorSection struct {
	IfacePrefix       []string `yaml:"ifacePrefix"`
	MountPoint        []string `yaml:"mountPoint"`
	MountIgnorePrefix []string `yaml:"mountIgnorePrefix"`
	SyncCollect       bool     `yaml:"syncCollect"`
	Addrs             []string `yaml:"addrs"`
	Timeout           int      `yaml:"timeout"`
	Interval          int      `yaml:"interval"`
}

var (
	Config   *ConfYaml
	lock     = new(sync.RWMutex)
	Hostname string
	Cwd      string
)

// Get configuration file
func Get() *ConfYaml {
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

	l := len(c.IgnoreMetricsLst)
	m := make(map[string]struct{}, l)
	for i := 0; i < l; i++ {
		m[c.IgnoreMetricsLst[i]] = struct{}{}
	}

	c.IgnoreMetricsMap = m

	Config = &c

	return nil
}
