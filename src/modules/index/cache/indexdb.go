package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/open-falcon/falcon-ng/src/modules/index/config"

	"github.com/toolkits/pkg/concurrent/semaphore"
	"github.com/toolkits/pkg/logger"
)

var EndpointDBObj *EndpointMetricsStruct

const PERMANENCE_DIR = "./.data/"

var semaPermanence = semaphore.NewSemaphore(1)

func InitDB() {
	EndpointDBObj = &EndpointMetricsStruct{Metrics: make(map[string]*MetricsStruct, 0)}
}

func RebuildFromDisk() {
	permanDir := fmt.Sprintf("%s%s", PERMANENCE_DIR, "db")
	logger.Info("Try to Rebuild index from Disk.")
	if _, err := os.Stat(PERMANENCE_DIR); os.IsNotExist(err) {
		logger.Info("Permanence_dir not exists.")
		return
	}

	//遍历目录
	files, err := ioutil.ReadDir(permanDir)
	if err != nil {
		logger.Errorf("read dir error, [reason:%s]\n", err.Error())
		return
	}
	logger.Infof("There're [%d] ns need rebuild", len(files))

	limit := 1
	if Config.RebuildWorker > 0 {
		limit = Config.RebuildWorker
	}

	concurrency := semaphore.NewSemaphore(limit)

	for _, fileObj := range files {
		if fileObj.IsDir() {
			continue
		}
		endpoint := fileObj.Name()

		concurrency.Acquire()
		go func(endpoint string) {
			defer concurrency.Release()

			body, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", permanDir, endpoint))
			if err != nil {
				logger.Errorf("read file error, [endpoint:%s][reason:%v]", endpoint, err)
				return
			}

			metrics := new(MetricsStruct)

			err = json.Unmarshal(body, metrics)
			if err != nil {
				logger.Errorf("json unmarshal failed, [endpoint:%s][reason:%v]", endpoint, err)
				return
			}

			EndpointDBObj.Lock()
			EndpointDBObj.Metrics[endpoint] = metrics
			EndpointDBObj.Unlock()
		}(endpoint)

	}
	logger.Infof("rebuild from disk , [%d%%] complete\n", 100)
}
