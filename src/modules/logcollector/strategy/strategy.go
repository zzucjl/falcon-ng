package strategy

import (
	"github.com/open-falcon/falcon-ng/src/modules/logcollector/config"
	"github.com/open-falcon/falcon-ng/src/modules/logcollector/schema"

	"fmt"
)

// 后续开发者切记 : 没有锁，不能修改globalStrategy，更新的时候直接替换，否则会panic
var (
	globalStrategy map[int64]*schema.Strategy
)

func init() {
	globalStrategy = make(map[int64]*schema.Strategy, 0)
}

func UpdateGlobalStrategy(sts []*schema.Strategy) error {
	tmpStrategyMap := make(map[int64]*schema.Strategy, 0)
	for _, st := range sts {
		if st.Degree == 0 {
			st.Degree = int64(config.Config.Strategy.DefaultDegree)
		}
		tmpStrategyMap[st.ID] = st
	}
	globalStrategy = tmpStrategyMap
	return nil
}

func GetListAll() []*schema.Strategy {
	stmap := GetDeepCopyAll()
	var ret []*schema.Strategy
	for _, st := range stmap {
		ret = append(ret, st)
	}
	return ret
}

func GetDeepCopyAll() map[int64]*schema.Strategy {
	ret := make(map[int64]*schema.Strategy, len(globalStrategy))
	for k, v := range globalStrategy {
		ret[k] = schema.DeepCopyStrategy(v)
	}
	return ret
}

func GetAll() map[int64]*schema.Strategy {
	return globalStrategy
}

func GetByID(id int64) (*schema.Strategy, error) {
	st, ok := globalStrategy[id]

	if !ok {
		return nil, fmt.Errorf("ID : %d is not exists in global Cache")
	}
	return st, nil

}
