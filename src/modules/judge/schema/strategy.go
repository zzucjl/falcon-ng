package schema

import (
	"strconv"
)

func (s *Strategy) SetPartition(partition string) {
	s.Partition = partition
}

// 数据缓存需要的点数
func (s *Strategy) MaxBufferSizeAndSpan(interval int) (int, []int) {
	spanM := make(map[int]struct{})
	spanM[0] = struct{}{} // 默认同比0s, 即当前值

	base := 4      // 默认保存4个点, 硬编码
	maxPeriod := 1 // 再保存前一个周期的1个点
	for _, judgement := range s.Judgements {
		for _, expression := range judgement.Execution.Expressions {
			if expression.Func == TRIGGER_DURATION_HAPPEN ||
				expression.Func == TRIGGER_DURATION_STAT {
				if len(expression.Params) == 2 {
					if period, err := strconv.Atoi(expression.Params[0]); err == nil {
						if period/interval > maxPeriod {
							maxPeriod = period / interval
						}
					}
				}
			}
		}
	}
	span := make([]int, 0)
	for spanP := range spanM {
		span = append(span, spanP)
	}
	return base + maxPeriod, span
}

// 现场值需要的点数
func (s *Strategy) MaxEventHistorySize(interval int) int {
	base := s.Alert.AlertCountThreshold
	if s.Alert.RecoverCountThreshold > base {
		base = s.Alert.RecoverCountThreshold
	}
	if s.Alert.AlertDurationThreshold/interval > base {
		base = s.Alert.AlertDurationThreshold / interval
	}
	if s.Alert.RecoverDurationThreshold/interval > base {
		base = s.Alert.RecoverDurationThreshold / interval
	}

	maxPeriod := 0
	for _, judgement := range s.Judgements {
		for _, expression := range judgement.Execution.Expressions {
			if expression.Func == TRIGGER_DURATION_HAPPEN ||
				expression.Func == TRIGGER_DURATION_STAT {
				if len(expression.Params) == 2 {
					if period, err := strconv.Atoi(expression.Params[0]); err == nil {
						if period/interval > base {
							base = period / interval
						}
					}
				}
			}
		}
	}
	return base + maxPeriod
}

func (sj StrategyJudgement) Xclude() (include map[string][]string,
	exclude map[string][]string) {
	include = make(map[string][]string)
	exclude = make(map[string][]string)
	if len(sj.Tags) > 0 {
		for i := range sj.Tags {
			flag, tagk, tagv := sj.Tags[i].Xclude()
			if len(tagk) > 0 {
				if flag {
					include[tagk] = tagv
				} else {
					exclude[tagk] = tagv
				}
			}
		}
	}
	return
}

// 仅支持 通用的 = 和 !=
func (filter StrategyTagFilter) Xclude() (flag bool, tagk string, tagv []string) {
	if filter.Operator == "=" {
		return true, filter.Tagk, filter.Tagv
	}

	if filter.Operator == "!=" {
		return false, filter.Tagk, filter.Tagv
	}

	return // false, empty, empty
}

// true: 符合, false: 不符合
func (th StrategyThreshold) Compare(value float64) bool {
	switch th.Operator {
	case "=":
		return value == th.Threshold
	case ">=":
		return value >= th.Threshold
	case "<=":
		return value <= th.Threshold
	case ">":
		return value > th.Threshold
	case "<":
		return value < th.Threshold
	case "!=":
		return value != th.Threshold
	}
	return false
}
