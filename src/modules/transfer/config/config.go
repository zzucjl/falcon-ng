package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/toolkits/pkg/file"
)

type ConfYaml struct {
	Debug   bool          `yaml:"debug"`
	MinStep int           `yaml:"minStep"`
	Logger  LoggerSection `yaml:"logger"`
	HTTP    HTTPSection   `yaml:"http"`
	RPC     RPCSection    `yaml:"rpc"`
	Judge   JudgeSection  `yaml:"judge"`
	Tsdb    TsdbSection   `yaml:"tsdb"`
	Index   IndexSection  `yaml:"index"`
}

type IndexSection struct {
	Addrs   []string `yaml:"addrs"`
	Timeout int      `yaml:"timeout"`
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

type JudgeSection struct {
	Enabled     bool              `yaml:"enabled"`
	Batch       int               `yaml:"batch"`
	ConnTimeout int               `yaml:"connTimeout"`
	CallTimeout int               `yaml:"callTimeout"`
	WorkerNum   int               `yaml:"workerNum"`
	MaxConns    int               `yaml:"maxConns"`
	MaxIdle     int               `yaml:"maxIdle"`
	Replicas    int               `yaml:"replicas"`
	Cluster     map[string]string `yaml:"cluster"`
}

type TsdbSection struct {
	Enabled     bool                    `yaml:"enabled"`
	Batch       int                     `yaml:"batch"`
	ConnTimeout int                     `yaml:"connTimeout"`
	CallTimeout int                     `yaml:"callTimeout"`
	WorkerNum   int                     `yaml:"workerNum"`
	MaxConns    int                     `yaml:"maxConns"`
	MaxIdle     int                     `yaml:"maxIdle"`
	Replicas    int                     `yaml:"replicas"`
	Cluster     map[string]string       `yaml:"cluster"`
	ClusterList map[string]*ClusterNode `json:"clusterList"`
}

var (
	Config *ConfYaml
	lock   = new(sync.RWMutex)
)

// CLUSTER NODE
type ClusterNode struct {
	Addrs []string `json:"addrs"`
}

func NewClusterNode(addrs []string) *ClusterNode {
	return &ClusterNode{addrs}
}

// map["node"]="host1,host2" --> map["node"]=["host1", "host2"]
func formatClusterItems(cluster map[string]string) map[string]*ClusterNode {
	ret := make(map[string]*ClusterNode)
	for node, clusterStr := range cluster {
		items := strings.Split(clusterStr, ",")
		nitems := make([]string, 0)
		for _, item := range items {
			nitems = append(nitems, strings.TrimSpace(item))
		}
		ret[node] = NewClusterNode(nitems)
	}

	return ret
}

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
	c.Tsdb.ClusterList = formatClusterItems(c.Tsdb.Cluster)

	lock.Lock()
	defer lock.Unlock()
	Config = &c

	return err
}
