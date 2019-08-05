package worker

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/bitmap"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/entity"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/query"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/series"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

var (
	// schema/entity部分更多的关注到"逻辑的实现", 需要实体化的程序放在worker中
	_nodataDrivers map[int64]map[string]entity.AlertDriverEntity
	_options       StrategyConfigOption
)

type NodataJudgement struct {
	Sid     int64
	Metric  string
	Include map[string][]string
	Exclude map[string][]string
}

type NodataStrategy struct {
	*schema.Strategy
	Judgements []*NodataJudgement
	Endpoints  []string
	Metrics    string
	Alert      schema.StrategyAlert
	Info       string
}

// nodata策略的前置判断, 索引不存在(与预期的tag过滤规则不匹配)则触发报警
// 索引存在时的数据上报中断由 TriggerNodata 解决
func PreHandlerNodata(ss []*schema.Strategy, now time.Time) {
	if len(ss) == 0 {
		return
	}
	logger.Debugf(0, "nodata pre handler start, timestamp:%d", now.Unix())
	for i := range ss {
		ok, nstra := GenerateNodataStrategy(ss[i], now)
		if ok {
			logger.Debugf(nstra.ID, "nodata pre handler strategy start to run, timestamp:%d", now.Unix())

			events := nstra.Run(now)
			if len(events) > 0 {
				nstra.Publish(events)
			}
		}
	}
}

func GenerateNodataStrategy(s *schema.Strategy, now time.Time) (bool, *NodataStrategy) {
	if len(s.Judgements) == 0 {
		return false, nil
	}

	if len(s.Endpoints) == 0 {
		return false, nil
	}

	var (
		mmap       = make(map[string]struct{})
		judgements []*NodataJudgement
		metric     string
		duration   int
	)
	for i := range s.Judgements {
		ok, param := isNodataExecution(s.Judgements[i].Execution, now)
		if ok {
			if i == 0 {
				duration = param
			} else if duration > param {
				// 取最小值
				duration = param
			}

			include, exclude := s.Judgements[i].Xclude()
			judgement := &NodataJudgement{
				Sid:     s.ID,
				Metric:  s.Judgements[i].Metric,
				Include: include,
				Exclude: exclude,
			}
			judgements = append(judgements, judgement)
			mmap[s.Judgements[i].Metric] = struct{}{}
		}
	}
	if len(judgements) == 0 {
		return false, nil
	}

	if len(mmap) == 0 {
		return false, nil
	}

	i := 0
	for k, _ := range mmap {
		metric += k
		if i < len(mmap)-1 {
			metric += ","
			i++
		}
	}

	alert := schema.StrategyAlert{
		AlertDurationThreshold:   duration,
		RecoverCountThreshold:    s.Alert.RecoverCountThreshold,
		RecoverDurationThreshold: s.Alert.RecoverDurationThreshold,
		LimitCountThreshold:      s.Alert.LimitCountThreshold,
		LimitDurationThreshold:   s.Alert.LimitDurationThreshold,
	}

	info := "nodata(" + metric + ",#" +
		strconv.Itoa(alert.AlertDurationThreshold) + "s)"

	return true, &NodataStrategy{
		Strategy:   s,
		Judgements: judgements,
		Endpoints:  s.Endpoints,
		Metrics:    metric,
		Alert:      alert,
		Info:       info,
	}
}

func (s *NodataStrategy) Run(now time.Time) []*schema.Event {
	var events []*schema.Event
	if len(s.Judgements) == 0 {
		return events
	}
	fstatus := make(map[string]int)
	for i := range s.Judgements {
		// 不同指标的nodata与条件, 包含的tag条件不同时, 永远无法触发报警
		counters := s.Judgements[i].Run(s.Endpoints)
		if i == 0 {
			for j := range counters {
				fstatus[counters[j]] = schema.STATUS_ALERT
			}
		} else {
			for j := range counters {
				fstatus[counters[j]] = logicalOperate(s.Operator,
					schema.STATUS_ALERT, fstatus[counters[j]])
			}
		}
	}

	drivers, found := _nodataDrivers[s.ID]
	if !found {
		drivers = make(map[string]entity.AlertDriverEntity)
		_nodataDrivers[s.ID] = drivers
	}

	for counter, driver := range drivers {
		// 更新阈值判断
		driver.SetThreshold(s.Alert)

		var (
			shouldDelete = true
			event        *schema.Event
		)
		status, _ := fstatus[counter]
		if status == schema.STATUS_ALERT {
			shouldDelete = false
			driver.Happen(now.Unix(), schema.STATUS_ALERT)
		} else {
			// !found  or status != schema.STATUS_ALERT
			driver.Happen(now.Unix(), schema.STATUS_RECOVER)
		}

		eventCode, _ := driver.DumpEvent(now.Unix())

		if eventCode == schema.EVENT_CODE_ALERT {
			event = schema.NewEvent(s.ID, s.Partition, schema.EVENT_ALERT)
			logger.Debugf(s.ID, "nodata pre handler alert, timestamp:%d", now.Unix())
		} else if eventCode == schema.EVENT_CODE_RECOVER {
			if shouldDelete {
				delete(drivers, counter)
			}
			event = schema.NewEvent(s.ID, s.Partition, schema.EVENT_RECOVER)
			logger.Debugf(s.ID, "nodata pre handler recovery, timestamp:%d", now.Unix())
		}
		if event != nil {
			event.SetPoints(
				counter,
				s.Metrics,
				series.CounterString2TagMap(counter),
				0,
				newNodataPoint(eventCode, now),
			)
			if event.Finalize() {
				event.Info = s.Info
				events = append(events, event)
			}
		}
	}

	for counter, status := range fstatus {
		if _, found := drivers[counter]; !found && status == schema.STATUS_ALERT {
			driver, err := entity.NewAlertPointDriver(s.Alert)
			if err != nil {
				logger.Warningf(s.ID, "nodata pre handler new driver failed: error %v", err)
				continue
			}
			drivers[counter] = driver
			driver.Happen(now.Unix(), schema.STATUS_ALERT)

			eventCode, _ := driver.DumpEvent(now.Unix())

			if eventCode == schema.EVENT_CODE_ALERT {
				event := schema.NewEvent(s.ID, s.Partition, schema.EVENT_ALERT)
				logger.Debugf(s.ID, "nodata pre handler recovery, timestamp:%d", now.Unix())

				event.SetPoints(
					counter,
					s.Metrics,
					series.CounterString2TagMap(counter),
					0,
					newNodataPoint(eventCode, now),
				)
				if event.Finalize() {
					event.Info = s.Info
					events = append(events, event)
				}
			}
		}
	}

	return events
}

func (s *NodataStrategy) Publish(events []*schema.Event) {
	if len(events) == 0 {
		return
	}
	if pub == nil {
		logger.Error(s.ID, "nodata pre handler event push failed: publisher not set")
		return
	}

	for i := range events {
		err := pub.Publish(events[i])
		if err != nil {
			logger.Warningf(s.ID, "nodata pre handler event push failed: %v", err)
			continue
		}
		// 特殊使用, 需要执行json打印出详细的信息
		if logger.GetSeverity(s.ID) == logger.DEBUG {
			bytes, _ := json.Marshal(events[i])
			logger.Debugf(s.ID, "nodata pre handler event push success: %s", string(bytes))
		}
	}
}

// 避免无效的计算, 按照endpoint 一个个查询
func (j *NodataJudgement) Run(endpoints []string) []string {
	if len(endpoints) == 0 {
		return []string{}
	}
	expected := j.ExpectedCounters(endpoints)
	if len(expected) == 0 {
		return []string{}
	}
	counters := make(map[string]struct{})
	for _, endpoint := range endpoints {
		result, err := stg.Index(
			query.NewIndexRequest(endpoint, j.Metric, j.Include, j.Exclude),
		)
		if err != nil {
			logger.Warningf(j.Sid, "nodata pre handler judgement run failed: index error %v", err)
			continue
		}
		for j := range result {
			counters[result[j].Counter] = struct{}{}
		}
	}
	if len(counters) == 0 {
		return expected
	}
	var ret []string
	for i := range expected {
		found := false
		for counter, _ := range counters {
			if strings.Contains(counter, expected[i]) {
				found = true
				break
			}
		}
		if !found {
			ret = append(ret, expected[i])
		}
	}
	return ret
}

func (j *NodataJudgement) ExpectedCounters(endpoints []string) []string {

	var tagkvs Tagkvs
	// 如果tag过滤中配置 include endpoint, 以此为准
	if _, include := j.Include[schema.ENDPOINT_KEYWORD]; include {
		if len(j.Include[schema.ENDPOINT_KEYWORD]) == 0 {
			return []string{}
		}
		tagkvs = append(tagkvs, Tagkv{
			Tagk: schema.ENDPOINT_KEYWORD,
			Tagv: j.Include[schema.ENDPOINT_KEYWORD],
		})
	} else if _, exclude := j.Exclude[schema.ENDPOINT_KEYWORD]; exclude {
		// 如果tag过滤中配置 exclude endpoint, 则排除掉
		var left []string
		for i := range endpoints {
			remain := true
			for _, endopoint := range j.Exclude[schema.ENDPOINT_KEYWORD] {
				if endpoints[i] == endopoint {
					remain = false
					break
				}
			}
			if remain {
				left = append(left, endpoints[i])
			}
		}
		if len(left) == 0 {
			return []string{}
		}
		tagkvs = append(tagkvs, Tagkv{
			Tagk: schema.ENDPOINT_KEYWORD,
			Tagv: left,
		})
	} else {
		// 否则 以节点的endpoints为准
		if len(endpoints) == 0 {
			return []string{}
		}
		tagkvs = append(tagkvs, Tagkv{
			Tagk: schema.ENDPOINT_KEYWORD,
			Tagv: endpoints,
		})
	}

	// include规则和exclude存在相同的k-v对时, exclude的优先级更高
	for k, v := range j.Include {
		if k == schema.ENDPOINT_KEYWORD {
			continue
		}
		if len(v) > 0 {
			var left []string
			if _, exclude := j.Exclude[k]; exclude {
				for i := range v {
					remain := true
					for _, ex := range j.Exclude[k] {
						if v[i] == ex {
							remain = false
							break
						}
					}
					if remain {
						left = append(left, v[i])
					}
				}
			} else {
				left = v
			}
			if len(left) > 0 {
				tagkvs = append(tagkvs, Tagkv{
					Tagk: k,
					Tagv: left,
				})
			}
		}
	}
	sort.Sort(tagkvs)

	return CartesianIteration(tagkvs)
}

func isNodataExecution(ee schema.StrategyExecution, now time.Time) (bool, int) {
	if len(ee.Expressions) == 0 {
		return false, 0
	}
	for i := range ee.Expressions {
		if ee.Expressions[i].Func == schema.TRIGGER_NODATA {
			// nodata的参数错误
			if len(ee.Expressions[i].Params) < 1 {
				continue
			}
			param, err := strconv.Atoi(ee.Expressions[i].Params[0])
			if err != nil {
				continue
			}
			if len(ee.EffectiveDay) == 0 {
				continue
			}
			effectiveDay := bitmap.NewBitMap(7)
			effectiveDay.Set(ee.EffectiveDay...)
			effectiveMinute := bitmap.NewBitMap(1440)
			effectiveMinute.SetRange(
				ee.EffectiveStart,
				ee.EffectiveEnd,
			)
			if effectiveDay.IsSet(int(now.Weekday())) &&
				effectiveMinute.IsSet(now.Hour()*60+now.Minute()) {
				return true, param
			}
		}
	}
	return false, 0
}

func logicalOperate(operator string, a, b int) int {
	if operator == schema.LOGIC_OPERATOR_OR {
		if a == schema.STATUS_RECOVER &&
			b == schema.STATUS_RECOVER {
			return schema.STATUS_RECOVER
		}
		return schema.STATUS_ALERT
	}

	if operator == schema.LOGIC_OPERATOR_AND {
		if a == schema.STATUS_ALERT &&
			b == schema.STATUS_ALERT {
			return schema.STATUS_ALERT
		}
		return schema.STATUS_RECOVER
	}
	return schema.STATUS_RECOVER
}

func newNodataPoint(code int, now time.Time) []*dataobj.RRDData {
	var point *dataobj.RRDData
	if code == schema.EVENT_CODE_ALERT {
		point = &dataobj.RRDData{
			Timestamp: now.Unix(),
			Value:     dataobj.JsonFloat(math.NaN()),
		}
	} else {
		point = &dataobj.RRDData{
			Timestamp: now.Unix(),
			Value:     dataobj.JsonFloat(0), // 目前无意义
		}
	}

	return []*dataobj.RRDData{point}
}

type Tagkv struct {
	Tagk string
	Tagv []string
}

type Tagkvs []Tagkv

func (t Tagkvs) Len() int           { return len(t) }
func (t Tagkvs) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t Tagkvs) Less(i, j int) bool { return t[i].Tagk < t[j].Tagk }

// 笛卡尔乘积, 迭代
func CartesianIteration(l Tagkvs) []string {
	max := int(1)
	for i := 0; i < len(l); i++ {
		max *= len(l[i].Tagv)
	}
	repeat := max

	ret := make([]string, max)
	for i := 0; i < len(l); i++ {
		repeat = repeat / len(l[i].Tagv)
		for j := 0; j < max; {
			tagk := l[i].Tagk
			for k := 0; k < len(l[i].Tagv); k++ {
				if repeat == 1 {
					if i < len(l)-1 {
						ret[j] += tagk + "=" + l[i].Tagv[k] + ","
					} else {
						ret[j] += tagk + "=" + l[i].Tagv[k]
					}
					j++
				} else {
					for m := 0; m < repeat; m++ {
						ret[j] += tagk + "=" + l[i].Tagv[k] + ","
						j++
					}
				}
			}
		}
	}

	return ret
}
