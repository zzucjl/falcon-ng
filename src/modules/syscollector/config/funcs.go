package config

import (
	"log"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/sys"
)

func GetEndpoint() (string, error) {
	if Config.Endpoint.Specify != "" {
		return Config.Endpoint.Specify, nil
	}

	return sys.CmdOutTrim("bash", "-c", Config.Endpoint.Shell)
}

// InitLogger init logger toolkits
func InitLogger() {
	c := Config.Logger

	lb, err := logger.NewFileBackend(c.Dir)
	if err != nil {
		log.Fatalln(err)
	}

	logger.SetLogging(c.Level, lb)
	lb.Rotate(c.Rotatenum, 1024*1024*c.Rotatemb)
}
