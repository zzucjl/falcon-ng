package shard

import (
	"sync/atomic"

	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/series"

	"github.com/spaolacci/murmur3"
)

type shardHash uint32

type shardHashFunc func(k string) shardHash

// ShardSet 是Shard的集合, 取模计算的分片, 避免单个map过大, 影响效率
type ShardSet struct {
	hash    shardHashFunc  // 计算曲线的hash值
	size    int            // 分片数, 分片逻辑在HashFunc外面做, 确保hash值只需要计算一次
	shards  map[int]*Shard // shard分片, 不需要锁, 不存在map的并发读写
	created uint32         // 唯一索引ID, 自增
}

func NewShardSet(size int, fn ...shardHashFunc) *ShardSet {
	hash := defaultHashFunc()
	if len(fn) > 0 {
		hash = fn[0]
	}

	shards := make(map[int]*Shard)
	for i := 0; i < size; i++ {
		shards[i] = NewShard(hash)
	}
	return &ShardSet{
		hash:    hash,
		size:    size,
		shards:  shards,
		created: 0,
	}
}

func (s *ShardSet) Contains(e *series.Series) bool {
	hash := s.hash(e.Key())
	shard, _ := s.shards[int(hash)%s.size]

	shard.RLock()
	defer shard.RUnlock()
	return shard.ContainsWithRLock(hash, e)
}

// Put 生成(查询)全局唯一ID
func (s *ShardSet) Put(e *series.Series) uint32 {
	// 全局加锁, 避免重复计数
	hash := s.hash(e.Key())

	shard, _ := s.shards[int(hash)%s.size]

	shard.Lock()
	if series, ok := shard.GetWithRLock(hash, e); ok {
		e.ID = series.ID
		shard.Unlock()
		return series.ID
	}

	id := atomic.AddUint32(&s.created, 1)
	e.ID = id
	shard.SetWithLock(hash, e)
	shard.Unlock()

	return id
}

// CleanUp 清理过期的series记录, 避免内存过大, 暂时不实现
func (s *ShardSet) CleanUp() {

}

func (s *ShardSet) Size() int {
	return s.size
}

func (s *ShardSet) Max() uint32 {
	return atomic.LoadUint32(&s.created)
}

func defaultHashFunc() shardHashFunc {
	return func(k string) shardHash {
		return shardHash(murmur3.Sum32([]byte(k)))
	}
}
