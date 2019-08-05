package entity

import (
	"sync"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/bitmap"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/publish"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/trigger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/buffer"

	nsema "github.com/toolkits/concurrent/semaphore"
)

// 策略实体、判断实体、执行实体的定义

// StrategyEntity 策略实体, 对应唯一的一个报警策略
type StrategyEntity struct {
	sync.RWMutex                            // 读写锁
	*schema.Strategy                        // 策略详情
	cache            *StrategyEntity        // 缓冲区, 指针, 使用后置空
	stop             chan struct{}          // 接受stop命令
	status           int                    // 状态
	indexing         bool                   // 是否正在执行曲线更新
	interval         int                    // 调度周期, 与指标的最小step有关
	indexInterval    int                    // 执行曲线更新的频率
	publisher        publish.EventPublisher // 报警事件推送到某个目的地
	storage          storage.Storage        // 数据缓存, 全局唯一
	concurrency      *nsema.Semaphore       // judgements的并发数
	seriesCount      int                    // 曲线总数

	Judgements map[uint32]*JudgementEntity // key是judgementEntity的唯一标识, 与曲线ID
	Executions []*ExecutionEntity          // 每个trigger关联一个metric, 每个metric关联一堆线
}

// JudgementEntity 判断实体, 对应唯一的"一条曲线"
// 多个指标时, step应该相同, 否则无效, 信任schema层传递的结果
type JudgementEntity struct {
	sid         int64 // 关联的策略ID, 打印日志时使用
	next        int64 // 初始进度
	deadline    int64 // 截止进度
	lastEvent   int   // 上一次报警状态, 减少history结构体的写入
	interval    int   // 每条线维护一个独立的interval
	historySize int   // metricEntity.History 的大小, 与算子和alert参数有关, 与部分算子的参数也有关
	windowSize  int   // 查询数据时允许的最大时间范围, 与 schema.Strategy 的配置有关, windowSize == 0 永远等待

	Metrics []*MetricEntity   // judgement的唯一key
	Driver  AlertDriverEntity // 报警(解除)判断 驱动实体
}

type MetricEntity struct {
	key     string         // key就是metric名, 后续可能会是counter
	ID      uint32         // 唯一索引, 只有一个元素
	History buffer.History // 历史点, 运行时初始化
}

// ExecutionEntity 执行实体
type ExecutionEntity struct {
	sid int64 // 关联的策略ID, 打印日志时使用
	key string

	EffectiveDay    *bitmap.BitMap    // size = 7
	EffectiveMinute *bitmap.BitMap    // size = 1440
	Triggers        []trigger.Trigger // 指标关联trigger, 支持同指标的与、或
	Operator        string            // 支持全局的 and / or
}

// 报警(解除)判断
// alert/recovery的状态记录
type AlertDriverEntity interface {
	Happen(ts int64, status int) (valid bool, changed bool)        // 返回时间戳是否有效, 状态是否发生变化
	DumpEvent(nowTs int64, interval ...int) (code int, clean bool) // EVENT_CODE
	SetThreshold(alert schema.StrategyAlert) error                 // 更新配置
}
