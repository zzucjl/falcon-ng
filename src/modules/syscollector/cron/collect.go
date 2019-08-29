package cron

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"

	"github.com/parnurzeal/gorequest"
)

type CollectResp struct {
	Dat model.Collect `json:"dat"`
	Err string        `json:"err"`
}

func GetCollects() {
	if !config.Config.Collector.SyncCollect {
		return
	}

	detect()
	go loopDetect()
}

func loopDetect() {
	t1 := time.NewTicker(time.Duration(config.Config.Collector.Interval) * time.Second)
	for {
		<-t1.C
		detect()
	}
}

func detect() {
	c, err := GetCollectsRetry()
	if err != nil {
		logger.Errorf("get collect err:%v", err)
		return
	}
	config.Collect.Update(&c)
}

func GetCollectsRetry() (model.Collect, error) {
	count := len(config.Config.Collector.Addrs)
	var resp CollectResp
	var err error
	for i := 0; i < count; i++ {
		resp, err = getCollects()
		if err == nil {
			if resp.Err != "" {
				err = fmt.Errorf(resp.Err)
				logger.Warningf("get collect err:%v", err)
				continue
			}
			return resp.Dat, err
		}
	}

	return resp.Dat, err
}

func getCollects() (CollectResp, error) {
	var res CollectResp
	var err error
	i := rand.Intn(len(config.Config.Collector.Addrs))
	addr := config.Config.Collector.Addrs[i]

	url := fmt.Sprintf("http://%s/api/portal/collects/%s", addr, config.Hostname)

	_, _, errs := gorequest.New().Timeout(time.Duration(config.Config.Collector.Timeout) * time.Second).Get(url).EndStruct(&res)
	if len(errs) != 0 {
		err = fmt.Errorf("get collects from remote failed, error:%v", errs)
	}

	return res, err
}
