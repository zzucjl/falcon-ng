package trigger

import (
	"math"
	"strconv"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

// TriggerDurationHappen 一段时间内发生
type TriggerDurationHappen struct {
	TriggerThreshold
	TriggerInfo          // info结构, 预初始化, 避免重复计算
	Duration       int64 // 一段时间的定义, 以秒作为单位, int64方便和时间戳做比较
	LimitThreshold int   // 发生次数, 阈值
}

func NewTriggerDurationHappen(
	thresholds []schema.StrategyThreshold, operator string,
	params []string) (TriggerDurationHappen, error) {

	threshold, err := NewTriggerThreshold(thresholds, operator)
	if err != nil {
		return TriggerDurationHappen{}, err
	}

	if len(params) != 2 {
		return TriggerDurationHappen{}, ErrorParamIllegal
	}
	var (
		duration int64
		limit    int
		info     string
	)
	duration, err = strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return TriggerDurationHappen{}, ErrorParamIllegal
	}
	limit, err = strconv.Atoi(params[1])
	if err != nil {
		return TriggerDurationHappen{}, ErrorParamIllegal
	}

	if duration <= 0 || limit <= 0 {
		return TriggerDurationHappen{}, ErrorParamIllegal

	}

	// happen(xxx,10s,1)
	info = "happen(%s," + params[0] + "s," + params[1] + ")"
	left, right := threshold.info()
	ti := NewTriggerInfo(info, left, right)

	return TriggerDurationHappen{
		TriggerThreshold: threshold,
		TriggerInfo:      ti,
		Duration:         int64(duration),
		LimitThreshold:   limit,
	}, nil
}

func (tr TriggerDurationHappen) Run(
	stg storage.Storage,
	ID uint32,
	current int64,
	granularity int) (status int, points []*dataobj.RRDData, info string, err error) {

	if ID <= 0 {
		return schema.STATUS_EMPTY, nil, "", ErrorQueryIDIllegal
	}
	// 如果配置了策略 30秒内有100个点, 显然永远不会触发, 直接跳过
	// 允许一个周期的边界, 比如 12s内有2个点, 也算是合理
	if int64(tr.LimitThreshold*granularity) > tr.Duration+int64(granularity) {
		return schema.STATUS_EMPTY, nil, "", ErrorParamIllegal
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
		// 中间有断点 导致不足N个点, 按照EMPTY状态处理
		if len(ps) < tr.LimitThreshold {
			return schema.STATUS_EMPTY, nil, "", nil
		}
		count := 0
		for i := range ps {
			if tr.Compare(float64(ps[i].Value)) {
				count++
			}
		}
		if count >= tr.LimitThreshold {
			return schema.STATUS_ALERT, ps,
				tr.Info(float64(ps[len(ps)-1].Value)), nil
		}
		return schema.STATUS_RECOVER, ps,
			tr.Info(float64(ps[len(ps)-1].Value)), nil
	}

	return schema.STATUS_NULL, nil, "", nil
}
