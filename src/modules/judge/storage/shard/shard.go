package shard

import (
	"sync"

	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/series"
)

// Shard 的主要数据结构是shardMap
type Shard struct {
	sync.RWMutex
	lookup *shardMap
}

func NewShard(fn shardHashFunc) *Shard {
	return &Shard{
		lookup: newShardMap(fn),
	}
}

// ContainsWithRLock 判断某条曲线是否在shard中
func (s *Shard) ContainsWithRLock(k shardHash, e *series.Series) bool {
	return s.lookup.Contains(k, e)
}

// GetWithRLock 获取某条曲线的详情, 主要是ID
func (s *Shard) GetWithRLock(k shardHash, e *series.Series) (*series.Series, bool) {
	element, found := s.lookup.Get(k, e)

	if !found {
		return nil, false
	}
	return element, true
}

// SetWithLock 更新/新增series的元信息
func (s *Shard) SetWithLock(k shardHash, e *series.Series) {
	s.lookup.Set(k, e)
}
