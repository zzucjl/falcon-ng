package trigger

import (
	"errors"
	"fmt"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

var (
	ErrorParamNotEnough    = errors.New("param not enough")
	ErrorParamIllegal      = errors.New("param illegal")
	ErrorThresholdOperator = errors.New("illegal expression operator")
	ErrorThresholdEmpty    = errors.New("empty expression threshold")
	ErrorQueryIDIllegal    = errors.New("illegal ID")
)

// 算子

// Trigger 不需要保留任何属性字段, 全部从上层继承
// trigger
type Trigger interface {
	// Run 查询数据并判断
	// 参数: Storage, 曲线ID(大于0), 当前时刻(stime, etime=current+granularity), 时间粒度
	// 返回: 报警状态, 现场值(如果故障或恢复),
	Run(stg storage.Storage,
		ID uint32,
		current int64,
		granularity int) (status int, points []*dataobj.RRDData, info string, err error)
}

type TriggerThreshold struct {
	thresholds []schema.StrategyThreshold
	operator   string
}

func NewTriggerThreshold(thresholds []schema.StrategyThreshold,
	operator string) (TriggerThreshold, error) {
	if len(thresholds) == 0 {
		return TriggerThreshold{}, ErrorThresholdEmpty
	}
	if operator != schema.LOGIC_OPERATOR_AND &&
		operator != schema.LOGIC_OPERATOR_OR {
		return TriggerThreshold{}, ErrorThresholdOperator
	}

	return TriggerThreshold{
		thresholds: thresholds,
		operator:   operator,
	}, nil
}

// true: 符合, false: 不符合
func (th TriggerThreshold) Compare(value float64) bool {
	var status bool
	for i := range th.thresholds {
		triggered := th.thresholds[i].Compare(value)
		if i == 0 {
			status = triggered
			continue
		}
		if th.operator == schema.LOGIC_OPERATOR_AND {
			status = status && triggered
		}
		if th.operator == schema.LOGIC_OPERATOR_OR {
			status = status || triggered
		}
	}
	return status
}

func (th TriggerThreshold) info() (left []string, right []string) {
	if len(th.thresholds) == 1 {
		var info string
		info = th.thresholds[0].Operator + fmt.Sprintf("%.2f", th.thresholds[0].Threshold)
		return []string{}, []string{info}
	}
	return []string{}, []string{}
}

type TriggerInfo struct {
	info   string
	prefix string
	suffix string
}

func NewTriggerInfo(info string, left []string, right []string) TriggerInfo {
	var (
		prefix string
		suffix string
	)
	if len(left) > 0 {
		for i := range left {
			prefix += left[i]
			if i < len(left)-1 {
				prefix += ","
			}
		}
	}
	if len(right) > 0 {
		for i := range right {
			suffix += right[i]
			if i < len(right)-1 {
				suffix += ","
			}
		}
	}
	return TriggerInfo{
		info:   info,
		prefix: prefix,
		suffix: suffix,
	}
}

func (ti TriggerInfo) Info(value ...float64) string {
	var info string
	if len(ti.prefix) > 0 {
		info = ti.prefix
	}
	if len(value) > 0 {
		info += ti.info + fmt.Sprintf("=%.2f", value[0])
	} else {
		info += ti.info
	}

	if len(ti.suffix) > 0 {
		info += " " + ti.suffix
	}
	return info
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a > b {
		return b
	}
	return a
}
