package shard

import (
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/series"
)

type shardMapElementEqual func(*series.Series, *series.Series) bool

// shardMap without lock
type shardMap struct {
	lookup map[shardHash]shardMapEntry
	hash   shardHashFunc // 从上层继承, 清理过期指标时使用
	equal  shardMapElementEqual
}

type shardMapEntry struct {
	key   shardHash
	value *series.Series
}

func newShardMap(fn shardHashFunc) *shardMap {
	return &shardMap{
		lookup: make(map[shardHash]shardMapEntry),
		hash:   fn,
		equal:  defaultShardMapElementEqual(),
	}
}

func (m *shardMap) Get(k shardHash, e *series.Series) (*series.Series, bool) {
	hash := k

	for entry, ok := m.lookup[hash]; ok; entry, ok = m.lookup[hash] {
		if m.equal(entry.value, e) {
			return entry.value, true
		}
		// rehash
		hash++
	}

	var empty *series.Series
	return empty, false
}

func (m *shardMap) Contains(k shardHash, e *series.Series) bool {
	_, ok := m.Get(k, e)
	return ok
}

func (m *shardMap) Set(k shardHash, v *series.Series) {
	hash := k

	for entry, ok := m.lookup[hash]; ok; entry, ok = m.lookup[hash] {
		if m.equal(entry.value, v) {
			m.lookup[hash] = shardMapEntry{
				key:   entry.key,
				value: v,
			}
			return
		}
		// rehash
		hash++
	}

	m.lookup[hash] = shardMapEntry{
		key:   k,
		value: v,
	}
}

func (m *shardMap) Len() int {
	l := len(m.lookup)

	return l
}

func (m *shardMap) Delete(k shardHash, e *series.Series) {
	hash := k

	for entry, ok := m.lookup[hash]; ok; entry, ok = m.lookup[hash] {
		if m.equal(entry.value, e) {
			delete(m.lookup, hash)
			return
		}
		// rehash
		hash++
	}
}

func (m *shardMap) Iter() []*series.Series {
	size := m.Len()
	iters := make([]*series.Series, size)
	i := 0

	for _, e := range m.lookup {
		iters[i] = e.value
	}
	return iters
}

func defaultShardMapElementEqual() shardMapElementEqual {
	return func(a, b *series.Series) bool {
		return a.Equal(b)
	}
}
