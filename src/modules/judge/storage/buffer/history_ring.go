package buffer

import (
	"math"
	"sort"
	"sync"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

// RingHistory做了很多复杂的事情, ChainHistory相对简单, 性能更好一些
// History是一个环形结构, 不用主动清理数据
// 只支持时间向前新增, 不支持插入的数据, 历史数据不允许更新,
// 不保留null点
type RingHistory struct {
	sync.RWMutex
	span        int               // 时间跨度
	granularity int               // 时间粒度, 用于对齐时间戳
	data        []dataobj.RRDData // 不使用指针, 没有必要保留引用, 更新即可
	size        int
	start       int // 环内第一个有效元素
	end         int // 环内最后一个有效元素, 对应范围 [start, end], 不是 [start, end)
}

// NewHistory 返回一个数据缓存结构体
func NewRingHistory(size int, span int, granularity int) *RingHistory {
	data := make([]dataobj.RRDData, size)
	for i := 0; i < size; i++ {
		data[i] = dataobj.RRDData{}
	}
	return &RingHistory{
		span:        span,
		size:        size,
		granularity: granularity,
		data:        data,
		start:       -1,
		end:         -1,
	}
}

func (h *RingHistory) ID() int {
	return h.span
}

func (h *RingHistory) Size() int {
	return h.size
}

// 扩容, 清空现有的数据
func (h *RingHistory) SetSize(size int) {
	if h.size >= size {
		return
	}

	data := make([]dataobj.RRDData, size)
	for i := 0; i < size; i++ {
		data[i] = dataobj.RRDData{}
	}

	h.Lock()
	h.size = size
	h.start = -1
	h.end = -1
	h.data = data
	h.Unlock()
}

func (h *RingHistory) Dump() []*dataobj.RRDData {
	return nil
}

func (h *RingHistory) Cleanup() {
}

func (h *RingHistory) Reset() {
}

// 返回最新的一个点
func (h *RingHistory) Last() *dataobj.RRDData {
	h.RLock()
	defer h.RUnlock()
	if h.start == -1 {
		return nil
	}
	return &dataobj.RRDData{
		Timestamp: h.data[h.end].Timestamp,
		Value:     h.data[h.end].Value,
	}
}

// Read 尽力而为, 有几个返回几个, 不关心预期的个数
func (h *RingHistory) Read(start, end int64) []*dataobj.RRDData {
	h.RLock()
	defer h.RUnlock()

	ret := make([]*dataobj.RRDData, 0)
	if h.start == -1 {
		return ret
	}
	for i := h.start; i%h.size != h.end; i++ {
		if h.data[i%h.size].Timestamp >= start &&
			h.data[i%h.size].Timestamp <= end {
			// TODO: 由于频繁操作, 考虑引入对象池
			ret = append(ret, &dataobj.RRDData{
				Timestamp: h.data[i%h.size].Timestamp,
				Value:     h.data[i%h.size].Value,
			})
		}
	}
	if h.data[h.end].Timestamp >= start &&
		h.data[h.end].Timestamp <= end {
		ret = append(ret, &dataobj.RRDData{
			Timestamp: h.data[h.end].Timestamp,
			Value:     h.data[h.end].Value,
		})
	}

	return ret
}

// Write 尽力而为, 有多少写多少, 超出的部分直接覆盖最老的数据, 重复的忽略掉
func (h *RingHistory) Write(points []*dataobj.RRDData) {
	if len(points) == 0 {
		return
	}
	ps := make(Points, 0, len(points))
	for _, p := range points {
		if p == nil {
			continue
		}
		if math.IsNaN(float64(p.Value)) ||
			math.IsInf(float64(p.Value), 0) {
			continue
		}
		ps = append(ps, p)
	}
	if len(ps) > 0 {
		sort.Sort(ps)
		h.Lock()
		// 批次最新的点时间戳已经存在了, 直接返回, 该批次不用处理
		if h.start != -1 {
			last := ps[len(ps)-1]
			if h.data[h.end].Timestamp == last.Timestamp {
				// TODO: 这里可以改为 更新最新的点, 历史点不更新
				h.Unlock()
				return
			}
		}
		for i := range ps {
			if ps[i].Timestamp == 0 {
				continue
			}
			pos := h.indexOf(ps[i].Timestamp)
			if pos != -1 {
				// 第一次写入
				if h.start == -1 {
					h.start = 0
					h.end = 0
					h.data[pos].Timestamp = ps[i].Timestamp
					h.data[pos].Value = ps[i].Value
					continue
				}
				// 步进, 向end后面新增
				if pos == (h.end+1)%h.size {
					// 覆盖了旧的start(环写满了)
					if (h.end+1)%h.size == h.start {
						h.start = (h.start + 1) % h.size
					}
					h.end = (h.end + 1) % h.size
					// 插入新数据
					h.data[pos].Timestamp = ps[i].Timestamp
					h.data[pos].Value = ps[i].Value
				}
			}
		}
		h.Unlock()
	}
}

// 用于定位ts的点写入缓存的位置
// TODO: 后续改成 二分查找, 提高效率
func (h *RingHistory) indexOf(ts int64) int {
	if h.start == -1 {
		return 0
	}
	if h.data[h.end].Timestamp == ts {
		return h.end
	}
	if h.data[h.end].Timestamp < ts {
		return (h.end + 1) % h.size
	}
	for i := h.start; i%h.size != h.end; i++ {
		if h.data[i%h.size].Timestamp == ts {
			return i % h.size
		}
		if h.data[i%h.size].Timestamp < ts {
			continue
		}
		if h.data[i%h.size].Timestamp > ts {
			return -1 // 不允许插入操作
		}
	}

	return -1
}
