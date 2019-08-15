package cron

import (
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/index/cache"
	. "github.com/open-falcon/falcon-ng/src/modules/index/config"

	"github.com/toolkits/pkg/logger"
)

func StartCleaner() {
	t1 := time.NewTicker(time.Duration(Config.CleanInterval) * time.Second)
	for {
		<-t1.C

		start := time.Now()
		cache.EndpointDBObj.Clean(int64(Config.CacheDuration))
		logger.Infof("clean took %.2f ms\n", float64(time.Since(start).Nanoseconds())*1e-6)
	}
}
