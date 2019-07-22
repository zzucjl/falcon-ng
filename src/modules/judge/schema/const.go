package schema

const (
	STATUS_EMPTY   = iota // 代表时间戳直接跳过, 下次不用重试
	STATUS_RECOVER        // 恢复
	STATUS_ALERT          // 报警
	STATUS_NULL           // 最新点不存在, 下次需要重试
	STATUS_INIT           // 初始状态, 不参与计算
)

const (
	EVENT_CODE_NULL = iota // 既不报警, 也不报恢复
	EVENT_CODE_RECOVER
	EVENT_CODE_ALERT
)

const (
	CATEGORY_THRESHOLD = 1 // 阈值

	EVENT_ALERT   = "alert"
	EVENT_RECOVER = "recovery"

	LOGIC_OPERATOR_AND = "and" // 与
	LOGIC_OPERATOR_OR  = "or"  // 或

	TRIGGER_DURATION_HAPPEN = "duration_happen"
	TRIGGER_DURATION_STAT   = "duration_stat"
	TRIGGER_NODATA          = "nodata"

	MATH_OPERATOR_MAX = "max"
	MATH_OPERATOR_MIN = "min"
	MATH_OPERATOR_AVG = "avg"
	MATH_OPERATOR_SUM = "sum"
	MATH_OPERATOR_OBO = "all" // one by one

	// 曲线相关
	ENDPOINT_KEYWORD  = "endpoint"
	COUNTER_SEPERATOR = "/"
)

func MapStatus(code int) string {
	if code == STATUS_RECOVER {
		return "RECOVER"
	} else if code == STATUS_ALERT {
		return "ALERT"
	} else if code == STATUS_NULL {
		return "NULL"
	} else if code == STATUS_INIT {
		return "INIT"
	}

	return "EMPTY"
}
