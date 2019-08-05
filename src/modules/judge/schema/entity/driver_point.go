package entity

import (
	"errors"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
)

// AlertPointDriver 点驱动模型, 只记录异常/恢复的时间, 不记录没有点的情况
type AlertPointDriver struct {
	AlertTimestamp      []int64 `json:"alert_timestamp"`       // 异常时间, 从小到大排列, 默认写入是有序的, 乱序的直接丢掉, 打印fatal日志
	LastAlert           int64   `json:"last_alert"`            // AlertTimestamp 数组最后一个值
	RecoverTimestamp    []int64 `json:"recover_timestamp"`     // 恢复时间, 从小到大排列, 与AlertTimestamp不能同时存在
	LastRecover         int64   `json:"last_recover"`          // RecoverTimestamp 数组最后一个值
	LastStatus          int     `json:"last_status"`           // 上次写入的状态
	EventAlertTimestamp []int64 `json:"event_alert_timestamp"` // 触发报警的时间,
	LastEventStatus     int     `json:"last_event_status"`     // 恢复之后再恢复就不要报了, 报警之后再报警就是 再次报警

	AlertCountThreshold      int   `json:"alert_count"`      // 按次数判断的报警
	RecoverCountThreshold    int   `json:"recover_count"`    // 按次数判断的解除
	AlertDurationThreshold   int64 `json:"alert_duration"`   // 按持续时间判断的报警, 单位 秒
	RecoverDurationThreshold int64 `json:"recover_duration"` // 按持续时间判断的解除, 单位 秒
	LimitCountThreshold      int   `json:"limit_count"`      // xx秒最多报警xx次, 次数
	LimitDurationThreshold   int64 `json:"limit_duration"`   // xx秒最多报警xx次, 时间窗口; 连续xx秒再次发出报警概念类似
}

// NewAlertPointDriver 返回一个结构体指针
func NewAlertPointDriver(alert schema.StrategyAlert) (*AlertPointDriver, error) {
	// 永远无法触发报警
	if alert.AlertCountThreshold <= 0 && alert.AlertDurationThreshold <= 0 {
		return nil, errors.New("alert threshold illegal")
	}
	// 永远无法触发恢复, count 不允许为0, duration 可以为0
	if alert.RecoverCountThreshold <= 0 && alert.RecoverDurationThreshold < 0 {
		return nil, errors.New("recover threshold illegal")
	}
	return &AlertPointDriver{
		AlertTimestamp:      make([]int64, 0),
		LastAlert:           0,
		RecoverTimestamp:    make([]int64, 0),
		LastRecover:         0,
		LastStatus:          schema.STATUS_RECOVER, // 运行过程中只会出现 alert/recover 两个状态
		EventAlertTimestamp: make([]int64, 0),
		LastEventStatus:     schema.EVENT_CODE_RECOVER, // 初始化是recover状态

		AlertCountThreshold:      alert.AlertCountThreshold,
		RecoverCountThreshold:    alert.RecoverCountThreshold,
		AlertDurationThreshold:   int64(alert.AlertDurationThreshold),
		RecoverDurationThreshold: int64(alert.RecoverDurationThreshold),
		LimitCountThreshold:      alert.LimitCountThreshold,
		LimitDurationThreshold:   int64(alert.LimitDurationThreshold),
	}, nil
}

func (pd *AlertPointDriver) SetThreshold(alert schema.StrategyAlert) error {
	// 永远无法触发报警
	if alert.AlertCountThreshold <= 0 && alert.AlertDurationThreshold <= 0 {
		return errors.New("alert threshold illegal")
	}
	// 永远无法触发恢复, count 不允许为0, duration 可以为0
	if alert.RecoverCountThreshold <= 0 && alert.RecoverDurationThreshold < 0 {
		return errors.New("recover threshold illegal")
	}

	pd.AlertCountThreshold = alert.AlertCountThreshold
	pd.RecoverCountThreshold = alert.RecoverCountThreshold
	pd.AlertDurationThreshold = int64(alert.AlertDurationThreshold)
	pd.RecoverDurationThreshold = int64(alert.RecoverDurationThreshold)
	pd.LimitCountThreshold = alert.LimitCountThreshold
	pd.LimitDurationThreshold = int64(alert.LimitDurationThreshold)

	return nil
}

func (pd *AlertPointDriver) Happen(ts int64, status int) (valid bool, changed bool) {
	valid = false
	changed = false
	// EMPTY 无状态的意义
	if status == schema.STATUS_EMPTY {
		return // false, false
	}
	// NULL状态, 不需要记录历史点, 直接跳过
	if status == schema.STATUS_NULL {
		return // false, false
	}
	if status == schema.STATUS_RECOVER {
		if ts <= pd.LastRecover {
			logger.Debugf(0, "driver invald, lastRecover:%d, current:%d", pd.LastRecover, ts)
			return // false, false
		}
		valid = true

		if pd.LastStatus != status {
			pd.LastStatus = status
			changed = true
			// 状态变化时, 相反的状态要清空
			pd.AlertTimestamp = pd.AlertTimestamp[:0]
		}
		// 已经报过恢复 or 进程刚启动, 不用记录
		if pd.LastEventStatus == schema.EVENT_CODE_RECOVER {
			// changed = true 触发history的cleanup动作
			changed = true
			return
		}
		pd.RecoverTimestamp = append(pd.RecoverTimestamp, ts)
		pd.LastRecover = ts
		return
	}
	if status == schema.STATUS_ALERT {
		if ts <= pd.LastAlert {
			logger.Debugf(0, "driver invald, lastAlert:%d, current:%d", pd.LastAlert, ts)
			return // false, false
		}
		valid = true

		if pd.LastStatus != status {
			pd.LastStatus = status
			changed = true
			// 状态变化时, 相反的状态要清空
			pd.RecoverTimestamp = pd.RecoverTimestamp[:0]
		}
		pd.AlertTimestamp = append(pd.AlertTimestamp, ts)
		pd.LastAlert = ts
		return
	}

	return
}

// 注意传入的 now, 是 上层结构的当前进度, 而不是 time.Now()
// interval 用来解决边界问题, nowTs 都是对齐到时间窗口起点的, 会导致按照时间跨度做阈值的判断, 都多出一个interval
func (pd *AlertPointDriver) DumpEvent(nowTs int64, interval ...int) (int, bool) {
	shift := int64(0)
	if len(interval) > 0 && interval[0] > 0 {
		shift = int64(interval[0])
	}
	event := schema.EVENT_CODE_NULL
	clean := false

	// 恢复/故障 --> 故障
	if len(pd.AlertTimestamp) > 0 {
		// 按次数判断的报警
		if (pd.AlertCountThreshold > 0 &&
			len(pd.AlertTimestamp) >= pd.AlertCountThreshold) ||
			// 按持续时间判断的报警
			(pd.AlertDurationThreshold > 0 &&
				nowTs-pd.AlertTimestamp[0] >= pd.AlertDurationThreshold-shift) {

			// 触发报警后, 所有的记录都用过了, 需要丢弃掉
			// 比如 时刻 1,2,3 触发了报警, 时刻4 不再使用2,3 的状态
			pd.AlertTimestamp = pd.AlertTimestamp[:0]

			// xx秒最多报警xx次, 恢复不清零
			// 清理过期记录
			if pd.LimitCountThreshold > 0 && pd.LimitDurationThreshold > 0 {
				ts := pd.LastAlert
				if pd.LastEventStatus == schema.EVENT_CODE_ALERT {
					start := 0
					end := len(pd.EventAlertTimestamp)
					for ; start < end &&
						pd.EventAlertTimestamp[start] <= nowTs-(pd.LimitDurationThreshold-shift); start++ {
					}
					pd.EventAlertTimestamp = pd.EventAlertTimestamp[start:end]
				}
				if len(pd.EventAlertTimestamp) < pd.LimitCountThreshold {
					pd.EventAlertTimestamp = append(pd.EventAlertTimestamp, ts)
					// 发出报警事件
					event = schema.EVENT_CODE_ALERT

					pd.LastEventStatus = schema.EVENT_CODE_ALERT
				} else {
					clean = true
				}
			} else {
				// 不限制 xx秒最多报警xx次
				// 发出报警事件
				event = schema.EVENT_CODE_ALERT
				pd.LastEventStatus = schema.EVENT_CODE_ALERT
			}
		}
	}
	// 故障 --> 恢复
	if pd.LastEventStatus == schema.EVENT_CODE_ALERT &&
		len(pd.RecoverTimestamp) > 0 {
		// 按次数判断的解除
		if (pd.RecoverCountThreshold > 0 &&
			len(pd.RecoverTimestamp) >= pd.RecoverCountThreshold) ||
			// 按持续时间判断的解除, 单位 秒
			(pd.RecoverDurationThreshold > 0 &&
				nowTs-pd.RecoverTimestamp[0] >= pd.RecoverDurationThreshold-shift) ||
			// 出现恢复则认为解除
			(pd.RecoverCountThreshold == 0 && pd.RecoverDurationThreshold == 0) {

			// 触发事件后, 所有记录都不需要了, 直接丢弃掉
			pd.RecoverTimestamp = pd.RecoverTimestamp[:0]
			// 恢复发生时, 报警次数限制的统计归零, 重新开始计数
			pd.EventAlertTimestamp = pd.EventAlertTimestamp[:0]

			// 发出解除事件
			event = schema.EVENT_CODE_RECOVER
			pd.LastEventStatus = schema.EVENT_CODE_RECOVER
		}
	}

	// 恢复 --> 恢复, 不发出任何事件
	if pd.LastAlert == schema.EVENT_CODE_RECOVER &&
		len(pd.RecoverTimestamp) > 0 {
		// 清空记录, 如果有
		pd.RecoverTimestamp = pd.RecoverTimestamp[:0]
	}

	return event, clean
}
