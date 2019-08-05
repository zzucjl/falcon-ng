package worker

import (
	"log"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
)

func InitLog(opts LoggerOption) {
	backend, err := logger.NewFileBackend(opts.Path)
	if err != nil {
		log.Fatalln("[F] InitLog failed:", err)
	}

	// 初始化日志库
	logger.SetLogging(opts.Level, backend)
	backend.SetRotateByHour(true)
	backend.SetKeepHours(uint(opts.KeepHours))
}
