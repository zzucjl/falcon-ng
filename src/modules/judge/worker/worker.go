package worker

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/entity"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish"
	filep "github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish/file"
	nsqp "github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish/nsq"
	redisp "github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish/redis"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/buffer"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/query"
)

var (
	stg      storage.Storage        // 全局 唯一storage对象
	pub      publish.EventPublisher // 全局 唯一publisher对象
	identity string                 // 全局, 分片信息

	_strategy     = make(map[int64]*entity.StrategyEntity) // 全局 策略实体列表
	_strategyLock = &sync.RWMutex{}                        // 策略实体列表加锁
)

// Start 全局worker
func Start() {
	// 初始化配置文件
	opts := InitOptions()

	ident, err := GetIdentity(opts.Identity)
	if err != nil {
		log.Fatalln("[F] cannot get identity:", err)
	}

	log.Printf("[I] identity -> %s", ident)
	identity = ident

	// 初始化日志
	InitLog(opts.Log)

	// 初始化存储组件
	qm, err := query.NewSeriesQueryManager(opts.Query)
	if err != nil {
		log.Fatalln("[F] init transfer/index failed:", err)
	}
	stg = buffer.NewStorageBuffer(opts.Storage, qm)

	// 初始化publisher组件
	switch opts.Publisher.Type {
	case "nsq":
		pub, err = nsqp.NewNsqPublisher(opts.Publisher.Nsq)
	case "redis":
		pub, err = redisp.NewRedisPublisher(opts.Publisher.Redis)
	case "file":
		pub, err = filep.NewFilePublisher(opts.Publisher.File)
	default:
		err = errors.New("unknown publish type")
	}
	if err != nil {
		log.Fatalln("[F] init publisher failed:", err)
	}

	// 初始化nodata driver
	_nodataDrivers = make(map[int64]map[string]entity.AlertDriverEntity)
	_options = opts.Strategy

	// 策略初始化
	err = strategyManageLoop(true)
	if err != nil {
		log.Fatalln("[F] init strategy failed:", err)
	}

	// 进入循环
	go func() {
		for {
			select {
			case <-time.After(time.Duration(_options.UpdateInterval) * time.Millisecond):
				err := strategyManageLoop()
				if err != nil {
					logger.Warningf(0, "strategyManageLoop failed, error:%v", err)
				}
			}
		}
	}()
}

func Stop() {
	_strategyLock.Lock()
	for _, se := range _strategy {
		se.Stop()
	}
	_strategyLock.Unlock()

	pub.Close()
}

type Summary struct {
	StrategyCount int `json:"strategy_count"`
	SeriesCount   int `json:"series_count"`
}

func GetWorkerSummary() Summary {
	var count, count2 int
	_strategyLock.RLock()
	count = len(_strategy)
	for _, se := range _strategy {
		count2 += se.SeriesCount()
	}
	_strategyLock.RUnlock()

	return Summary{
		StrategyCount: count,
		SeriesCount:   count2,
	}
}

func GetStrategySummary(ID int64) (*entity.StrategySummary, bool) {
	_strategyLock.RLock()
	se, found := _strategy[ID]
	_strategyLock.RUnlock()
	if !found {
		return nil, false
	}
	return se.Summary(), true
}

func strategyManageLoop(first ...bool) error {
	ss, err := GetStrategyFromRemote(_options)
	if err != nil {
		logger.Warningf(0, "GetStrategyFromRemote failed:%v", err)
		return err
	}

	// 程序第一次启动不执行preHandler
	if len(first) == 0 || !first[0] {
		PreHandlerNodata(ss, time.Now())
	}

	// 更新 && 删除 && 新增
	// TODO: 清理logger中无效的ID列表
	var (
		updated int
		newest  = make(map[int64]struct{})
		deletes = make(map[int64]*entity.StrategyEntity)
		adds    = make(map[int64]*entity.StrategyEntity)
	)
	_strategyLock.RLock()
	for i := range ss {
		newest[ss[i].ID] = struct{}{}
		newEntity := entity.NewStrategyEntity(ss[i], stg, pub)
		if newEntity == nil {
			logger.Warning(ss[i].ID, "generate strategyEntity failed")
			continue
		}

		if old, found := _strategy[ss[i].ID]; found {
			updated++
			old.SetCache(newEntity)
			continue
		}
		adds[ss[i].ID] = newEntity
	}
	for i, se := range _strategy {
		if _, found := newest[i]; !found {
			deletes[i] = se
		}
	}
	_strategyLock.RUnlock()

	if len(adds) > 0 {
		_strategyLock.Lock()
		for ID, se := range adds {
			se.Start(_options.IndexInterval)
			_strategy[ID] = se
		}
		_strategyLock.Unlock()
	}

	if len(deletes) > 0 {
		_strategyLock.Lock()
		for ID, se := range deletes {
			if se.Status() == entity.ENTITY_STATUS_STOPPING {
				// do nothing
			}
			if se.Status() == entity.ENTITY_STATUS_STOPPED {
				delete(_strategy, ID)
				continue
			}
			se.Stop()
		}
		_strategyLock.Unlock()
	}
	logger.Infof(0, "strategyManageLoop finished, total:%d, add:%d, update:%d, delete:%d",
		len(ss), len(adds), updated, len(deletes))
	return nil
}
