package buffer

import (
	"github.com/open-falcon/falcon-ng/src/dataobj"
)

// 支持插入的数据, 历史数据允许更新
type History interface {
	Dump() []*dataobj.RRDData
	Cleanup()
	Reset()
	Last() *dataobj.RRDData
	Read(start, end int64) []*dataobj.RRDData
	Write(points []*dataobj.RRDData)
	ID() int
	Size() int
	SetSize(size int)
}

type Points []*dataobj.RRDData

func (ps Points) Len() int           { return len(ps) }
func (ps Points) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }
func (ps Points) Less(i, j int) bool { return ps[i].Timestamp < ps[j].Timestamp }
