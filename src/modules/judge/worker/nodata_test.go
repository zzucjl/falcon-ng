package worker

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/open-falcon/falcon-ng/src/modules/judge/schema"
	"github.com/open-falcon/falcon-ng/src/modules/judge/schema/entity"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/buffer"
	"github.com/open-falcon/falcon-ng/src/modules/judge/storage/query"
)

func Test_Nodata(t *testing.T) {
	start()
	ss := mockStrategy()
	for j := 0; j < 20; j++ {
		now := time.Now()
		for i := range ss {
			ok, nstra := GenerateNodataStrategy(ss[i], now)
			if ok {
				events := nstra.Run(now)
				if len(events) > 0 {
					bytes, _ := json.Marshal(events)
					fmt.Printf("event: %s\n", string(bytes))
				}
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func mockStrategy() []*schema.Strategy {
	alert := schema.StrategyAlert{
		AlertCountThreshold:      1,
		RecoverDurationThreshold: 100,
		LimitCountThreshold:      1,
		LimitDurationThreshold:   300,
	}

	happen := []schema.StrategyJudgement{
		schema.StrategyJudgement{
			Metric: "cpu.idle",
			Execution: schema.StrategyExecution{
				EffectiveDay:   []int{0, 1, 2, 3, 4, 5, 6},
				EffectiveStart: 0,
				EffectiveEnd:   1439,                      // 默认设置为一直有效
				Operator:       schema.LOGIC_OPERATOR_AND, // 默认是 "and"
				Expressions: []schema.StrategyExpression{
					schema.StrategyExpression{
						Func:     "happen",
						Params:   []string{"300", "1"},
						Operator: schema.LOGIC_OPERATOR_AND, // 默认是 "and"
						Thresholds: []schema.StrategyThreshold{
							schema.StrategyThreshold{
								Operator: "=", Threshold: float64(100)},
						},
					},
				},
			},
		},
	}

	nodata := []schema.StrategyJudgement{
		schema.StrategyJudgement{
			Metric: "cpu.core.idle",
			Tags: []schema.StrategyTagFilter{schema.StrategyTagFilter{
				Tagk:     "core",
				Operator: "=",
				Tagv:     []string{"60", "1"},
			}},
			Execution: schema.StrategyExecution{
				EffectiveDay:   []int{0, 1, 2, 3, 4, 5, 6},
				EffectiveStart: 0,
				EffectiveEnd:   1439,                      // 默认设置为一直有效
				Operator:       schema.LOGIC_OPERATOR_AND, // 默认是 "and"
				Expressions: []schema.StrategyExpression{
					schema.StrategyExpression{
						Func:     "nodata",
						Params:   []string{"100"},
						Operator: schema.LOGIC_OPERATOR_AND, // 默认是 "and"
					},
				},
			},
		},
	}

	return []*schema.Strategy{
		&schema.Strategy{
			ID:         1,
			Name:       "happen测试",
			Priority:   2,
			Category:   1,
			Operator:   schema.LOGIC_OPERATOR_AND,
			WindowSize: 180,
			Nids:       []int64{55},
			Partition:  "/event/p2",
			Judgements: happen,
			Alert:      alert,
		},
		&schema.Strategy{
			ID:         2,
			Name:       "nodata测试",
			Priority:   2,
			Category:   1,
			Operator:   schema.LOGIC_OPERATOR_AND,
			WindowSize: 180,
			Nids:       []int64{55},
			Partition:  "/event/p2",
			Judgements: nodata,
			Alert:      alert,
		},
	}
}

func start() {
	// 初始化存储组件
	qm, err := query.NewSeriesQueryManager(
		query.NewSeriesQueryOption([]string{"127.0.0.1:8041"},
			[]string{"http://127.0.0.1:8030/api/index/counter/clude"}),
	)
	if err != nil {
		log.Fatalln("[F] init transfer/index failed:", err)
	}
	stg = buffer.NewStorageBuffer(buffer.NewStorageBufferOption(), qm)

	// 初始化nodata driver
	_nodataDrivers = make(map[int64]map[string]entity.AlertDriverEntity)
	_options = StrategyConfigOption{
		Addrs:   []string{"127.0.0.1:8022"},
		Timeout: 5000,
	}
}
