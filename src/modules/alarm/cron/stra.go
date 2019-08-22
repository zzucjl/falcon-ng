package cron

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/json-iterator/go"
	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/cache"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/config"
)

func SyncStraLoop() {
	interval := config.GetCfgYml().Interval
	for {
		SyncStra()
		time.Sleep(time.Second * time.Duration(interval))
	}
}

func SyncStra() error {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	servers := config.GetCfgYml().API.Portal.Server
	var err error
	for i := range servers {
		url := servers[i]
		if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
			url = "http://" + url
		}

		url = fmt.Sprintf("%s/api/portal/stras", url)
		resp, err := client.Get(url)
		if err != nil {
			logger.Errorf("sync stra failed, url: %s, err: %v", url, err)
			continue
		}

		if resp.Body != nil {
			defer resp.Body.Close()
			response, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Errorf("read response body failed, err: %v", err)
				continue
			}

			var dat dataobj.StraData
			if err = json.Unmarshal(response, &dat); err != nil {
				logger.Errorf("unmarshal response failed, response: %+v, err: %v", response, err)
				continue
			}
			if dat.Err != "" {
				logger.Errorf("response err: %s", dat.Err)
				continue
			}

			straMap := make(map[int64]*dataobj.Stra)
			for i := 0; i < len(dat.Dat); i++ {
				straMap[dat.Dat[i].ID] = &dat.Dat[i]
			}

			cache.StraCache.SetAll(straMap)
		}
		break
	}

	return err
}
