package buffer

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/didi/nightingale/src/modules/judge/storage/query"
	"github.com/didi/nightingale/src/modules/judge/storage/series"

	"github.com/stretchr/testify/assert"
)

// 必须依赖下游地址做测试, 不能写单测
func Test_Buffer(t *testing.T) {
	query, _ := query.NewSeriesQueryManager(
		query.NewSeriesQueryOption([]string{"127.0.0.1:8041"},
			[]string{"http://127.0.0.1:8030/api/index/counter/clude"}))

	buffer := NewStorageBuffer(NewStorageBufferOption(), query)
	s, _ := series.NewSeries(
		"cpu.idle",
		map[string]string{
			"endpoint": "mock",
		},
		10,
		"GAUGE",
	)
	assert.NotNil(t, s)

	buffer.GenerateAndSet(s, 12, []int{0})
	assert.Equal(t, uint32(1), s.ID)

	buffer.GenerateAndSet(s, 12, []int{0})
	assert.Equal(t, uint32(1), s.ID)

	resp := make(chan *QueryResponse, 1)

	start := time.Now().Unix() - 600
	// 不要查询最新的一个周期, 至少要往前推一个周期
	end := time.Now().Unix()
	size := 0

	go buffer.queuedQuery(1, start, end, 0, resp)
	select {
	case ret := <-resp:
		assert.Nil(t, ret.Error)
		assert.NotEqual(t, 0, len(ret.Data))
		size = len(ret.Data)
		bytes, _ := json.Marshal(ret.Data)
		fmt.Println(string(bytes))

	case <-time.After(time.Second * 2):
	}

	go buffer.queuedQuery(1, start, end, 0, resp)
	select {
	case ret := <-resp:
		assert.Nil(t, ret.Error)
		// maybe will not pass at sometime
		assert.Equal(t, size, len(ret.Data))
		bytes, _ := json.Marshal(ret.Data)
		fmt.Println(string(bytes))
	}
}
