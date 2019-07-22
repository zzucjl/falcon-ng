package schema

import (
	"github.com/open-falcon/falcon-ng/src/dataobj"
)

// Strategy 一条唯一的策略
type Strategy struct {
	ID         int64               `json:"id"`          // 报警策略ID, 唯一
	Name       string              `json:"name"`        // 报警策略名
	Priority   int                 `json:"priority"`    // 报警等级
	Category   int                 `json:"category"`    // 模型, 1:阈值 2:模型
	Operator   string              `json:"operator"`    // 多个Judgements的逻辑运算符
	Judgements []StrategyJudgement `json:"judgements"`  // 支持不同指标的与
	Alert      StrategyAlert       `json:"alert"`       // 触发报警(恢复)的条件配置
	Action     StrategyAction      `json:"action"`      // 报警后的动作
	Partition  string              `json:"partition"`   // 事件推送的分片信息, 分布式alarm场景必须, 由config模块设置
	WindowSize int                 `json:"window_size"` // 数据断点时允许查询的最小起始时间
	Endpoints  []string            `json:"endpoints"`   // 策略关联的endpoints
	Updated    int64               `json:"updated"`     // 策略最后修改时间, 和同步/异步更新有关
}

// StrategyJudgement 对应的指标列表(描述文档)和策略配置
type StrategyJudgement struct {
	Metric    string              `json:"metric"`    // 指标名
	Tags      []StrategyTagFilter `json:"tags"`      // tags过滤规则
	Execution StrategyExecution   `json:"execution"` // 指标对应的判断表达式
}

// StrategyTagFilter 指标描述文档的一部分, tags过滤相关
type StrategyTagFilter struct {
	Tagk     string   `json:"tagk"`     // tagk
	Operator string   `json:"operator"` // 目前只支持 = 和 !=
	Tagv     []string `json:"tagv"`     // tagv
}

// action, StrategyExecution 策略配置, 包含多个表达式
type StrategyExecution struct {
	EffectiveDay   []int                `json:"effective_day"`   // 报警生效日, 星期
	EffectiveStart int                  `json:"effective_start"` // 报警生效时, 一天1440分钟, 第N分钟 开始
	EffectiveEnd   int                  `json:"effective_end"`   // 报警生效时, 第N分结束
	Operator       string               `json:"operator"`        // 多个Expression的逻辑运算, 支持 与、或
	Expressions    []StrategyExpression `json:"expressions"`     // 支持同指标的与和或
}

// StrategyExpression 策略表达式, 唯一对应一个算子
type StrategyExpression struct {
	Func       string              `json:"func"`       // 算子, 目前有10个
	Params     []string            `json:"params"`     // 不同算子的参数不同
	Operator   string              `json:"operator"`   // 多个阈值的逻辑运算, 支持 与、或
	Thresholds []StrategyThreshold `json:"thresholds"` // 阈值
}

// StrategyThreshold 策略阈值
type StrategyThreshold struct {
	Threshold float64 `json:"threshold"` // 阈值
	Operator  string  `json:"operator"`  // 判断表达式, > < = != >= <= etc.
}

// StrategyAlert 报警(解除)判断条件
type StrategyAlert struct {
	AlertCountThreshold      int `json:"alert_count"`      // 按次数判断的报警
	RecoverCountThreshold    int `json:"recover_count"`    // 按次数判断的解除
	AlertDurationThreshold   int `json:"alert_duration"`   // 按秒判断的报警
	RecoverDurationThreshold int `json:"recover_duration"` // 按秒判断的解除
	LimitCountThreshold      int `json:"limit_count"`      // 避免报警轰炸, judge内部生效, xx秒最多报警xx次, 次数
	LimitDurationThreshold   int `json:"limit_duration"`   // 避免报警轰炸, xx秒最多报警xx次, 时间窗口
}

// StrategyAction 报警行为
type StrategyAction struct {
	Groups          []int64  `json:"groups"`           // 报警组
	Callbacks       []string `json:"callbacks"`        // 报警回调地址
	SilenceExtended bool     `json:"silence_extended"` // 屏蔽规则是否对callback动作上生效
	NotifyMode      int      `json:"notify_mode"`      // 通知类型 0:不发送通知 1:发送告警和恢复通知 2:仅发送告警通知
}

// Event 传递到alarm的结构体, 尽可能少的字段, 发出通知需要的信息由alarm自己补全
type Event struct {
	Sid       int64     `json:"sid"`
	EventType string    `json:"event_type"` // alert/recover
	Hashid    uint64    `json:"hashid"`     // 全局唯一 根据counter计算
	Etime     int64     `json:"etime"`
	Endpoint  string    `json:"endpoint"`
	History   []History `json:"-"`
	Detail    string    `json:"detail"`
	Info      string    `json:"info"`
	Value     string    `json:"value"`
	Partition string    `json:"-"`
}

type History struct {
	Key         string             `json:"-"`              // 用于计算event的hashid, 不能用曲线ID, 曲线ID可能会变
	Metric      string             `json:"metric"`         // 指标名
	Tags        map[string]string  `json:"tags,omitempty"` // endpoint/counter
	Granularity int                `json:"-"`              // alarm补齐数据时需要
	Points      []*dataobj.RRDData `json:"points"`         // 现场值
}
