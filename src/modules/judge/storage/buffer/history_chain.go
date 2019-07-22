package buffer

import (
	"github.com/open-falcon/falcon-ng/src/dataobj"
)

// 所有的点 保留一个批次信息, 内部维护一个变量, 每次write的是一个批次
// 上游在使用时, 没有并发的场景, 暂时不需要加锁
type ChainHistory struct {
	head   *chainNode
	tail   *chainNode
	batch  uint32
	size   int
	length int
}

type chainNode struct {
	next  *chainNode
	data  *dataobj.RRDData
	batch uint32
}

func NewChainHistory(size int) *ChainHistory {
	return &ChainHistory{
		head:   nil,
		tail:   nil,
		batch:  0,
		size:   size,
		length: 0,
	}
}

// ID() 方法 暂时没有意义
func (h *ChainHistory) ID() int {
	return 0
}

func (h *ChainHistory) Size() int {
	return h.size
}

func (h *ChainHistory) SetSize(size int) {
	if h.size >= size {
		return
	}
	h.size = size
}

func (h *ChainHistory) Dump() []*dataobj.RRDData {
	if h.head == nil {
		return []*dataobj.RRDData{}
	}
	ret := make([]*dataobj.RRDData, 0)
	for node := h.head; node != nil; node = node.next {
		ret = append(ret, &dataobj.RRDData{
			Timestamp: node.data.Timestamp,
			Value:     node.data.Value,
		})
	}
	return ret
}

// Cleanup 除最新的批次外, 其他全部清空
func (h *ChainHistory) Cleanup() {
	if h.head == nil {
		return
	}
	// 找到新的head
	for node := h.head; node != nil; {
		if node.batch != h.batch {
			h.head = node.next
			node = h.head
			h.length--
			continue
		}
		break
	}
	// 全被删除
	if h.head == nil {
		h.tail = nil
		return
	}
	previous := h.head
	for node := previous.next; node != nil; node = previous.next {
		// 删除中间的点
		if node.batch != h.batch {
			if node == h.tail {
				h.tail = previous
			}
			previous.next = node.next
			h.length--
		} else {
			previous = previous.next
		}
	}
}

func (h *ChainHistory) Reset() {
	h.head = nil
	h.tail = nil
	h.length = 0
}

func (h *ChainHistory) Last() *dataobj.RRDData {
	return nil
}

// TODO: 在现有场景下也没有意义
func (h *ChainHistory) Read(start, end int64) []*dataobj.RRDData {
	return nil
}

// TODO: 考虑对point进行校验, 值为NaN或者无穷大，没有意义
func (h *ChainHistory) Write(points []*dataobj.RRDData) {
	if h.size == 0 {
		return
	}
	if len(points) == 0 {
		return
	}

	h.batch++
	for i := range points {
		if points[i] == nil {
			continue
		}
		if h.head == nil {
			node := newChainNode(h.batch, points[i])
			h.head = node
			h.tail = node
			h.length++
			continue
		}
		pos, insert, forward := h.indexOfInsert(points[i].Timestamp)
		if pos == nil {
			continue
		}

		if insert {
			node := newChainNode(h.batch, points[i])
			if pos == h.tail && forward {
				// append到链表尾
				pos.next = node
				h.tail = node
				h.length++

			} else if pos == h.head && !forward {
				// 追加到链表头
				node.next = pos
				h.head = node
				h.length++

			} else {
				// 在链表中间插入
				node.next = pos.next
				pos.next = node
				h.length++

			}
			// 链超长, 丢掉起始的部分
			if h.length > h.size {
				j := 0
				node := h.head
				for node != nil && j < h.length-h.size {
					node = node.next
					j++
				}
				h.length = h.size
				h.head = node
			}
			continue
		}
		// 更新已有节点
		pos.data.Value = points[i].Value
		pos.batch = h.batch
	}
}

func newChainNode(batch uint32, point *dataobj.RRDData) *chainNode {
	return &chainNode{
		next:  nil,
		data:  point,
		batch: batch,
	}
}

func (h *ChainHistory) indexOfInsert(ts int64) (
	node *chainNode, insert bool, forward bool) {
	if h.head == nil || h.tail == nil {
		return nil, false, false
	}
	// 直接更新tail
	if h.tail.data.Timestamp == ts {
		return h.tail, false, false
	}
	// 向tail后追加
	if h.tail.data.Timestamp < ts {
		return h.tail, true, true
	}
	// 向head前追加
	if h.head.data.Timestamp > ts {
		return h.head, true, false
	}

	previous := h.head
	for node := h.head; node != nil && node.data.Timestamp <= ts; {
		if node.data.Timestamp < ts {
			previous = node
			node = node.next
			continue
		}
		if node.data.Timestamp == ts {
			// 直接更新node
			return node, false, false
		}
	}

	// 向previous后追加
	return previous, true, true
}
