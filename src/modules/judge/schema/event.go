package schema

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/json-iterator/go"
	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/spaolacci/murmur3"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	bufferPool = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
)

func NewEvent(sid int64, partition string, etype string) *Event {
	return &Event{
		Sid:       sid,
		EventType: etype,
		Partition: partition,
		History:   make([]History, 0),
	}
}

func (e *Event) SetInfo(info string) {
	e.Info = info
}

func (e *Event) SetPoints(key string,
	metric string,
	tags map[string]string,
	granularity int,
	points []*dataobj.RRDData) {

	var endpoint string
	ntags := make(map[string]string)
	if len(tags) > 0 {
		for k, v := range tags {
			if k == ENDPOINT_KEYWORD {
				endpoint = v
				continue
			}
			ntags[k] = v
		}
	}
	// 设置时间为 最后一个点的时间戳
	if len(e.History) == 0 && len(points) > 0 {
		e.Etime = points[len(points)-1].Timestamp
	}
	// 设置时间为 最大的一个点的时间戳
	if len(e.History) > 0 &&
		len(points) > 0 &&
		points[len(points)-1].Timestamp > e.Etime {
		e.Etime = points[len(points)-1].Timestamp
	}
	e.Endpoint = endpoint
	history := History{
		Key:         key,
		Metric:      metric,
		Tags:        ntags,
		Granularity: granularity,
		Points:      points,
	}
	e.History = append(e.History, history)
}

// 计算hashid, 要求 history 的写入顺序是有序的,
func (e *Event) Finalize() bool {
	if len(e.History) == 0 {
		return false
	}
	pk := bufferPool.Get().(*bytes.Buffer)
	pk.Reset()
	defer bufferPool.Put(pk)

	pk.WriteString(strconv.FormatInt(e.Sid, 16))
	pk.WriteByte('/')

	for i := range e.History {
		pk.WriteString(e.History[i].Key)
	}
	hashid := murmur3.Sum64(pk.Bytes())

	//因为xorm不支持uint64，为解决数据溢出的问题，此处将hashid转化为60位
	//具体细节：将64位值 高4位与低60位进行异或操作
	e.Hashid = (hashid >> 60) ^ (hashid & 0xFFFFFFFFFFFFFFF)

	// Detail字段保存为 字符串
	bytes, _ := json.Marshal(e.History)
	e.Detail = string(bytes)

	// 更新Value字段
	var value string
	if len(e.History) == 1 && len(e.History[0].Points) > 0 {
		f := e.History[0].Points[len(e.History[0].Points)-1].Value
		if math.IsNaN(float64(f)) {
			value = "null"
		} else {
			value = fmt.Sprintf("%.2f", f)
		}

	} else {
		for i := range e.History {
			if len(e.History[i].Points) > 0 {
				f := e.History[i].Points[len(e.History[i].Points)-1].Value
				if math.IsNaN(float64(f)) {
					value += e.History[i].Metric + ":null"
				} else {
					value += e.History[i].Metric + ":" + fmt.Sprintf("%.2f", f)
				}
			}
			if i < len(e.History)-1 {
				value += ","
			}
		}
	}
	e.Value = value

	return true
}
