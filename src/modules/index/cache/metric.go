package cache

//Metric
type MetricStruct struct {
	Metric  string `json:"metric"`
	Updated int64  `json:"updated"`
	Step    int    `json:"step"`
	DsType  string `json:"dstype"`

	Tagks    *TagksStruct    `json:"ns_metric_tagks"`
	Counters *CountersStruct `json:"ns_metric_counters"`
}

func (this MetricStruct) New(metric string, now int64) *MetricStruct {
	return &MetricStruct{
		Metric:   metric,
		Updated:  now,
		Tagks:    TagksStruct{}.New(),
		Counters: NewCountersStruct(),
	}
}
