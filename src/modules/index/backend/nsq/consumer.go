package nsq

import (
	"encoding/json"
	"os"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/index/cache"
	. "github.com/open-falcon/falcon-ng/src/modules/index/config"

	nsq "github.com/bitly/go-nsq"
	"github.com/toolkits/pkg/concurrent/semaphore"
	"github.com/toolkits/pkg/logger"
)

type NSQWorker struct{}

var nsemaPush *semaphore.Semaphore

func StartNsqWorker() {
	nsqConf := Config.NSQ

	nsemaPush = semaphore.NewSemaphore(Config.BuildWorker)
	//g.NSQConsumerCnt = 0

	conf := nsq.NewConfig()
	conf.MaxInFlight = nsqConf.Worker

	// for full topic
	{
		consumer, err := nsq.NewConsumer(nsqConf.FullTopic, nsqConf.Chan, conf)
		if nil != err {
			logger.Fatal("create nsq worker failed : %v", err)
			os.Exit(1)
		}

		consumer.AddConcurrentHandlers(&NSQWorker{}, int(nsqConf.Worker))
		err = consumer.ConnectToNSQLookupds(nsqConf.Lookupds)
		if err != nil {
			logger.Fatal("nsq worker connect failed : %v", err)
			os.Exit(1)
		}
	}
	// for incr topic
	{
		consumer, err := nsq.NewConsumer(nsqConf.IncrTopic, nsqConf.Chan, conf)
		if nil != err {
			logger.Fatal("create nsq worker failed : %v", err)
			os.Exit(1)
		}

		consumer.AddConcurrentHandlers(&NSQWorker{}, int(nsqConf.Worker))
		err = consumer.ConnectToNSQLookupds(nsqConf.Lookupds)
		if err != nil {
			logger.Fatal("nsq worker connect failed : %v", err)
			os.Exit(1)
		}
	}
}

func (this *NSQWorker) HandleMessage(message *nsq.Message) error {
	//atomic.AddInt64(&g.NSQConsumerCnt, 1)
	var indexList []dataobj.IndexModel

	err := json.Unmarshal(message.Body, &indexList)
	if err != nil {
		logger.Errorf("nsq message unmarshal error : %v\n", err)
	}

	for _, item := range indexList {
		itemByte, _ := json.Marshal(item)
		itemCopy := dataobj.IndexModel{}
		json.Unmarshal(itemByte, &itemCopy)

		nsemaPush.Acquire()
		go func(item dataobj.IndexModel) {
			defer nsemaPush.Release()

			logger.Debugf("<index %v", item)
			if _, exists := DEFAULT_METRIC[item.Metric]; exists {
				return
			}
			err := cache.EndpointDBObj.Push(item)
			if err != nil {
				logger.Errorf("nsq message push failed : %v", err)
			}
		}(itemCopy)
	}

	return nil
}
