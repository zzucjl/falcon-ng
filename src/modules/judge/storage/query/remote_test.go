package query

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage"

	"github.com/stretchr/testify/assert"
)

func Test_Data(t *testing.T) {
	// 线下测试地址
	opts := NewSeriesQueryOption([]string{"127.0.0.1:8041"},
		[]string{"http://127.0.0.1:8030/api/v2/index/counter/clude"})
	manager, _ := NewSeriesQueryManager(opts)

	for {
		// 原始数据
		req1 := &dataobj.QueryData{
			Start:      time.Now().Unix() - 90,
			End:        time.Now().Unix(),
			ConsolFunc: "AVERAGE",
			Endpoints:  []string{"127.0.0.1"},
			Counters:   []string{"cpu.idle"},
			Step:       10,
			DsType:     "GAUGE",
		}
		fmt.Println(time.Now().Unix(), req1.Start, req1.End)

		resp, err := manager.Query([]*dataobj.QueryData{req1})
		assert.Nil(t, err)
		for i := range resp {
			bytes, _ := json.Marshal(resp[i])
			fmt.Println(string(bytes))
		}

		time.Sleep(10 * time.Second)
	}

}

// 必须依赖下游地址做测试, 不能写单测
func Test_Query(t *testing.T) {
	// 线下测试地址
	opts := NewSeriesQueryOption([]string{"127.0.0.1:8041"},
		[]string{"http://127.0.0.1:8030/api/v2/index/counter/clude"})
	manager, _ := NewSeriesQueryManager(opts)

	// 原始数据
	req1 := &dataobj.QueryData{
		Start:      time.Now().Unix() - 600,
		End:        time.Now().Unix(),
		ConsolFunc: "AVERAGE",
		Endpoints:  []string{"mock"},
		Counters:   []string{"cpu.steal"},
		Step:       10,
		DsType:     "GAUGE",
	}
	// 原始数据
	req2 := &dataobj.QueryData{
		Start:      time.Now().Unix() - 600,
		End:        time.Now().Unix(), // 返回2个点
		ConsolFunc: "AVERAGE",
		Endpoints:  []string{"mock"},
		Counters:   []string{"cpu.util"},
		Step:       10,
		DsType:     "GAUGE",
	}

	resps, err := manager.Query([]*dataobj.QueryData{req1, req2})
	assert.Nil(t, err)
	for i := range resps {
		bytes, _ := json.Marshal(resps[i])
		fmt.Printf("values: %v\n", string(bytes))
	}

	include := map[string][]string{}
	exclude := make(map[string][]string)
	counters, err := manager.Xclude(&storage.IndexRequest{
		Endpoints: []string{"mock"},
		Metric:    "cpu.core.idle",
		Include:   include,
		Exclude:   exclude,
	})
	assert.Nil(t, err)
	fmt.Println(counters)
}
