package cache

import (
	"sync"
)

type MetricsStruct struct { // metrics
	sync.RWMutex
	MetricMap map[string]*MetricStruct
}

func (m *MetricsStruct) Clean(now, timeDuration int64) {
	m.Lock()
	defer m.Unlock()
	for metric, metricStruct := range m.MetricMap {
		if now-metricStruct.Updated > timeDuration {
			//清理metric
			delete(m.MetricMap, metric)
		} else {
			//清理tagk
			metricStruct.Tagks.Clean(now, timeDuration)

			//清理counter
			metricStruct.Counters.Clean(now, timeDuration)

		}
	}
}

func (m *MetricsStruct) CleanEndpoint(endpoint string) {
	m.Lock()
	defer m.Unlock()
	for _, metricStruct := range m.MetricMap {
		metricStruct.Tagks.CleanEndpoint(endpoint)

		//清理counter
		metricStruct.Counters.CleanEndpoint(endpoint)

	}
	return
}

func (m *MetricsStruct) CleanMetric(metric string) {
	m.Lock()
	defer m.Unlock()
	delete(m.MetricMap, metric)
	return
}

func (m *MetricsStruct) Len() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.MetricMap)
}

func (m *MetricsStruct) MustGetMetricStruct(metric string, now int64) *MetricStruct {
	var nsmetric *MetricStruct
	m.RLock()
	if _, ok := m.MetricMap[metric]; !ok {
		m.RUnlock()
		m.Lock()
		if _, ok := m.MetricMap[metric]; !ok {
			m.MetricMap[metric] = MetricStruct{}.New(metric, now)
		}
		nsmetric = m.MetricMap[metric]
		m.Unlock()

	} else {
		nsmetric = m.MetricMap[metric]
		m.RUnlock()
	}
	return nsmetric
}

func (m *MetricsStruct) GetMetricStepAndDstype(metric string) (int, string, bool) {
	m.RLock()
	defer m.RUnlock()
	metricStruct, exists := m.MetricMap[metric]
	if !exists {
		return 0, "", exists
	}
	return metricStruct.Step, metricStruct.DsType, exists
}

func (m *MetricsStruct) GetMetricStructCounters(metric string) (*CountersStruct, bool) {
	m.RLock()
	defer m.RUnlock()
	metricStruct, exists := m.MetricMap[metric]
	if !exists {
		return nil, exists
	}
	return metricStruct.Counters, exists
}

func (m *MetricsStruct) GetTagksStruct(metric string) (*TagksStruct, bool) {
	m.RLock()
	defer m.RUnlock()

	nsmetric, exists := m.MetricMap[metric]
	if !exists {
		return nil, exists
	}

	return nsmetric.Tagks, exists
}

func (m *MetricsStruct) GetMetrics() []string {
	m.RLock()
	defer m.RUnlock()
	var metrics []string
	for k, _ := range m.MetricMap {
		metrics = append(metrics, k)
	}
	return metrics
}
