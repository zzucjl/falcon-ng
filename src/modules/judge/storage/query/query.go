package query

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/series"

	"github.com/parnurzeal/gorequest"
)

var (
	ErrorIndexParamIllegal = errors.New("index param illegal")
	ErrorQueryParamIllegal = errors.New("query param illegal")
)

// SeriesQueryManager 查询数据 && 查询索引
type SeriesQueryManager struct {
	opts     SeriesQueryOption
	connPool *SafeRpcConnPools
}

// NewSeriesQueryManager
func NewSeriesQueryManager(opts SeriesQueryOption) (SeriesQueryManager, error) {
	if len(opts.Addrs) == 0 {
		return SeriesQueryManager{}, errors.New("empty query addr")
	}
	if len(opts.IndexAddrs) == 0 {
		return SeriesQueryManager{}, errors.New("empty index addr")
	}

	conn := CreateSafeRpcWithCodecConnPools(opts.MaxConn, opts.MaxIdle,
		opts.ConnTimeout, opts.CallTimeout, opts.Addrs)
	if conn == nil {
		return SeriesQueryManager{}, errors.New("query conn init failed")
	}

	return SeriesQueryManager{
		opts:     opts,
		connPool: conn,
		// client:   client,
	}, nil
}

// 执行Query操作
// 默认不重试, 如果要做重试, 在这里完成
func (mg *SeriesQueryManager) Query(reqs []*dataobj.QueryData) ([]*dataobj.TsdbQueryResponse, error) {
	pool := mg.fetchPool()
	if pool == nil {
		return nil, errors.New("nil conn pool")
	}
	conn, err := pool.Fetch()
	if err != nil {
		return nil, err
	}

	rpcConn := conn.(RpcClient)
	if rpcConn.Closed() {
		pool.ForceClose(conn)
		return nil, errors.New("conn closed")
	}
	var resp *dataobj.QueryDataResp
	err = rpcConn.Call("Transfer.Query", reqs, &resp)
	if err != nil {
		pool.ForceClose(conn)
		return nil, err
	}
	pool.Release(conn)

	if resp.Msg != "" {
		return nil, errors.New(resp.Msg)
	}
	return resp.Data, nil
}

func (mg *SeriesQueryManager) Index(req *storage.IndexRequest) ([]storage.Counter, error) {
	return mg.Xclude(req)
}

func (ms *SeriesQueryManager) fetchPool() *ConnPool {
	i := rand.Intn(len(ms.opts.Addrs))
	addr := ms.opts.Addrs[i]
	cpool, _ := ms.connPool.Get(addr)
	return cpool
}

func (ms *SeriesQueryManager) NewQueryRequest(s *series.Series,
	start, end int64) (*dataobj.QueryData, error) {
	return NewQueryRequest(s, start, end)
}

func NewQueryRequest(s *series.Series,
	start, end int64) (*dataobj.QueryData, error) {
	if end <= start || start < 0 {
		return nil, ErrorQueryParamIllegal
	}
	return &dataobj.QueryData{
		Start:      start,
		End:        end,
		ConsolFunc: "AVERAGE", // 硬编码
		Endpoints:  []string{s.Endpoint()},
		Counters:   []string{s.Counter},
		Step:       s.Granularity,
		DsType:     s.Dstype,
	}, nil
}

func NewIndexRequest(endpoint string, metric string,
	include, exclude map[string][]string) *storage.IndexRequest {
	return &storage.IndexRequest{
		Endpoints: []string{endpoint},
		Metric:    metric,
		Include:   include,
		Exclude:   exclude,
	}
}

/********* 补全索引相关 *********/
type XCludeStruct struct {
	Tagk string   `json:"tagk"`
	Tagv []string `json:"tagv"`
}

type IndexReq struct {
	Endpoints []string       `json:"endpoints"`
	Metric    string         `json:"metric"`
	Include   []XCludeStruct `json:"include,omitempty"`
	Exclude   []XCludeStruct `json:"exclude,omitempty"`
}

type IndexResp struct {
	Data []struct {
		Endpoints []string `json:"endpoints"`
		Metric    string   `json:"metric"`
		Tags      []string `json:"tags"`
		Step      int      `json:"step"`
		Dstype    string   `json:"dstype"`
	} `json:"dat"`
	Err string `json:"err"`
}

// index的xclude 不支持批量查询, 暂时不做
// 默认不重试
func (mg *SeriesQueryManager) Xclude(request *storage.IndexRequest) ([]storage.Counter, error) {
	if len(mg.opts.IndexAddrs) == 0 {
		return nil, errors.New("empty index addr")
	}
	req := IndexReq{
		Endpoints: request.Endpoints,
		Metric:    request.Metric,
	}
	for k, v := range request.Include {
		req.Include = append(req.Include, XCludeStruct{
			Tagk: k,
			Tagv: v,
		})
	}

	for k, v := range request.Exclude {
		req.Exclude = append(req.Exclude, XCludeStruct{
			Tagk: k,
			Tagv: v,
		})
	}

	var (
		result IndexResp
		succ   bool = false
	)
	perm := rand.Perm(len(mg.opts.IndexAddrs))
	for i := range perm {
		resp, body, errs := gorequest.New().
			Timeout(time.Duration(mg.opts.IndexCallTimeout) * time.Millisecond).
			Post(mg.opts.IndexAddrs[perm[i]]).
			Type("json").
			Send([]IndexReq{req}).
			End()

		var code int
		if len(errs) != 0 {
			logger.Debugf(0, "index xclude failed, error:%v, req:%v", errs, req)
			continue
		}

		if resp != nil {
			code = resp.StatusCode
		}
		if code != http.StatusOK {
			logger.Debugf(0, "index xclude failed, code:%d, body:%s, error:%v, req:%v",
				code, body, errs, req)
			continue
		}

		err := json.Unmarshal([]byte(body), &result)
		if err == nil {
			succ = true
			break
		}
		logger.Debugf(0, "index xclude failed, error:%v", err)
	}

	if !succ {
		return nil, errors.New("index xclude failed")
	}

	if len(result.Data) == 1 && len(result.Data[0].Endpoints) == 1 &&
		result.Data[0].Step > 0 {
		var (
			ret      []storage.Counter
			endpoint string = result.Data[0].Endpoints[0]
		)
		// 新接口counter字段不包含 endpoint, 补齐
		if len(result.Data[0].Tags) == 0 {
			ret = append(ret, storage.Counter{
				Counter: schema.ENDPOINT_KEYWORD + "=" + endpoint,
				Step:    result.Data[0].Step,
				Dstype:  result.Data[0].Dstype,
			})
			return ret, nil
		}
		// 新接口counter字段不包含 endpoint, 补齐
		for _, tag := range result.Data[0].Tags {
			if len(tag) > 0 {
				ret = append(ret, storage.Counter{
					Counter: schema.ENDPOINT_KEYWORD + "=" + endpoint + "," + tag,
					Step:    result.Data[0].Step,
					Dstype:  result.Data[0].Dstype,
				})
			}
		}
		return ret, nil
	}

	return nil, nil
}
