package backend

import (
	"bytes"
	"math"
	"sort"

	"github.com/toolkits/pkg/logger"

	"github.com/open-falcon/falcon-ng/src/dataobj"
)

var (
	validFuncName = map[string]struct{}{
		"sum": struct{}{},
		"avg": struct{}{},
		"max": struct{}{},
		"min": struct{}{},
	}
)

type AggrTsValue struct {
	Value dataobj.JsonFloat
	Count int
}

func AggrFuncValide(f string) bool {
	_, ok := validFuncName[f]
	return ok
}

func GetAggrCounter(input dataobj.QueryDataForUI) map[string][]map[string]string {
	groupCounter := make(map[string][]map[string]string)

	for _, endpoint := range input.Endpoints {
		tagIsNull := false
		if len(input.Tags) == 0 { //添加endpoint
			tag := "endpoint=" + endpoint
			input.Tags = append(input.Tags, tag)
			tagIsNull = true
		}

		for _, tag := range input.Tags {
			tagMap, err := dataobj.SplitTagsString(tag)
			if err != nil {
				logger.Warning(err, tag)
				continue
			}

			tagMap["endpoint"] = endpoint
			validCounter := true
			//得到聚合维度的tagkv
			var b bytes.Buffer
			for i, key := range input.GroupKey {
				if v, exists := tagMap[key]; exists {
					b.WriteString(key)
					b.WriteString("=")
					b.WriteString(v)
					if i != len(input.GroupKey)-1 {
						b.WriteString(",")
					}
				} else {
					validCounter = false
					break
				}
			}

			if !validCounter {
				continue
			}
			groupTag := b.String()

			if groupTag == "" {
				groupTag = input.AggrFunc
			}

			if _, exists := groupCounter[groupTag]; exists {
				groupCounter[groupTag] = append(groupCounter[groupTag], tagMap)
			} else {
				groupCounter[groupTag] = []map[string]string{tagMap}
			}
		}
		if tagIsNull {
			input.Tags = []string{}
		}
	}

	return groupCounter
}

func compute(f string, datas []*dataobj.TsdbQueryResponse) []*dataobj.RRDData {
	datasLen := len(datas)
	if datasLen < 1 {
		return nil
	}

	dataMap := make(map[int64]*AggrTsValue)
	switch f {
	case "sum":
		dataMap = sum(datas)
	case "avg":
		dataMap = avg(datas)
	case "max":
		dataMap = max(datas)
	case "min":
		dataMap = min(datas)
	}

	var tmpValues dataobj.RRDValues
	for ts, v := range dataMap {
		tmp := dataobj.RRDData{
			Timestamp: ts,
			Value:     dataobj.JsonFloat(v.Value),
		}
		tmpValues = append(tmpValues, &tmp)
	}
	sort.Sort(tmpValues)
	return tmpValues
}

func sum(datas []*dataobj.TsdbQueryResponse) map[int64]*AggrTsValue {
	dataMap := make(map[int64]*AggrTsValue)
	datasLen := len(datas)
	for i := 0; i < datasLen; i++ {
		for j := 0; j < len(datas[i].Values); j++ {
			value := datas[i].Values[j].Value
			if math.IsNaN(float64(value)) {
				continue
			}
			if _, exists := dataMap[datas[i].Values[j].Timestamp]; exists {
				dataMap[datas[i].Values[j].Timestamp].Value += value
			} else {
				v := AggrTsValue{
					Value: value,
				}
				dataMap[datas[i].Values[j].Timestamp] = &v
			}
		}
	}
	return dataMap
}

func avg(datas []*dataobj.TsdbQueryResponse) map[int64]*AggrTsValue {
	dataMap := make(map[int64]*AggrTsValue)
	datasLen := len(datas)
	for i := 0; i < datasLen; i++ {
		for j := 0; j < len(datas[i].Values); j++ {
			value := datas[i].Values[j].Value
			if math.IsNaN(float64(value)) {
				continue
			}

			if _, exists := dataMap[datas[i].Values[j].Timestamp]; exists {
				dataMap[datas[i].Values[j].Timestamp].Count += 1
				dataMap[datas[i].Values[j].Timestamp].Value += (datas[i].Values[j].Value - dataMap[datas[i].Values[j].Timestamp].Value) / dataobj.JsonFloat(dataMap[datas[i].Values[j].Timestamp].Count)
			} else {
				v := AggrTsValue{
					Value: value,
					Count: 1,
				}
				dataMap[datas[i].Values[j].Timestamp] = &v
			}
		}
	}
	return dataMap
}

func max(datas []*dataobj.TsdbQueryResponse) map[int64]*AggrTsValue {
	dataMap := make(map[int64]*AggrTsValue)
	datasLen := len(datas)
	for i := 0; i < datasLen; i++ {
		for j := 0; j < len(datas[i].Values); j++ {
			value := datas[i].Values[j].Value
			if math.IsNaN(float64(value)) {
				continue
			}

			if _, exists := dataMap[datas[i].Values[j].Timestamp]; exists {
				if value > dataMap[datas[i].Values[j].Timestamp].Value {
					dataMap[datas[i].Values[j].Timestamp].Value = value
				}
			} else {
				v := AggrTsValue{
					Value: value,
				}
				dataMap[datas[i].Values[j].Timestamp] = &v
			}
		}
	}
	return dataMap
}

func min(datas []*dataobj.TsdbQueryResponse) map[int64]*AggrTsValue {
	dataMap := make(map[int64]*AggrTsValue)
	datasLen := len(datas)
	for i := 0; i < datasLen; i++ {
		for j := 0; j < len(datas[i].Values); j++ {
			value := datas[i].Values[j].Value
			if math.IsNaN(float64(value)) {
				continue
			}

			if _, exists := dataMap[datas[i].Values[j].Timestamp]; exists {
				if value < dataMap[datas[i].Values[j].Timestamp].Value {
					dataMap[datas[i].Values[j].Timestamp].Value = value
				}
			} else {
				v := AggrTsValue{
					Value: value,
				}
				dataMap[datas[i].Values[j].Timestamp] = &v
			}
		}
	}
	return dataMap
}
