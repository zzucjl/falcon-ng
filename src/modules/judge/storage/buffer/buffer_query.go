package buffer

import (
	"errors"
	"runtime"
	"sort"
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/series"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

var (
	ErrorIndexNotFound   = errors.New("id not found")           // fatal
	ErrorHistoryEmpty    = errors.New("history empty")          // fatal
	ErrorSpanNotMatch    = errors.New("no history match")       // fatal
	ErrorQueryTimeout    = errors.New("data query timeout")     // warning
	ErrorPointsNotEnough = errors.New("not enough points")      // warning
	ErrorPamraError      = errors.New("start greater than end") // warning
)

type QueryResponse struct {
	Data  []*dataobj.RRDData
	Error error
}

type QueryRequest struct {
	*dataobj.QueryData
	ID          uint32        // 唯一ID
	granularity int           // 时间粒度, 用于计算 合并窗口
	done        chan struct{} // channel, 用于向上传递
	finish      bool          // 用于内部记录进度
	Data        []*dataobj.RRDData
}

func (b *StorageBuffer) lookup(ID uint32) (*SeriesBuffer, bool) {
	sf, found := b.storage.Get(ID)
	if !found {
		return nil, false
	}
	buffer, ok := sf.(*SeriesBuffer)
	if ok {
		return buffer, true
	}
	return nil, false
}

// Query 查询固定范围的所有点
func (b *StorageBuffer) queuedQuery(ID uint32, start, end int64,
	span int, resp chan *QueryResponse) {
	// 0.start向前对齐, end向后对齐
	// 1.从本地history查询, 将能查到的部分填充到结果里
	// 2.如果足够直接返回给上一层; 否则 创建一个query request, 等待执行
	// 3.执行完成后 将结果回写到history中; 如果执行超时, 返回上一层
	// 4.重新从history中读, 主要是history中实现了点的重新填充, 不用再实现一遍
	ret := b.newQueryResponse()
	if start > end {
		ret.Error = ErrorPamraError
		resp <- ret
		return
	}

	buffer, ok := b.lookup(ID)
	if !ok {
		// fatal error
		ret.Error = ErrorIndexNotFound
		resp <- ret
		return
	}
	if len(buffer.data) == 0 {
		ret.Error = ErrorHistoryEmpty
		resp <- ret
		return
	}
	var history History
	for i := range buffer.data {
		if buffer.data[i].ID() == span {
			history = buffer.data[i]
		}
	}
	if history == nil {
		ret.Error = ErrorSpanNotMatch
		resp <- ret
		return
	}

	// 需要从query查询的时间范围
	points := history.Read(start, end)
	size := len(points)
	queryStart := start
	queryEnd := end
	tooEarly := false // start是否早于history最小的时间戳

	if len(points) == 0 {
		// 还有一种情况, 这一段就是null
		tooEarly = true
	}
	if len(points) > 0 {
		if points[0].Timestamp == start &&
			points[size-1].Timestamp == end {
			ret.Data = points
			resp <- ret
			return
		}
		// start早于history, 历史的点不够, 认为一个都没有, start~end 都需要重新查询
		if points[0].Timestamp > start {
			tooEarly = true
		}
		// 最新的点不够, start~end都需要重新查询
	}

	// query不要解决边界问题, 由上游来处理
	req, err := b.newQueryRequest(buffer.Series, queryStart, queryEnd)
	if err != nil {
		ret.Error = err
		resp <- ret
		return
	}

	select {
	case b.queue <- req:
		// do nothing
	case <-time.After(Duration(b.opts.EnqueueTimeout)):
		// 入队列超时, 向上返回 queryTimeout
		ret.Error = ErrorQueryTimeout
		resp <- ret
		return
	}

	select {
	case <-req.done:
		if len(req.Data) == 0 {
			// 查询结果为空, 直接返回
			// 屏蔽掉 查询结果为空的错误
			// ret.Error = ErrorPointsNotEnough
			ret.Data = req.Data
			resp <- ret
			return
		}
		// 如果start小于history[0] 直接返回重新查询的点
		if tooEarly {
			points = make([]*dataobj.RRDData, 0)
			for i := range req.Data {
				if req.Data[i].Timestamp >= start &&
					req.Data[i].Timestamp <= end {
					points = append(points, req.Data[i])
				}
			}
			ret.Data = points
			resp <- ret
		}

		// 更新缓存内容
		history.Write(req.Data)

		if !tooEarly {
			// 重新填充points结构, 并返回(补上最新的点)
			// 这里复用了history中对点排序的逻辑
			points = history.Read(start, end)
			ret.Data = points
			resp <- ret
		}
		return

	case <-time.After(Duration(b.opts.QueryTimeout)):
		ret.Error = ErrorQueryTimeout
		resp <- ret
		return
	}
}

// 执行queues中的任务
func (b *StorageBuffer) loop() {
	for {
		requests := make([]*QueryRequest, 0, b.opts.QueryBatch)
		timeout := false
		for {
			select {
			case req := <-b.queue:
				requests = append(requests, req)
				if len(requests) >= b.opts.QueryBatch {
					timeout = true
				}
			case <-time.After(Duration(b.opts.DequeueTimeout)):
				timeout = true
			}
			if timeout {
				break
			}
		}

		if len(requests) > 0 {
			b.queryConcurrency.Acquire()
			go func(requests []*QueryRequest) {
				defer func() {
					if r := recover(); r != nil {
						var buf [8192]byte
						n := runtime.Stack(buf[:], false)
						logger.Errorf(0, "storage query panic recover: -> %v\n%s", r, buf[:n])
					}
				}()
				defer b.queryConcurrency.Release()

				// 对req进行分组, 同一个ID的req根据时间跨度进行合并
				groups := make([]*RequestGroup, 1)
				group0 := newRequestGroup()
				groups[0] = group0
				for _, req := range requests {
					if req == nil {
						continue
					}
					doing := false
					for _, group := range groups {
						if group.Merge(req, b.opts.QueryMergeSize) {
							doing = true
							break
						}
					}
					if !doing {
						group1 := newRequestGroup()
						if !group1.Merge(req, b.opts.QueryMergeSize) {
							// 严重错误
							logger.Errorf(0, "storage query loop, merge failed, ID:%d", req.ID)
						}
						groups = append(groups, group1)
					}
				}
				for _, group := range groups {
					querys := make([]*dataobj.QueryData, 0, group.Size())
					for _, mr := range group.request {
						query := mr.deriveQueryRequest()
						if query == nil {
							logger.Warningf(0, "storage query loop, nil query pointer")
							continue
						}
						querys = append(querys, query)
					}
					if len(querys) > 0 {
						resps, err := b.query.Query(querys)
						if err != nil {
							logger.Warningf(0, "storage query downstream failed: %v", err)
							continue
						}
						for _, resp := range resps {
							if resp == nil {
								continue
							}
							ID, found := group.GetID(resp.Key())
							if !found {
								// 严重错误
								logger.Errorf(0, "storage query loop, key mapping ID failed")
								continue
							}
							// 向上返回 数据
							group.finalize(ID, []*dataobj.RRDData(resp.Values))
						}
					}
					// 最后对可能遗漏的部分都执行一次finalize
					// 比如 下游报错, 或返回的resp解析出错
					for _, mr := range group.request {
						mr.finalize()
					}
				}
			}(requests)
		}
	}
}

// map结构不存在同时读/写, 可以不加锁
type RequestGroup struct {
	request map[uint32]*MergeRequest
	lookup  map[string]uint32 // 用于反解resp中对应的曲线ID, key 对应请求query的counter, 要求唯一
}

// MergeRequest 合并同一个ID的请求, (start,end) 包含了requests中要求的所有时间范围
type MergeRequest struct {
	requests    []*QueryRequest
	start       int64
	end         int64
	granularity int // 用于判断合并窗口, 合并窗口配置项是 windowByGrl, 即N*grl才是窗口大小
}

func newRequestGroup() *RequestGroup {
	return &RequestGroup{
		request: make(map[uint32]*MergeRequest),
		lookup:  make(map[string]uint32),
	}
}

func newMergeRequest() *MergeRequest {
	return &MergeRequest{
		requests:    make([]*QueryRequest, 0),
		start:       0,
		end:         0,
		granularity: 0,
	}
}

func (g *RequestGroup) Size() int {
	return len(g.request)
}

func (g *RequestGroup) Merge(req *QueryRequest, windowByGrl int) bool {
	mr, found := g.request[req.ID]
	if !found {
		mr = newMergeRequest()
		if mr.Merge(req, windowByGrl) {
			g.request[req.ID] = mr
			g.lookup[req.Key()] = req.ID
			return true
		}
		return false
	}

	return mr.Merge(req, windowByGrl)
}

func (g *RequestGroup) GetID(key string) (uint32, bool) {
	ID, found := g.lookup[key]
	return ID, found
}

func (g *RequestGroup) finalize(ID uint32, points ...[]*dataobj.RRDData) {
	if mr, found := g.request[ID]; found {
		mr.finalize(points...)
	}
}

func (m *MergeRequest) Merge(req *QueryRequest, windowByGrl int) bool {
	if len(m.requests) == 0 {
		m.requests = append(m.requests, req)
		m.start = req.Start
		m.end = req.End
		m.granularity = req.granularity
		return true
	}

	// 判断 start/end 是否能合并
	window := int64(windowByGrl * m.granularity)
	if req.End-m.start <= window &&
		m.end-req.Start <= window &&
		req.End-req.Start <= window {
		m.requests = append(m.requests, req)
		if req.Start < m.start {
			m.start = req.Start
		}
		if req.End > m.end {
			m.end = req.End
		}
		return true
	}
	return false
}

// 多个请求合并成一个, 以最大的start~end 执行一次查询
func (m *MergeRequest) deriveQueryRequest() *dataobj.QueryData {
	if len(m.requests) == 0 {
		return nil
	}
	return &dataobj.QueryData{
		Start:      m.start,
		End:        m.end,
		ConsolFunc: m.requests[0].ConsolFunc,
		Endpoints:  m.requests[0].Endpoints,
		Counters:   m.requests[0].Counters,
		Step:       m.requests[0].Step,
		DsType:     m.requests[0].DsType,
	}
}

func (m *MergeRequest) finalize(points ...[]*dataobj.RRDData) {
	data := make([]*dataobj.RRDData, 0)
	if len(points) > 0 {
		data = points[0]
	}
	for i := range m.requests {
		if !m.requests[i].finish {
			m.requests[i].Data = data
			m.requests[i].finish = true
			m.requests[i].done <- struct{}{}
		}
	}
}

func isIntSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Ints(a)
	sort.Ints(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (b *StorageBuffer) newQueryResponse() *QueryResponse {
	return &QueryResponse{
		Data: make([]*dataobj.RRDData, 0),
	}
}

func (b *StorageBuffer) newQueryRequest(
	s *series.Series, start, end int64) (*QueryRequest, error) {
	req, err := b.query.NewQueryRequest(s, start, end)
	if err != nil {
		return nil, err
	}

	return &QueryRequest{
		QueryData:   req,
		ID:          s.ID,
		granularity: s.Granularity,
		done:        make(chan struct{}, 1),
		finish:      false,
		Data:        make([]*dataobj.RRDData, 0),
	}, nil
}
