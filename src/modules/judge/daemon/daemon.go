package daemon

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/worker"
)

func Start() {
	worker.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-c:
		logger.Info(0, "stop signal caught, try to stop judge server")
		worker.Stop()
	}
	logger.Info(0, "judge server stopped succefully")
	logger.Close()
}
