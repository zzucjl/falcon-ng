package entity

import (
	"fmt"
	"sort"
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/buffer"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

var (
	defaultWindowSizeByInterval = 30 // 默认最多等待30个周期的数据延迟
)

type MetricEntitys []*MetricEntity

func (es MetricEntitys) Len() int           { return len(es) }
func (es MetricEntitys) Swap(i, j int)      { es[i], es[j] = es[j], es[i] }
func (es MetricEntitys) Less(i, j int) bool { return es[i].key < es[j].key }

// 统一向前推一个周期, 避免数据上报延时
func timeWindow(now time.Time, granularity int) (start int64, end int64) {
	ts := now.Unix()
	start = ts - ts%int64(granularity) - int64(granularity)
	// start = ts - ts%int64(granularity)
	end = start + int64(granularity)
	return
}

// NewJudgementEntity
func NewJudgementEntity(
	strategy *schema.Strategy,
	metrics []*MetricEntity,
	interval int) *JudgementEntity {

	if interval == 0 {
		logger.Warning(strategy.ID, "new judgement failed: zero interval")
		return nil
	}

	if len(metrics) == 0 {
		logger.Warning(strategy.ID, "new judgement failed: empty metric")
		return nil
	}

	// 默认允许30个周期的数据延迟
	window := interval * defaultWindowSizeByInterval
	// 如果策略中配置了, 以配置为准
	if strategy.WindowSize > 0 {
		// 与interval对齐
		window = strategy.WindowSize - strategy.WindowSize%interval + interval
	}

	driver, err := NewAlertPointDriver(strategy.Alert)
	if err != nil {
		logger.Warningf(strategy.ID, "new judgement failed: driver failed %v", err)
		return nil
	}

	if len(metrics) == 1 {
		return &JudgementEntity{
			interval:    interval,
			sid:         strategy.ID,
			lastEvent:   schema.EVENT_CODE_RECOVER,
			historySize: strategy.MaxEventHistorySize(interval),
			windowSize:  window,

			Metrics: metrics,
			Driver:  driver,
		}
	}
	sortedMetrics := MetricEntitys(metrics)
	sort.Sort(sortedMetrics)
	return &JudgementEntity{
		interval:    interval,
		sid:         strategy.ID,
		lastEvent:   schema.EVENT_CODE_RECOVER,
		historySize: strategy.MaxEventHistorySize(interval),
		windowSize:  window,

		Metrics: []*MetricEntity(sortedMetrics),
		Driver:  driver,
	}
}

func (je *JudgementEntity) ID() uint32 {
	return je.Metrics[0].ID
}

type infoTuple struct {
	ID    uint32
	Infos []string
	Op    string
}

// 执行trigger内的逻辑, 不触发dumpEvent
func (je *JudgementEntity) Run(
	stg storage.Storage,
	strategy *schema.Strategy,
	executions []*ExecutionEntity,
	now time.Time) []*schema.Event {

	var events []*schema.Event

	start, end := timeWindow(now, je.interval)
	if je.next == 0 {
		logger.Debugf(je.sid, "judgement first start, initialize")
		je.initialize()
	}
	// now对应的时刻已经执行过
	if je.next > start {
		logger.Warningf(je.sid, "judgement timestamp[%d] already finished, next is %d", start, je.next)
		return events
	}

	if je.windowSize > 0 {
		// je.next 远小于 now, 直接跳到最大窗口的起始时间
		// 对连续出现的null结果要有一个容忍度, 避免性能问题
		if start-je.next > int64(je.windowSize) {
			logger.Warningf(je.sid, "judgement timestamp[%d] is too old, reset to %d",
				je.next, start-int64(je.windowSize))

			je.next = start - int64(je.windowSize)
		}
	}

	je.deadline = end
	ongoings := make(map[int64]int)

	for current := je.next; current < je.deadline; current += int64(je.interval) {
		logger.Debugf(je.sid, "judgement timestamp[%d] start, series[%d], size:%d, execution:%d",
			current, je.Metrics[0].ID, len(je.Metrics), len(executions))

		final := schema.STATUS_INIT                    // 初始值
		fpoints := make(map[uint32][]*dataobj.RRDData) // 待写入缓存中的点
		finfos := make([]infoTuple, 0)                 // 用于填充event.Info字段

		for i := range je.Metrics {
			ifinal := schema.STATUS_INIT
			for j := range executions {
				if executions[j].key == je.Metrics[i].key {
					status, points, infos, op := executions[j].Run(stg, je.Metrics[i].ID, current, je.interval)

					ifinal = logicalOperate(strategy.Operator, ifinal, status)
					if _, found := ongoings[current]; !found {
						ongoings[current] = ifinal
					} else {
						// 多个指标或算子时, 一旦有一个出现null状态, 该时间戳状态就是null
						if !(ongoings[current] != schema.STATUS_NULL) {
							ongoings[current] = ifinal
						}
					}

					if len(infos) > 0 {
						finfos = append(finfos, infoTuple{ID: je.Metrics[i].ID, Infos: infos, Op: op})
					}
					if len(points) > 0 {
						if _, found := fpoints[je.Metrics[i].ID]; !found {
							fpoints[je.Metrics[i].ID] = points
							continue
						}
						fpoints[je.Metrics[i].ID] = append(fpoints[je.Metrics[i].ID], points...)
					}
				}
			}

			final = logicalOperate(strategy.Operator, final, ifinal)
		}
		// 上一次已经发出恢复通知, 本次的恢复可以直接跳过
		// 避免反复写入历史点
		if je.lastEvent == schema.EVENT_CODE_RECOVER &&
			final == schema.STATUS_RECOVER {
			// 需要把recover状态传递到driver
			je.Driver.Happen(current, final)
			logger.Debugf(je.sid, "judgement timepstamp[%d] duplicate recover, continue", current)
			continue
		}

		// alert -> recover 或 recover -> alert
		// 状态发生变化, 除最后一批点, 其他都没有意义了, 清理掉(特殊的cleanup)
		logger.Debugf(je.sid, "judgement driver happen, series[%d], timestamp[%d], status:%d",
			je.ID(), current, final)
		valid, changed := je.Driver.Happen(current, final)
		// 只有时间戳有效时, 才应该记录历史点, 否则会导致 重复覆盖的问题
		if valid {
			for i := range je.Metrics {
				points, found := fpoints[je.Metrics[i].ID]
				if found {
					// 运行时初始化
					if je.Metrics[i].History == nil {
						logger.Debugf(je.sid, "judgement new chain history, series[%d]", je.Metrics[i].ID)
						je.Metrics[i].History = buffer.NewChainHistory(je.historySize)
					}
					je.Metrics[i].History.Write(points)
				}
			}
		} else {
			// 时间戳错误, 打印日志
			if final == schema.STATUS_ALERT ||
				final == schema.STATUS_RECOVER {
				logger.Warningf(je.sid, "judgement timestamp[%d] invalid for happen() method of driver, series[%d]",
					current, je.Metrics[0].ID)
			}
		}

		// 如果当前记录的状态和之前比有变化, 需要清空历史记录, 仅保留最新的批次即可
		if changed {
			for i := range je.Metrics {
				if je.Metrics[i].History != nil {
					logger.Debugf(je.sid, "judgement timestamp[%d] cleanup, series[%d]",
						current, je.Metrics[i].ID)
					je.Metrics[i].History.Cleanup()
				}
			}
		}
		// 判断是否触发报警 or 报警解除
		// Metrics 用来组装历史点
		// strategy 用来填充必要字段
		// dump出事件(触发报警 或 触发解除) 需要对history执行 执行reset操作
		eventCode, clean := je.Driver.DumpEvent(current, je.interval)

		var event *schema.Event
		if eventCode == schema.EVENT_CODE_NULL {
			// 无事件
			// alert事件被频率限制, history中的数据不再有意义, 全部清空
			if clean {
				for i := range je.Metrics {
					if je.Metrics[i].History != nil {
						je.Metrics[i].History.Reset()
					}
				}
			}
		} else if eventCode == schema.EVENT_CODE_RECOVER {
			je.lastEvent = eventCode
			// 报警解除
			event = schema.NewEvent(strategy.ID,
				strategy.Partition, schema.EVENT_RECOVER)

		} else if eventCode == schema.EVENT_CODE_ALERT {
			je.lastEvent = eventCode
			// 触发报警
			event = schema.NewEvent(strategy.ID,
				strategy.Partition, schema.EVENT_ALERT)
		}
		if event != nil {
			for i := range je.Metrics {
				series, found := stg.Get(je.Metrics[i].ID)
				if !found {
					logger.Errorf(je.sid, "judgement series[%d] not found", je.Metrics[i].ID)
					continue
				}

				if je.Metrics[i].History == nil {
					logger.Warningf(je.sid, "judgement series[%d] history nil but event triggered",
						je.Metrics[i].ID)
					continue
				}

				points := je.Metrics[i].History.Dump()
				if len(points) == 0 {
					logger.Warningf(je.sid, "judgement series[%d] no points but event triggered",
						je.Metrics[i].ID)
					continue
				}
				// dump后, 需要显示调用一次 reset
				je.Metrics[i].History.Reset()
				event.SetPoints(series.Key(),
					series.Metric,
					series.Tags,
					series.Granularity,
					points)
			}
			// 更新info字段
			var info string
			for i := range finfos {
				series, found := stg.Get(finfos[i].ID)
				if !found {
					logger.Errorf(je.sid, "judgement series[%d] not found", finfos[i].ID)
					continue
				}
				for j := range finfos[i].Infos {
					if len(finfos) > 1 {
						if j == 0 {
							info += "("
						}
					}
					// info形如 happen(%s,3,1)=11 >10  其中 11是判断的中间结果, 10 是阈值
					// 通过 %s 填充 metric 信息
					info += fmt.Sprintf(finfos[i].Infos[j], series.Metric)
					if j < len(finfos[i].Infos)-1 {
						info += finfos[i].Op
					}
					if len(finfos) > 1 {
						if j == len(finfos[i].Infos)-1 {
							info += ")"
						}
					}
				}
				if i < len(finfos)-1 {
					info += " " + strategy.Operator + " "
				}
			}
			event.SetInfo(info)

			// 计算 hashid, 用于追踪 一个策略中 "一条曲线"的状态
			// 更新 detail 字段
			if event.Finalize() {
				events = append(events, event)
			}
		}
		logger.Debugf(je.sid, "judgement timestamp[%d] finished successfully", current)
	}

	logger.Debugf(je.sid, "judgement ongoing record:%v", ongoings)
	tss := make([]int, len(ongoings))
	i := 0
	for ts, _ := range ongoings {
		tss[i] = int(ts)
		i++
	}
	sort.Ints(tss)
	for i := len(tss) - 1; i >= 0; i-- {
		status, _ := ongoings[int64(tss[i])]
		if status == schema.STATUS_ALERT ||
			status == schema.STATUS_RECOVER ||
			status == schema.STATUS_EMPTY {
			je.next = int64(tss[i]) + int64(je.interval)
			logger.Debugf(je.sid, "judgement set next timestamp to %d", tss[i]+je.interval)
			break
		}
	}

	return events
}

func (je *JudgementEntity) Update(newest *JudgementEntity) {
	shouldUpdate := false
	if je.interval != newest.interval ||
		je.sid != newest.sid ||
		je.historySize != newest.historySize {
		shouldUpdate = true
	} else if !equalMetricEntitySlice(je.Metrics, newest.Metrics) {
		shouldUpdate = true
	}
	if shouldUpdate {
		logger.Debugf(je.sid, "judgement update, ID:%d", je.ID())
		newest.initialize()
		je = newest
	}
}

// initialize 初始化 deadline/next 字段
func (je *JudgementEntity) initialize() {
	start, end := timeWindow(time.Now(), je.interval)
	je.next = start
	je.deadline = end
}

func equalMetricEntitySlice(a, b []*MetricEntity) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].key != b[i].key {
			return false
		}
		if a[i].ID != b[i].ID {
			return false
		}
	}
	return true
}
