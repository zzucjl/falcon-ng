package entity

import (
	"fmt"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/bitmap"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/trigger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

const (
	EFFECTIVE_DAY_SIZE    = 7
	EFFECTIVE_MINUTE_SIZE = 1440
)

// ExecutionEntity 的具体实现
func NewExecutionEntity(strategy *schema.Strategy,
	se schema.StrategyExecution, key string) (*ExecutionEntity, error) {
	if len(se.Expressions) == 0 {
		return nil, fmt.Errorf("empty expression")
	}
	if se.Operator != schema.LOGIC_OPERATOR_AND &&
		se.Operator != schema.LOGIC_OPERATOR_OR {
		return nil, fmt.Errorf("illegal execution operator")
	}
	effectiveDay := bitmap.NewBitMap(7)
	if len(se.EffectiveDay) > 0 {
		effectiveDay.Set(se.EffectiveDay...)
	}
	effectiveMinute := bitmap.NewBitMap(1440)
	effectiveMinute.SetRange(se.EffectiveStart, se.EffectiveEnd)
	var triggers []trigger.Trigger

	for i := range se.Expressions {
		var (
			trg trigger.Trigger
			err error
		)
		switch se.Expressions[i].Func {
		case schema.TRIGGER_DURATION_HAPPEN:
			trg, err = trigger.NewTriggerDurationHappen(
				se.Expressions[i].Thresholds,
				se.Expressions[i].Operator,
				se.Expressions[i].Params,
			)

		case schema.TRIGGER_DURATION_STAT:
			trg, err = trigger.NewTriggerDurationStat(
				se.Expressions[i].Thresholds,
				se.Expressions[i].Operator,
				se.Expressions[i].Params,
			)

		case schema.TRIGGER_NODATA:
			trg, err = trigger.NewTriggerNodata(
				se.Expressions[i].Params,
			)
		default:
			err = fmt.Errorf("unknown func:%s", se.Expressions[i].Func)
		}
		// nil 参数错误报错
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, trg)
	}
	return &ExecutionEntity{
		sid: strategy.ID,
		key: key,

		EffectiveDay:    effectiveDay,
		EffectiveMinute: effectiveMinute,
		Operator:        se.Operator,
		Triggers:        triggers,
	}, nil
}

func (ee *ExecutionEntity) Effective(day int, minute int) bool {
	if ee.EffectiveDay.IsSet(day) &&
		ee.EffectiveMinute.IsSet(minute) {
		return true
	}
	return false
}

func (ee *ExecutionEntity) Run(stg storage.Storage,
	ID uint32, current int64, granularity int) (int, []*dataobj.RRDData, []string, string) {

	if len(ee.Triggers) == 0 {
		return schema.STATUS_EMPTY, nil, []string{}, ""
	}
	if ID <= 0 {
		return schema.STATUS_EMPTY, nil, []string{}, ""
	}

	var (
		points = make(map[int64]*dataobj.RRDData)
		infos  = make([]string, 0)
		status = schema.STATUS_INIT
	)

	for i := range ee.Triggers {
		istatus, ipoints, info, err := ee.Triggers[i].Run(stg, ID, current, granularity)
		if err != nil {
			logger.Warningf(ee.sid, "timestamp[%d] trigger run failed: %v", current, err)
		}
		status = logicalOperate(ee.Operator, status, istatus)

		if len(ipoints) > 0 {
			for j := range ipoints {
				points[ipoints[j].Timestamp] = ipoints[j]
			}
		}
		if len(info) > 0 {
			infos = append(infos, info)
		}
	}
	// null/empty 都代表跳过当前点; null 需要下次重试, empty 不需要下次重试
	if status == schema.STATUS_NULL || status == schema.STATUS_EMPTY {
		return status, nil, []string{}, ""
	}

	ret := make([]*dataobj.RRDData, len(points))
	i := 0
	for _, p := range points {
		ret[i] = p
		i++
	}
	return status, ret, infos, ee.Operator
}
