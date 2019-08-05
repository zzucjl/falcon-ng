package buffer

import (
	"time"
)

var (
	timeoutUnit = time.Millisecond * 1

	defaultStorageQueryTimeout       = 1500
	defaultStorageEnqueueTimeout     = 200
	defaultStorageDequeueTimeout     = 500
	defaultStorageQueuedQueryTimeout = 2200

	defaultStorageQueryQueueSize   = 10000
	defaultStorageQueryConcurrency = 10
	defaultStorageQueryBatch       = 10
	defaultStorageQueryMergeSize   = 30
	defaultStorageShardsetSize     = 10
	defaultStorageHistorySize      = 5
)

type StorageBufferOption struct {
	QueryTimeout       int `yaml:"queryTimeout"`       // 等待下游返回的时间
	QueryConcurrency   int `yaml:"queryConcurrency"`   // 查询下游的最大并发
	QueryBatch         int `yaml:"queryBatch"`         // 并包大小
	QueryMergeSize     int `yaml:"queryMergeSize"`     // 合并请求的窗口, 时间粒度的倍数
	EnqueueTimeout     int `yaml:"enqueueTimeout"`     // 请求入队列超时
	DequeueTimeout     int `yaml:"dequeueTimeout"`     // 请求出队超时
	QueryQueueSize     int `yaml:"queryQueueSize"`     // 请求队列大小
	QueuedQueryTimeout int `yaml:"queuedQueryTimeout"` // 异步读超时
	ShardsetSize       int `yaml:"shardsetSize"`       // shardset的大小
	HistorySize        int `yaml:"historySize"`        // 数据缓存的点数
}

func Duration(n int) time.Duration {
	return time.Duration(n) * timeoutUnit
}

func NewStorageBufferOption() StorageBufferOption {
	return StorageBufferOption{
		QueryTimeout:       defaultStorageQueryTimeout,
		QueryConcurrency:   defaultStorageQueryConcurrency,
		QueryBatch:         defaultStorageQueryBatch,
		QueryMergeSize:     defaultStorageQueryMergeSize,
		EnqueueTimeout:     defaultStorageEnqueueTimeout,
		DequeueTimeout:     defaultStorageDequeueTimeout,
		QueryQueueSize:     defaultStorageQueryQueueSize,
		QueuedQueryTimeout: defaultStorageQueuedQueryTimeout,
		ShardsetSize:       defaultStorageShardsetSize,
		HistorySize:        defaultStorageHistorySize,
	}
}
