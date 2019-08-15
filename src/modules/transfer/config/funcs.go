package config

import (
	"log"

	"github.com/toolkits/pkg/logger"
)

func InitLogger() {
	c := Config.Logger

	lb, err := logger.NewFileBackend(c.Dir)
	if err != nil {
		log.Fatalln(err)
	}

	logger.SetLogging(c.Level, lb)
	lb.Rotate(c.Rotatenum, 1024*1024*c.Rotatemb)
}
