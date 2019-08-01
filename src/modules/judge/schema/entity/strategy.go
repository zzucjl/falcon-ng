package entity

import (
	"encoding/json"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/query"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/series"

	nsema "github.com/toolkits/concurrent/semaphore"
)

const (
	ENTITY_STATUS_EMPTY = iota
	ENTITY_STATUS_STARTING
	ENTITY_STATUS_RUNNING
	ENTITY_STATUS_WAITING
	ENTITY_STATUS_STOPPING
	ENTITY_STATUS_STOPPED
)

var (
	statusMap = map[int]string{
		ENTITY_STATUS_EMPTY:    "unknown",
		ENTITY_STATUS_STARTING: "starting",
		ENTITY_STATUS_RUNNING:  "running",
		ENTITY_STATUS_WAITING:  "waiting",
		ENTITY_STATUS_STOPPING: "stopping",
		ENTITY_STATUS_STOPPED:  "stopped",
	}
)

func NewStrategyEntity(stra *schema.Strategy,
	stg storage.Storage, pubs ...publish.EventPublisher) *StrategyEntity {
	if len(stra.Judgements) == 0 {
		logger.Warning(stra.ID, "new strategy failed: empty judgement")
		return nil
	}
	if len(stra.Endpoints) == 0 {
		logger.Warning(stra.ID, "new strategy failed: empty endpoint list")
		return nil
	}

	if stra.Operator != schema.LOGIC_OPERATOR_AND &&
		stra.Operator != schema.LOGIC_OPERATOR_OR {
		logger.Warningf(stra.ID, "new strategy failed: unknown operator:%s",
			stra.Operator)
		return nil
	}

	var pub publish.EventPublisher
	if len(pubs) > 0 {
		pub = pubs[0]
	}

	executions := make([]*ExecutionEntity, 0)
	for i := range stra.Judgements {
		metric := stra.Judgements[i].Metric

		ee, err := NewExecutionEntity(stra, stra.Judgements[i].Execution, metric)
		if err != nil {
			logger.Warningf(stra.ID, "new strategy failed: execution error %v", err)
			return nil
		}
		executions = append(executions, ee)
	}

	return &StrategyEntity{
		Strategy:   stra,
		stop:       make(chan struct{}, 1),
		cache:      nil,
		status:     ENTITY_STATUS_EMPTY,
		indexing:   false,
		interval:   10, // 默认设置interval为10s
		publisher:  pub,
		storage:    stg,
		Judgements: make(map[uint32]*JudgementEntity),
		Executions: executions,
	}
}

func (se *StrategyEntity) loop() {
	defer func() {
		if r := recover(); r != nil {
			se.status = ENTITY_STATUS_STOPPED

			var buf [8192]byte
			n := runtime.Stack(buf[:], false)
			logger.Errorf(0, "strategy run panic recover: -> sid:%d, %v\n%s", se.ID, r, buf[:n])
		}
	}()

	for {
		select {
		case <-se.stop:
			se.status = ENTITY_STATUS_STOPPED
			logger.Info(se.ID, "strategy run stopped")
			return
		case <-time.After(time.Duration(se.interval) * time.Second):
			// 更新策略配置, 如果有
			se.Update(false)

			// 执行判断 && 生成事件
			se.Run()

			// 执行完成, 等待下一个周期
			se.status = ENTITY_STATUS_WAITING
		}
	}

}

func (se *StrategyEntity) Status() int {
	return se.status
}

// SetCache 更新cache缓冲区
func (se *StrategyEntity) SetCache(newest *StrategyEntity) {
	if newest == nil {
		return
	}

	se.Lock()
	defer se.Unlock()

	if se.status == ENTITY_STATUS_STARTING ||
		se.status == ENTITY_STATUS_RUNNING ||
		se.status == ENTITY_STATUS_WAITING {
		if se.cache == nil && newest.Updated >= se.Updated {
			se.cache = newest
		}
		if se.cache != nil && newest.Updated >= se.cache.Updated {
			se.cache = newest
		}
	}
}

func (se *StrategyEntity) Start(indexInterval int) {
	se.indexInterval = indexInterval

	logger.Info(se.ID, "strategy entity started")
	se.status = ENTITY_STATUS_STARTING

	se.Update(true)

	go se.loop()

	go func() {
		for {
			select {
			case <-time.After(time.Duration(se.indexInterval) * time.Millisecond):
				se.Update(true)
				logger.Debug(se.ID, "strategy index update finished")
			}
		}
	}()
}

func (se *StrategyEntity) Stop() {
	se.status = ENTITY_STATUS_STOPPING
	se.stop <- struct{}{}
}

// CounterUnit 策略解析过程的中间变量
type CounterUnit struct {
	ID          uint32
	Granularity int
}

// Update
func (se *StrategyEntity) Update(index bool) {
	// index = false 更新strategy配置
	if !index {
		se.Lock()
		if se.cache != nil {
			se.Strategy = se.cache.Strategy
			se.Executions = se.cache.Executions

			if len(se.Judgements) > 0 {
				for _, je := range se.Judgements {
					je.Driver.SetThreshold(se.Strategy.Alert)
				}
			}
			se.cache = nil
		}
		se.Unlock()
		return
	}

	if se.indexing {
		logger.Debug(se.ID, "strategy index update exist: indexing")
		return
	}
	// index = true 根据index索引更新曲线配置
	se.indexing = true
	defer func() { se.indexing = false }()

	// 如果缓冲区不为空(代表存在更"新"的Strategy配置), 以最新的为准
	se.RLock()
	stra := se.Strategy
	if se.cache != nil {
		stra = se.cache.Strategy
	}
	se.RUnlock()

	var (
		bufferSize int
		span       []int
		once       bool = false

		all      = make(map[uint32]struct{})
		adds     = make(map[uint32]*JudgementEntity)
		updates  = make(map[uint32]*JudgementEntity)
		IDmap    = make(map[uint32]struct{}) // map[ID]struct{}
		interval = 0                         // 最小的调度间隔
	)

	for _, endpoint := range stra.Endpoints {
		counters := make(map[string]map[string]CounterUnit) // map[key]{map[key]unit}

		for i := range stra.Judgements {
			metric := stra.Judgements[i].Metric
			mkey := encodeMetricKey(metric, strconv.Itoa(i))
			// 确保 len(counters) 不会小于1
			if _, found := counters[mkey]; !found {
				cmap := make(map[string]CounterUnit)
				counters[mkey] = cmap
			}

			include, exclude := stra.Judgements[i].Xclude()
			result, err := se.storage.Index(
				query.NewIndexRequest(endpoint, metric, include, exclude),
			)
			if err != nil {
				logger.Warningf(se.ID, "strategy index warning: index error %v", err)
				continue
			}

			for j := range result {
				if result[j].Step == 0 {
					logger.Warningf(se.ID, "strategy index warning: zero step %v", result[j])
					continue
				}
				if !once {
					bufferSize, span = stra.MaxBufferSizeAndSpan(result[j].Step)
					once = true
				}
				// 生成series
				tags := series.CounterString2TagMap(result[j].Counter)

				ss, err := series.NewSeries(metric, tags, result[j].Step, result[j].Dstype)
				if err != nil {
					logger.Debugf(se.ID, "strategy index warning: new series failed %v", err)
					continue
				}
				// 生成ID并更新缓存信息
				ID := se.storage.GenerateAndSet(ss, bufferSize, span)

				// 不能使用 ss.Key(), 因为不能包含 metric字段
				ckey := result[j].Counter
				if _, found := counters[mkey][ckey]; !found {
					counters[mkey][ckey] = CounterUnit{
						ID:          ID,
						Granularity: result[j].Step,
					}
				}
			}
		}

		// counter字符串一致 && step一致 的曲线可以用于 与/或 报警
		mkeys := make([]string, 0)
		metricUnits := make([]map[string]CounterUnit, 0) // map[metric]unit
		for mkey, _ := range counters {
			mkeys = append(mkeys, mkey)
		}

		if len(mkeys) == 1 {
			metric := decodeMetricKey(mkeys[0])
			for _, unit := range counters[mkeys[0]] {
				metricUnits = append(metricUnits, map[string]CounterUnit{
					metric: CounterUnit{ID: unit.ID, Granularity: unit.Granularity},
				})
			}
		} else if len(mkeys) > 1 {
			// 多指标的与条件, 要求counter字符串(包括step值)完全一致
			pairs := make(map[string]CounterUnit) // map[key]unit
			mkey := mkeys[0]
			for key, unit := range counters[mkey] {
				pairs[mkey] = CounterUnit{ID: unit.ID, Granularity: unit.Granularity}
				for i := 1; i < len(mkeys); i++ {
					if _, found := counters[mkeys[i]]; !found {
						break
					}
					for ikey, iunit := range counters[mkeys[i]] {
						if ikey == key && iunit.Granularity == unit.Granularity {
							pairs[mkeys[i]] = CounterUnit{ID: iunit.ID, Granularity: iunit.Granularity}
							break
						}
					}
				}
				if len(pairs) == len(mkeys) {
					copyPairs := make(map[string]CounterUnit) // map[metric]unit
					for mkey, unit := range pairs {
						metric := decodeMetricKey(mkey)
						copyPairs[metric] = CounterUnit{ID: unit.ID, Granularity: unit.Granularity}
					}
					metricUnits = append(metricUnits, copyPairs)
				}
			}
		}
		if len(metricUnits) > 0 {
			se.RLock()
			for i := range metricUnits {
				mentitys := make([]*MetricEntity, 0)
				minterval := 0
				for metric, unit := range metricUnits[i] {
					minterval = unit.Granularity
					mentitys = append(mentitys, &MetricEntity{
						key:     metric,
						ID:      unit.ID,
						History: nil,
					})
					IDmap[unit.ID] = struct{}{} // 记录有效的ID列表
				}
				if interval == 0 {
					interval = minterval
				} else if interval > minterval {
					interval = minterval
				}
				je := NewJudgementEntity(stra, mentitys, minterval)
				if je != nil {
					all[je.ID()] = struct{}{}
					if _, found := se.Judgements[je.ID()]; !found {
						adds[je.ID()] = je
					} else {
						updates[je.ID()] = je
					}
				}
			}
			se.RUnlock()
		}
	}

	// 更新曲线列表
	se.Lock()
	if interval == 0 {
		logger.Warning(se.ID, "strategy index update warning: empty counters")
		for id, _ := range se.Judgements {
			delete(se.Judgements, id)
		}
		se.interval = 10
		se.concurrency = nil
		se.seriesCount = 0
		se.Unlock()
		return
	}

	// 清空已下线的曲线
	for id, je := range se.Judgements {
		if _, found := all[id]; !found {
			delete(se.Judgements, id)
		} else {
			// 更新旧曲线的配置
			if updates[id] != nil {
				je.Update(updates[id])
			}
		}
	}
	// 添加新的曲线
	for id, je := range adds {
		if je != nil {
			se.Judgements[id] = je
		}
	}

	se.interval = interval
	concurrency := getConcurrency(interval, len(se.Judgements))
	se.concurrency = nsema.NewSemaphore(concurrency)
	se.seriesCount = len(IDmap)

	se.Unlock()

	return
}

func (se *StrategyEntity) Run() {
	se.status = ENTITY_STATUS_RUNNING
	now := time.Now()
	logger.Debugf(se.ID, "strategy start to run")

	// se.Strategy字段的更新只会发生在 se.Run() 之前, 不会存在并行
	// 所有的更新先放到se.cache, 然后再由 se.Update() 更新到se.Strategy
	// 因此使用 se.Strategy 和  se.Executions 时 不需要加读锁
	executions := se.ExecutionsByTime(now)
	if len(executions) == 0 {
		logger.Debug(se.ID, "strategy had no effective executions")
		return
	}

	var (
		wg sync.WaitGroup
		c  chan struct{}
	)
	c = make(chan struct{}, 1)

	se.RLock()
	for _, judgement := range se.Judgements {
		wg.Add(1)
		se.concurrency.Acquire()

		go func(judgement *JudgementEntity, executions []*ExecutionEntity) {
			defer func() {
				if r := recover(); r != nil {
					var buf [8192]byte
					n := runtime.Stack(buf[:], false)
					logger.Errorf(0, "strategy run panic recover: -> sid:%d, %v\n%s", se.ID, r, buf[:n])
				}
			}()
			defer wg.Done()
			defer se.concurrency.Release()

			events := judgement.Run(se.storage, se.Strategy, executions, now)

			// 推送事件
			if len(events) > 0 {
				se.Publish(events)
			}
		}(judgement, executions)
	}
	se.RUnlock()

	// wg.Wait()超时, 主动退出
	go func() {
		wg.Wait()
		c <- struct{}{}
	}()
	select {
	case <-c:
	case <-time.After(time.Duration(se.interval) * time.Second):
		logger.Warningf(se.ID, "strategy run timetout, exit")
	}

	logger.Debugf(se.ID, "strategy finished, cost: %d ms", time.Now().Sub(now)/time.Millisecond)
}

func (se *StrategyEntity) Publish(events []*schema.Event) {
	if len(events) == 0 {
		return
	}
	if se.publisher == nil {
		logger.Error(se.ID, "strategy event push failed: publisher not set")
		return
	}

	for i := range events {
		err := se.publisher.Publish(events[i])
		if err != nil {
			logger.Warningf(se.ID, "strategy event push failed: %v", err)
			continue
		}
		// 特殊使用, 需要执行json打印出详细的信息
		if logger.GetSeverity(se.ID) == logger.DEBUG {
			bytes, _ := json.Marshal(events[i])
			logger.Debugf(se.ID, "strategy event push success: %s", string(bytes))
		}
	}
}

// 根据时刻信息 选择可执行的executions
func (se *StrategyEntity) ExecutionsByTime(now time.Time) []*ExecutionEntity {
	ret := make([]*ExecutionEntity, 0)
	if len(se.Executions) == 0 {
		return ret
	}
	day := weekday(now)
	minute := minute(now)
	for i := range se.Executions {
		if se.Executions[i].Effective(day, minute) {
			ret = append(ret, se.Executions[i])
		}
	}
	return ret
}

func (se *StrategyEntity) SeriesCount() int {
	se.RLock()
	defer se.RUnlock()

	return se.seriesCount
}

type StrategySummary struct {
	schema.Strategy
	Status      string `json:"status"`
	Interval    int    `json:"interval"`
	SeriesCount int    `json:"series_count"`
}

func (se *StrategyEntity) Summary() *StrategySummary {
	se.RLock()
	defer se.RUnlock()
	return &StrategySummary{
		Strategy:    *se.Strategy,
		Status:      statusMapping(se.status),
		Interval:    se.interval,
		SeriesCount: se.seriesCount,
	}
}

func statusMapping(status int) string {
	s, found := statusMap[status]
	if !found {
		return statusMap[ENTITY_STATUS_EMPTY]
	}
	return s
}

func encodeMetricKey(metric string, random string) string {
	return metric + "&" + random
}

func decodeMetricKey(key string) string {
	return strings.SplitN(key, "&", 2)[0]
}

func weekday(now time.Time) int {
	return int(now.Weekday())
}

func minute(now time.Time) int {
	return now.Hour()*60 + now.Minute()
}

func getConcurrency(interval int, sum int) int {
	c := sum / interval
	if c <= 1 {
		return 1
	}
	return c
}

/*
 * true 或 true = true
 * true 或 false = true
 * false 或 false = false
 * true 或 null = true
 * false 或 null = null
 * null 或 null = null

 * true 且 true = true
 * true 且 false = false
 * false 且 false = false
 * true 且 null = null
 * false 且 null = null
 * null 且 null = null
 */
func logicalOperate(operator string, a, b int) int {
	if a == schema.STATUS_EMPTY ||
		b == schema.STATUS_EMPTY {
		return schema.STATUS_EMPTY
	}
	if a == schema.STATUS_INIT &&
		b == schema.STATUS_INIT {
		// 不应该出现
		return schema.STATUS_EMPTY
	}
	if a == schema.STATUS_INIT {
		return b
	}
	if b == schema.STATUS_INIT {
		return a
	}
	if a == schema.STATUS_NULL {
		if operator == schema.LOGIC_OPERATOR_OR &&
			b == schema.STATUS_ALERT {
			return schema.STATUS_ALERT
		}
		return schema.STATUS_NULL
	}
	if b == schema.STATUS_NULL {
		if operator == schema.LOGIC_OPERATOR_OR &&
			a == schema.STATUS_ALERT {
			return schema.STATUS_ALERT
		}
		return schema.STATUS_NULL
	}
	if operator == schema.LOGIC_OPERATOR_AND {
		if a == schema.STATUS_ALERT &&
			b == schema.STATUS_ALERT {
			return schema.STATUS_ALERT
		}
		return schema.STATUS_RECOVER
	} else if operator == schema.LOGIC_OPERATOR_OR {
		if a == schema.STATUS_RECOVER &&
			b == schema.STATUS_RECOVER {
			return schema.STATUS_RECOVER
		}
		return schema.STATUS_ALERT
	}
	return schema.STATUS_EMPTY
}
