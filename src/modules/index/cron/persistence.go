package cron

import (
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/modules/index/cache"
	. "github.com/open-falcon/falcon-ng/src/modules/index/config"
)

func StartPersist() {
	t1 := time.NewTicker(time.Duration(Config.PersistInterval) * time.Second)
	for {
		<-t1.C

		//start := time.Now()
		err := cache.EndpointDBObj.Persist("normal")
		if err != nil {
			logger.Error("Persist err:", err)
		}
		//logger.Infof("clean %+v, took %.2f ms\n", cleanRet, float64(time.Since(start).Nanoseconds())*1e-6)
	}
}
