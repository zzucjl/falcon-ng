package funcs

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"
)

func Push(v interface{}) {
	count := len(config.Config.Transfer.Addr)
	retry := 0
	for {
		for _, i := range rand.Perm(count) {
			addr := config.Config.Transfer.Addr[i]

			url := "http://" + addr + "/api/transfer/push"
			err := push(url, v)
			if err == nil {
				return
			}
			logger.Warning("push err:", url, err)
		}

		time.Sleep(time.Second)

		retry += 1
		if retry == 10 {
			break
		}
	}
}

func push(url string, v interface{}) error {
	logger.Debugf("item: %v", v)
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}

	bf := bytes.NewBuffer(bs)

	client := http.Client{
		Timeout: time.Duration(7) * time.Second,
	}

	_, err = client.Post(url, "application/json", bf)
	return err
}
