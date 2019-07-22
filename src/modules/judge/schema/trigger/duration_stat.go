package trigger

import (
	"math"
	"strconv"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

// TriggerDurationStat 一段时间内统计特征
type TriggerDurationStat struct {
	TriggerThreshold
	TriggerInfo
	Duration int64  // 一段时间的定义, 以秒作为单位, int64方便和时间戳做比较
	Operator string // 操作符, 支持 max/min/sum/avg/obo
}

func NewTriggerDurationStat(
	thresholds []schema.StrategyThreshold, operator string,
	params []string) (TriggerDurationStat, error) {

	threshold, err := NewTriggerThreshold(thresholds, operator)
	if err != nil {
		return TriggerDurationStat{}, err
	}

	if len(params) != 2 {
		return TriggerDurationStat{}, ErrorParamIllegal
	}
	duration, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return TriggerDurationStat{}, ErrorParamIllegal
	}

	if params[1] != schema.MATH_OPERATOR_MAX &&
		params[1] != schema.MATH_OPERATOR_MIN &&
		params[1] != schema.MATH_OPERATOR_AVG &&
		params[1] != schema.MATH_OPERATOR_SUM &&
		params[1] != schema.MATH_OPERATOR_OBO {
		return TriggerDurationStat{}, ErrorParamIllegal
	}

	info := params[1] + "(%s," + params[0] + "s)"
	left, right := threshold.info()
	ti := NewTriggerInfo(info, left, right)

	return TriggerDurationStat{
		TriggerThreshold: threshold,
		TriggerInfo:      ti,
		Duration:         duration,
		Operator:         params[1],
	}, nil
}

func (tr TriggerDurationStat) Run(
	stg storage.Storage,
	ID uint32,
	current int64,
	granularity int) (status int, points []*dataobj.RRDData, info string, err error) {

	if ID <= 0 {
		return schema.STATUS_EMPTY, nil, "", ErrorQueryIDIllegal
	}
	var delta int64
	if tr.Duration < int64(granularity) {
		delta = int64(granularity) // 确保可以查询到一个周期
	} else {
		delta = tr.Duration - tr.Duration%int64(granularity)
	}

	// 查询范围 current-delta ~ current
	var (
		tps []*dataobj.RRDData
		ps  []*dataobj.RRDData // 避免分配内存

		enough = false
		etime  = current
		stime  = current - delta // 避开边界问题
	)

	tps, err = stg.Query(ID, stime, etime, 0)
	if err != nil {
		return schema.STATUS_NULL, nil, "", err
	}

	for i := range tps {
		if tps[i] == nil {
			continue
		}

		if tps[i].Timestamp > stime &&
			tps[i].Timestamp <= etime &&
			!math.IsNaN(float64(tps[i].Value)) {
			// 最新时刻的点查到, 说明没有延迟, 可以用来判断
			if tps[i].Timestamp == etime {
				enough = true
			}
			ps = append(ps, tps[i])
		}
	}
	if enough {
		var (
			final       float64
			oboFinal    bool
			trigged     bool
			shouldBreak bool
		)
		for i := range ps {
			value := float64(ps[i].Value)
			switch tr.Operator {
			case schema.MATH_OPERATOR_MAX:
				if i == 0 {
					final = value
				} else {
					if value > final {
						final = value
					}
				}

			case schema.MATH_OPERATOR_MIN:
				if i == 0 {
					final = value
				} else {
					if value < final {
						final = value
					}
				}
			case schema.MATH_OPERATOR_AVG:
				if len(ps) == 0 {
					continue
				}
				final += value / float64(len(ps))
			case schema.MATH_OPERATOR_SUM:
				final += value
			case schema.MATH_OPERATOR_OBO:
				final = value
				oboFinal = tr.Compare(value)
				// 有一个为false, 最终结果就是false
				shouldBreak = !oboFinal
			}
			if shouldBreak {
				break
			}
		}
		if tr.Operator == schema.MATH_OPERATOR_OBO {
			trigged = oboFinal
		} else {
			trigged = tr.Compare(final)
		}

		if trigged {
			return schema.STATUS_ALERT, ps, tr.Info(final), nil
		}
		return schema.STATUS_RECOVER, ps, tr.Info(final), nil
	}

	return schema.STATUS_NULL, nil, "", nil
}
