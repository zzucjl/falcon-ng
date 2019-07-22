package trigger

import (
	"math"
	"strconv"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

// TriggerNodata 数据中断
type TriggerNodata struct {
	TriggerInfo
	Duration int64 // 一段时间的定义, 以秒作为单位, int64方便和时间戳做比较
}

func NewTriggerNodata(params []string) (TriggerNodata, error) {
	if len(params) == 0 {
		return TriggerNodata{}, ErrorParamIllegal
	}
	duration, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil {
		return TriggerNodata{}, ErrorParamIllegal
	}

	// nodata(xxx,#50s)
	info := "nodata(%s,#" + params[0] + "s)"
	ti := NewTriggerInfo(info, []string{}, []string{})

	return TriggerNodata{
		TriggerInfo: ti,
		Duration:    duration,
	}, nil
}

func (tr TriggerNodata) Run(
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
		delta = tr.Duration
	}

	// 查询范围 current-delta+1000ms ~ current
	var (
		tps []*dataobj.RRDData
		ps  []*dataobj.RRDData // 避免分配内存

		etime = current
		stime = current - delta + 1
	)

	// 依赖: 数据中断, 查到的是 null点
	tps, err = stg.Query(ID, stime, etime, 0)
	// 查询出错, 不应该跳过
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
			ps = append(ps, tps[i])
		}
	}
	if len(ps) == 0 {
		// 向上传递一个 null点
		ps = append(ps, &dataobj.RRDData{
			Timestamp: current - int64(granularity),
			Value:     dataobj.JsonFloat(math.NaN()),
		})
		return schema.STATUS_ALERT, ps, tr.Info(), nil
	}

	// 向上传递一个 非null点
	return schema.STATUS_RECOVER, ps[(len(ps) - 1):], tr.Info(), nil
}
