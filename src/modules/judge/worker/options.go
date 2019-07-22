package worker

import (
	"log"
	"path"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/buffer"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/query"

	"github.com/open-falcon/falcon-ng/src/system"
	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/sys"
)

type Options struct {
	Log       LoggerOption               `yaml:"log"`
	Storage   buffer.StorageBufferOption `yaml:"storage"`
	Query     query.SeriesQueryOption    `yaml:"query"`
	Publisher publish.PublisherOption    `yaml:"publisher"`
	Strategy  StrategyConfigOption       `yaml:"strategy"`
	Identity  IdentityOption             `yaml:"identity"`
}

func InitOptions() Options {
	dir := path.Join(system.Cwd, "etc")
	cfg := path.Join(dir, "judge.local.yml")
	if !file.IsExist(cfg) {
		cfg = path.Join(dir, "judge.yml")
	}

	var opts Options
	err := file.ReadYaml(cfg, &opts)
	if err != nil {
		log.Fatalf("cannot read yml[%s]: %v\n", opts, err)
	}

	return opts
}

func NewDefaultOptions(querys []string,
	index []string, config []string,
	redis []string) Options {
	return Options{
		Http:      ":8080",
		Log:       NewLoggerOption(),
		Publisher: publish.NewRedisPublisherOption(redis),
		Storage:   buffer.NewStorageBufferOption(),
		Query:     query.NewSeriesQueryOption(querys, index),
		Strategy:  NewStrategyConfigOption(config),
	}
}

var (
	defaultLogPath      = "./log"
	defaultLogLevel     = "INFO"
	defaultLogKeepHours = 12
)

type LoggerOption struct {
	Path      string `yaml:"path"`
	Level     string `yaml:"level"`
	KeepHours int    `yaml:"keepHours"`
}

func NewLoggerOption() LoggerOption {
	return LoggerOption{
		Path:      defaultLogPath,
		Level:     defaultLogLevel,
		KeepHours: defaultLogKeepHours,
	}
}

var (
	defaultStrategyConfigTimeout        = 5000
	defaultStrategyConfigUpdateInterval = 60000 // 1分钟
)

type StrategyConfigOption struct {
	Addrs          []string `yaml:"addrs"` // 形如 http://IP:port/url
	PartitionApi   string   `yaml:"partitionApi"`
	Timeout        int      `yaml:"timeout"`
	UpdateInterval int      `yaml:"updateInterval"`
	IndexInterval  int      `yaml:"indexInterval"`
}

func NewStrategyConfigOption(addrs []string) StrategyConfigOption {
	return StrategyConfigOption{
		Addrs:          addrs,
		PartitionApi:   "/api/stra/effective?ip=%s",
		Timeout:        defaultStrategyConfigTimeout,
		UpdateInterval: defaultStrategyConfigUpdateInterval,
		IndexInterval:  defaultStrategyConfigUpdateInterval,
	}
}

type IdentityOption struct {
	Specify string `yaml:"specify"`
	Shell   string `yaml:"shell"`
}

func GetIdentity(opts IdentityOption) (string, error) {
	if opts.Specify != "" {
		return opts.Specify, nil
	}

	return sys.CmdOutTrim("bash", "-c", opts.Shell)
}
