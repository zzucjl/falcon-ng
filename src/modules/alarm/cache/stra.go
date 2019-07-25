package cache

import (
	"sync"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

type StraCacheMap struct {
	sync.RWMutex
	Data map[int64]*dataobj.Stra
}

var StraCache *StraCacheMap

func NewStraCache() *StraCacheMap {
	return &StraCacheMap{
		Data: make(map[int64]*dataobj.Stra),
	}
}

func (this *StraCacheMap) SetAll(m map[int64]*dataobj.Stra) {
	this.Lock()
	defer this.Unlock()
	this.Data = m
}

func (this *StraCacheMap) GetById(id int64) (*dataobj.Stra, bool) {
	this.RLock()
	defer this.RUnlock()

	value, exists := this.Data[id]

	return value, exists
}
