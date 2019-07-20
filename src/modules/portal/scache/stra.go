package scache

import (
	"sync"

	"github.com/open-falcon/falcon-ng/src/model"
)

type StraCacheMap struct {
	sync.RWMutex
	Data map[string][]*model.Stra
}

var StraCache *StraCacheMap

func NewStraCache() *StraCacheMap {
	return &StraCacheMap{Data: make(map[string][]*model.Stra)}
}

func (s *StraCacheMap) GetByNode(node string) []*model.Stra {
	s.RLock()
	defer s.RUnlock()

	return s.Data[node]
}

func (s *StraCacheMap) Set(node string, stras []*model.Stra) {
	s.Lock()
	defer s.Unlock()

	s.Data[node] = stras
	return
}

func (s *StraCacheMap) SetAll(strasMap map[string][]*model.Stra) {
	s.Lock()
	defer s.Unlock()

	s.Data = strasMap
	return
}
