package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/index/config"

	"github.com/toolkits/pkg/logger"
)

type EndpointMetricsStruct struct { // ns -> metrics
	sync.RWMutex
	Metrics map[string]*MetricsStruct `json:"ns_metric"`
}

//push 索引数据
func (e *EndpointMetricsStruct) Push(item dataobj.IndexModel) error {
	now := time.Now().Unix()
	counter := dataobj.SortedTags(item.Tags)
	metric := item.Metric

	metricsItem := e.MustGetMetrics(item.Endpoint)
	metricItem := metricsItem.MustGetMetricStruct(metric, now)

	metricItem.Updated = now
	metricItem.Step = item.Step
	metricItem.DsType = item.DsType
	metricItem.Counters.Update(counter, now, int64(item.Step), item.DsType)

	tagks := metricItem.Tagks
	for k, v := range item.Tags {
		tagk := tagks.MustGetTagkStruct(k, now)
		tagk.Tagvs.Set(v, now)
	}
	return nil
}

func (e *EndpointMetricsStruct) Clean(timeDuration int64) {
	endpoints := e.GetEndpoints()
	now := time.Now().Unix()
	for _, endpoint := range endpoints {
		metricsItem, exists := e.GetMetrics(endpoint)
		if !exists {
			continue
		}
		if metricsItem.Len() < 1 {
			e.Lock()
			delete(e.Metrics, endpoint)
			e.Unlock()
		}

		metricsItem.Clean(now, timeDuration)
	}
}

func (e *EndpointMetricsStruct) GetMetrics(endpoint string) (*MetricsStruct, bool) {
	e.RLock()
	defer e.RUnlock()
	metricsStruct, exists := e.Metrics[endpoint]
	return metricsStruct, exists
}

func (e *EndpointMetricsStruct) MustGetMetrics(endpoint string) *MetricsStruct {
	e.RLock()
	var metricsStruct *MetricsStruct
	if _, exists := e.Metrics[endpoint]; !exists {
		e.RUnlock()

		e.Lock()
		e.Metrics[endpoint] = &MetricsStruct{MetricMap: make(map[string]*MetricStruct)} //MetricStruct{}.New(metric, now)
		metricsStruct = e.Metrics[endpoint]
		e.Unlock()
	} else {
		e.RUnlock()
		metricsStruct = e.Metrics[endpoint]
	}
	return metricsStruct
}

func (e *EndpointMetricsStruct) GetMetricsBy(endpoint string) []string {
	e.RLock()
	defer e.RUnlock()
	if _, exists := e.Metrics[endpoint]; !exists {
		return []string{}
	}
	return e.Metrics[endpoint].GetMetrics()
}

func (e *EndpointMetricsStruct) QueryCountersFullMatchByTags(endpoint, metric string, tags XCludeList) ([]string, error) {
	//check if over limit range
	err := tags.CheckFullMatch(int64(config.Config.Limit.UI)) //todo 改为int
	if err != nil {
		return []string{}, fmt.Errorf("err:%v  endpoint:%v metric:%v\n", err.Error(), endpoint, metric)
	}

	allCombination, err := tags.GetAllCombinationString()
	if err != nil {
		return []string{}, err
	}
	if len(allCombination) > config.Config.Limit.FullmatchLogCounter {
		// 超限 则代表tags数组非常大, 不打印详细信息
		logger.Warningf("fullmatch get too much counters, endpoint:%s metric:%s\n", endpoint, metric)
	}

	return allCombination, nil
}

func (e *EndpointMetricsStruct) QueryCountersByNsMetricXclude(endpoint, metric string, include, exclude XCludeList) ([]string, error) {
	if len(include) == 0 && len(exclude) == 0 {
		metricsItem, exists := e.GetMetrics(endpoint)
		if !exists {
			logger.Warningf("not found metric by endpoint:%s metric:%v\n", endpoint, metric)
			return []string{}, nil
		}

		countersItem, exists := metricsItem.GetMetricStructCounters(metric)
		if !exists {
			logger.Warningf("not found step by endpoint:%s metric:%v\n", endpoint, metric)
			return []string{}, nil
		}

		counterList := countersItem.GetCounters()
		return counterList, nil
	}

	tagkvs, err := e.QueryTagkvMapByNsMetric(endpoint, metric)
	if err != nil {
		return []string{}, err
	}
	if len(include) > 0 {
		// include合法性校验
		for _, clude := range include {
			_, exists := tagkvs[clude.TagK]
			if !exists {
				return []string{}, fmt.Errorf("include tagk %s 不存在", clude)
			}
		}
	}

	inMap := make(map[string]map[string]bool)
	exMap := make(map[string]map[string]bool)

	if len(include) > 0 {
		for _, clude := range include {
			if _, found := inMap[clude.TagK]; !found {
				inMap[clude.TagK] = make(map[string]bool)
			}
			for _, tagv := range clude.TagV {
				inMap[clude.TagK][tagv] = true
			}
		}
	}

	if len(exclude) > 0 {
		for _, clude := range exclude {
			if _, found := exMap[clude.TagK]; !found {
				exMap[clude.TagK] = make(map[string]bool)
			}
			for _, tagv := range clude.TagV {
				exMap[clude.TagK][tagv] = true
			}
		}
	}

	fullmatch := make(map[string][]string)
	for tagk, tagvs := range tagkvs {
		for _, tagv := range tagvs {
			// 排除必须排除的, exclude的优先级高于include
			if _, found1 := exMap[tagk]; found1 {
				if _, found2 := exMap[tagk][tagv]; found2 {
					continue
				}
			}
			// 包含必须包含的
			if _, found3 := inMap[tagk]; found3 {
				if _, found4 := inMap[tagk][tagv]; found4 {
					if _, found := fullmatch[tagk]; !found {
						fullmatch[tagk] = make([]string, 0)
					}
					fullmatch[tagk] = append(fullmatch[tagk], tagv)
				}
				continue
			}
			// 除此之外全都包含
			if _, found := fullmatch[tagk]; !found {
				fullmatch[tagk] = make([]string, 0)
			}
			fullmatch[tagk] = append(fullmatch[tagk], tagv)
		}
	}

	// 部分tagk的tagv全部被exclude 或者 完全没有匹配的
	if len(fullmatch) != len(tagkvs) || len(fullmatch) == 0 {
		return []string{}, nil
	}

	retrieve := false
	multiRes := 1
	for _, tagvs := range fullmatch {
		multiRes = multiRes * len(tagvs)
		if multiRes > config.Config.Limit.Clude {
			logger.Warningf("xclude fullmatch get too much counters, retrieve, endpoint:%s metric:%s, "+
				"include:%v, exclude:%v\n", endpoint, metric, include, exclude)
			retrieve = true
		}
	}
	if retrieve {
		logger.Info("retrieve:", retrieve)
	}

	var tags XCludeList
	for tagk, tagvs := range fullmatch {
		tags = append(tags, &TagkvStruct{
			TagK: tagk,
			TagV: tagvs,
		})
	}

	retList, err := e.QueryCountersFullMatchByTags(endpoint, metric, tags)

	return retList, err
}

func (e *EndpointMetricsStruct) QueryTagkvMapByNsMetric(endpoint, metric string) (map[string][]string, error) {
	tagkvs := make(map[string][]string)
	metricsItem, exists := e.GetMetrics(endpoint)
	if !exists {
		return tagkvs, nil
	}

	tagk, exists := metricsItem.GetTagksStruct(metric)
	if !exists {
		return tagkvs, nil
	}

	tagkvs = tagk.GetTagkvMap()
	return tagkvs, nil
}

func (e *EndpointMetricsStruct) GetEndpoints() []string {
	e.RLock()
	defer e.RUnlock()

	length := len(e.Metrics)
	ret := make([]string, length)
	i := 0
	for endpoint, _ := range e.Metrics {
		ret[i] = endpoint
		i++
	}
	return ret
}

func (e *EndpointMetricsStruct) Persist(mode string) error {
	if mode == "normal" {
		if !semaPermanence.TryAcquire() {
			return fmt.Errorf("Permanence operate is Already running...")
		}
	} else if mode == "end" {
		semaPermanence.Acquire()
	} else {
		return fmt.Errorf("Your mode is Wrong![normal,end]")
	}

	defer semaPermanence.Release()

	tmpDir := fmt.Sprintf("%s%s", PERMANENCE_DIR, "tmp")
	finalDir := fmt.Sprintf("%s%s", PERMANENCE_DIR, "db")

	var err error
	//清空tmp目录
	if err = os.RemoveAll(tmpDir); err != nil {
		return err
	}

	//创建tmp目录
	if err = os.MkdirAll(tmpDir, 0777); err != nil {
		return err
	}

	//填充tmp目录
	endpoints := e.GetEndpoints()
	logger.Infof("now start to save index data to disk...[ns-num:%d][mode:%s]\n", len(endpoints), mode)

	for i, endpoint := range endpoints {

		logger.Infof("sync [%s] to disk, [%d%%] complete\n", endpoint, int((float64(i)/float64(len(endpoints)))*100))
		metricsStruct, exists := e.GetMetrics(endpoint)
		if !exists || metricsStruct == nil {
			continue
		}

		metricsStruct.Lock()
		body, err_m := json.Marshal(metricsStruct)
		metricsStruct.Unlock()

		if err_m != nil {
			logger.Errorf("marshal struct to json failed : [endpoint:%s][msg:%s]\n", endpoint, err_m.Error())
			continue
		}

		err = ioutil.WriteFile(fmt.Sprintf("%s/%s", tmpDir, endpoint), body, 0666)
		if err != nil {
			logger.Errorf("write file error : [endpoint:%s][msg:%s]\n", endpoint, err.Error())
		}
	}
	logger.Infof("sync to disk , [%d%%] complete\n", 100)

	//清空db目录
	if err = os.RemoveAll(finalDir); err != nil {
		return err
	}

	//将tmp目录改名为final
	if err = os.Rename(tmpDir, finalDir); err != nil {
		return err
	}

	return nil
}
