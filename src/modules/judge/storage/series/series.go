package series

import (
	"errors"
	"strconv"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
)

type Series struct {
	ID          uint32            // 全局唯一ID
	Metric      string            // 指标名
	Tags        map[string]string // 曲线详情, 除了指标名 所有的都是tag
	Counter     string            // 请求数据的子字符串,直接拼接好
	Granularity int               // 即step, 请求数据时使用
	Dstype      string            // 保留
}

var (
	ErrorEndpointEmpty = errors.New("empty endpoint")
	ErrorTagEmpty      = errors.New("empty tag")
	ErrorTagIllegal    = errors.New("tag illegal")
	ErrorCounterEmpty  = errors.New("empty counter")
)

func NewSeries(metric string,
	tags map[string]string,
	granularity int, dstype string) (*Series, error) {

	if tags == nil || len(tags) == 0 {
		return nil, ErrorTagEmpty
	}
	if _, found := tags[schema.ENDPOINT_KEYWORD]; !found {
		return nil, ErrorEndpointEmpty
	}

	// 不包含 endpoint
	counter := Map2SortedString(tags)
	if len(counter) == 0 {
		counter = metric
	} else {
		// 使用 "/" 分隔 metric和counter
		counter = metric + "/" + counter
	}

	return &Series{
		Metric:      metric,
		Counter:     counter,
		Tags:        tags,
		Granularity: granularity,
		Dstype:      dstype,
	}, nil
}

// Counter中已经保留了metric信息,此处不再重复
func (s *Series) Key() string {
	return s.Counter + schema.COUNTER_SEPERATOR +
		s.Tags[schema.ENDPOINT_KEYWORD] + schema.COUNTER_SEPERATOR +
		strconv.Itoa(s.Granularity) + schema.COUNTER_SEPERATOR +
		s.Dstype
}

func (s *Series) Copy() *Series {
	tag := make(map[string]string)
	if s.Tags != nil {
		for k, v := range s.Tags {
			tag[k] = v
		}
	}
	return &Series{
		Metric:      s.Metric,
		Counter:     s.Counter,
		Tags:        tag,
		Granularity: s.Granularity,
		Dstype:      s.Dstype,
	}
}

func (s *Series) Equal(other *Series) bool {
	if s.Metric != other.Metric {
		return false
	}
	if s.Granularity != other.Granularity {
		return false
	}
	if s.Dstype != other.Dstype {
		return false
	}
	if len(s.Tags) != len(other.Tags) {
		return false
	}
	if s.Counter != other.Counter {
		return false
	}
	if len(s.Tags) > 0 {
		if s.Tags[schema.ENDPOINT_KEYWORD] !=
			other.Tags[schema.ENDPOINT_KEYWORD] {
			return false
		}
	}
	return true
}

func (s *Series) Endpoint() string {
	return s.Tags[schema.ENDPOINT_KEYWORD]
}
