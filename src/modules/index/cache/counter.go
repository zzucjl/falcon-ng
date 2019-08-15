package cache

import (
	"sync"
)

type CounterStruct struct {
	sync.RWMutex
	Updated int64  `json:"updated"`
	Step    int64  `json:"step"`
	DsType  string `json:"dstype"`
}

func NewCounterStruct(now, step int64, dsType string) *CounterStruct {
	return &CounterStruct{Updated: now, Step: step, DsType: dsType}
}

func (c *CounterStruct) Update(now, step int64, dstype string) {
	c.Lock()
	defer c.Unlock()

	c.Updated = now
	c.Step = step
	c.DsType = dstype
}

func (c *CounterStruct) GetUpdate() int64 {
	c.RLock()
	defer c.RUnlock()

	return c.Updated
}

func (c *CounterStruct) GetInfo() (int64, string) {
	c.RLock()
	defer c.RUnlock()

	return c.Step, c.DsType
}

type CounterStructRet struct {
	Counter string `json:"counter"`
	Step    int    `json:"step"`
	DsType  string `json:"dstype"`
}
