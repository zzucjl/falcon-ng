package buffer

import (
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/query"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/series"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/shard"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	nsema "github.com/toolkits/concurrent/semaphore"
)

// StorageBuffer 全局缓存,  包括ID与曲线详情的对应关系和"最新的"数据
type StorageBuffer struct {
	opts             StorageBufferOption
	storage          ConcurrentMap   // 以ID为key的内部数据库
	shardset         *shard.ShardSet // 其实是 ID生成器
	queue            chan *QueryRequest
	query            query.SeriesQueryManager
	queryConcurrency *nsema.Semaphore
}

type SeriesBuffer struct {
	*series.Series
	data []History
}

// NewStorageBuffer
func NewStorageBuffer(opts StorageBufferOption,
	qm query.SeriesQueryManager) *StorageBuffer {

	buffer := &StorageBuffer{
		opts:             opts,
		storage:          NewConcurrentMap(),
		shardset:         shard.NewShardSet(opts.ShardsetSize),
		queue:            make(chan *QueryRequest, opts.QueryQueueSize),
		query:            qm,
		queryConcurrency: nsema.NewSemaphore(opts.QueryConcurrency),
	}
	go buffer.loop()
	// TODO: buffer 定期执行cleanup
	return buffer
}

func (b *StorageBuffer) Query(ID uint32, start, end int64,
	span int) ([]*dataobj.RRDData, error) {

	resp := make(chan *QueryResponse, 1)
	go b.queuedQuery(ID, start, end, span, resp)
	select {
	case ret := <-resp:
		return ret.Data, ret.Error
	case <-time.After(Duration(b.opts.QueuedQueryTimeout)):
		return []*dataobj.RRDData{}, ErrorQueryTimeout
	}
}

func (b *StorageBuffer) Index(req *storage.IndexRequest) ([]storage.Counter, error) {
	return b.query.Index(req)
}

func (b *StorageBuffer) GenerateAndSet(s *series.Series,
	size int, spans []int) uint32 {
	// 生成ID, 如果没有
	b.shardset.Put(s)
	// 更新缓存信息
	b.Set(s, size, spans)

	return s.ID
}

func (b *StorageBuffer) Get(ID uint32) (*series.Series, bool) {
	buffer, found := b.lookup(ID)
	if !found {
		return nil, false
	}
	return buffer.Series, true
}

// TODO: 对缓存进行清理, seriesBuffer需要一个atime参数
// opts里需要有一个过期清理的配置
func (b *StorageBuffer) Cleanup() {

}

func (b *StorageBuffer) Set(s *series.Series, size int, spans []int) {
	if s.ID == 0 {
		return
	}
	if len(spans) == 0 {
		return
	}

	buffer, found := b.lookup(s.ID)
	if !found {
		buffer = NewSeriesBuffer(s, size, spans)
		b.storage.Set(s.ID, buffer)
		return
	}

	buffer.Series = s

	if size < b.opts.HistorySize {
		size = b.opts.HistorySize
	}

	old := make(map[int]struct{})
	for i := range buffer.data {
		buffer.data[i].SetSize(size)
		old[buffer.data[i].ID()] = struct{}{}
	}

	for i := range spans {
		if _, found := old[spans[i]]; !found {
			buffer.data = append(buffer.data, NewRingHistory(size, spans[i], buffer.Granularity()))
		}
	}
}

func NewSeriesBuffer(s *series.Series, size int,
	spans []int) *SeriesBuffer {
	var data []History
	for i := range spans {
		data = append(data, NewRingHistory(size, spans[i], s.Granularity))
	}
	return &SeriesBuffer{
		Series: s,
		data:   data,
	}
}

func (sf *SeriesBuffer) Granularity() int {
	return sf.Series.Granularity
}
