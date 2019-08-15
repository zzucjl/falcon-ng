package routes

import (
	"github.com/open-falcon/falcon-ng/src/modules/index/cache"
	"github.com/open-falcon/falcon-ng/src/modules/index/config"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/logger"
)

type EndpointsRecv struct {
	Endpoints []string `json:"endpoints"`
}

type EndpointMetricList struct {
	Metrics []string `json:"metrics"`
}

func GetMetricsByEndpoint(c *gin.Context) {
	recv := EndpointsRecv{}
	errors.Dangerous(c.ShouldBindJSON(&recv))

	m := make(map[string]struct{})
	resp := EndpointMetricList{}
	for _, endpoint := range recv.Endpoints {
		metrics := cache.EndpointDBObj.GetMetricsBy(endpoint)
		for _, metric := range metrics {
			if _, exists := m[metric]; exists {
				continue
			}
			m[metric] = struct{}{}
			resp.Metrics = append(resp.Metrics, metric)
		}
	}
	for metric, _ := range config.DEFAULT_METRIC {
		resp.Metrics = append(resp.Metrics, metric)
	}

	renderData(c, resp, nil)
}

func DeleteMetrics(c *gin.Context) {
	recv := EndpointMetricRecv{}
	errors.Dangerous(c.ShouldBindJSON(&recv))

	for _, endpoint := range recv.Endpoints {
		metricsItem, exists := cache.EndpointDBObj.GetMetrics(endpoint)
		if !exists {
			continue
		}
		for _, metric := range recv.Metrics {
			metricsItem.CleanMetric(metric)
		}
	}

	renderData(c, "ok", nil)
}

type EndpointMetricRecv struct {
	Endpoints []string `json:"endpoints"`
	Metrics   []string `json:"metrics"`
}

type EndpointTagkvResp struct {
	Endpoints []string             `json:"endpoints"`
	Metric    string               `json:"metric"`
	Tagkv     []*cache.TagkvStruct `json:"tagkv"`
}

func GetTagkvByEndpoint(c *gin.Context) {
	recv := EndpointMetricRecv{}
	errors.Dangerous(c.ShouldBindJSON(&recv))
	resp := []*EndpointTagkvResp{}
	tagkvFilter := make(map[string]map[string]struct{})
	for _, metric := range recv.Metrics {
		tagkvs := []*cache.TagkvStruct{}

		//metric 在默认metric列表
		if _, exists := config.DEFAULT_METRIC[metric]; exists {
			TagkvResp := &EndpointTagkvResp{
				Endpoints: recv.Endpoints,
				Metric:    metric,
				Tagkv:     tagkvs,
			}
			resp = append(resp, TagkvResp)
			continue
		}

		for _, endpoint := range recv.Endpoints {
			metricsItem, exists := cache.EndpointDBObj.GetMetrics(endpoint)
			if !exists {
				logger.Warningf("metrics not found by %s", endpoint)
				continue
			}

			tagks, exists := metricsItem.GetTagksStruct(metric)
			if !exists {
				logger.Warningf("tagkStruct not found by %s %s", endpoint, metric)
				continue
			}
			tagvs := tagks.GetTagkv()
			for _, kv := range tagvs {
				tagvFilter, exists := tagkvFilter[kv.TagK]
				if !exists {
					tagvFilter = make(map[string]struct{})
				}

				for _, tagv := range kv.TagV {
					if _, exists := tagvFilter[tagv]; exists {
						continue
					}
					tagvFilter[tagv] = struct{}{}
				}
				tagkvFilter[kv.TagK] = tagvFilter
			}
		}

		for tagk, tagvFilter := range tagkvFilter {
			tagvs := []string{}
			for v, _ := range tagvFilter {
				tagvs = append(tagvs, v)
			}
			tagkv := &cache.TagkvStruct{
				TagK: tagk,
				TagV: tagvs,
			}
			tagkvs = append(tagkvs, tagkv)
		}

		TagkvResp := EndpointTagkvResp{
			Endpoints: recv.Endpoints,
			Metric:    metric,
			Tagkv:     tagkvs,
		}
		resp = append(resp, &TagkvResp)
	}
	renderData(c, resp, nil)
}

type FullmatchByEndpointRecv struct {
	Endpoints []string             `json:"endpoints"`
	Metric    string               `json:"metric"`
	Tagkv     []*cache.TagkvStruct `json:"tagkv"`
}

type FullmatchByEndpointResp struct {
	Endpoints []string `json:"endpoints"`
	Metric    string   `json:"metric"`
	Tags      []string `json:"tags"`
	Step      int      `json:"step"`
	DsType    string   `json:"dstype"`
}

func FullmatchByEndpoint(c *gin.Context) {
	recv := []FullmatchByEndpointRecv{}
	errors.Dangerous(c.ShouldBindJSON(&recv))

	tagFilter := make(map[string]struct{})
	tagsList := []string{}

	var resp []FullmatchByEndpointResp

	for _, r := range recv {
		metric := r.Metric
		tagkv := r.Tagkv
		step := 0
		dsType := ""
		if _, exists := config.DEFAULT_METRIC[metric]; exists {
			resp = append(resp, FullmatchByEndpointResp{
				Endpoints: r.Endpoints,
				Metric:    r.Metric,
				Tags:      []string{},
				Step:      config.DEFAULT_STEP,
				DsType:    config.DEFAULT_DSTYPE,
			})

			continue
		}

		for _, endpoint := range r.Endpoints {

			if endpoint == "" {
				logger.Warningf("非法请求: endpoint字段缺失:%v", r)
				continue
			}
			if metric == "" {
				logger.Warningf("非法请求: metric字段缺失:%v", r)
				continue
			}
			metricsItem, exists := cache.EndpointDBObj.GetMetrics(endpoint)
			if !exists {
				logger.Warningf("not found metrics by endpoint:%s", endpoint)
				continue
			}
			if step == 0 || dsType == "" {
				step, dsType, exists = metricsItem.GetMetricStepAndDstype(metric)
				if !exists {
					logger.Warningf("not found step by endpoint:%s metric:%v\n", endpoint, metric)
					continue
				}
			}

			tags, err := cache.EndpointDBObj.QueryCountersFullMatchByTags(endpoint, metric, tagkv)
			if err != nil {
				logger.Warning(err)
				continue
			}

			for _, tag := range tags {
				if _, exists := tagFilter[tag]; !exists {
					tagsList = append(tagsList, tag)
					tagFilter[tag] = struct{}{}
				}
			}
		}

		resp = append(resp, FullmatchByEndpointResp{
			Endpoints: r.Endpoints,
			Metric:    r.Metric,
			Tags:      tagsList,
			Step:      step,
			DsType:    dsType,
		})
	}

	renderData(c, resp, nil)
}

type CludeByEndpointRecv struct {
	Endpoints []string         `json:"endpoints"`
	Metric    string           `json:"metric"`
	Include   cache.XCludeList `json:"include"`
	Exclude   cache.XCludeList `json:"exclude"`
}

func CludeByEndpoint(c *gin.Context) {
	recv := []CludeByEndpointRecv{}
	errors.Dangerous(c.ShouldBindJSON(&recv))

	tagFilter := make(map[string]struct{})
	tagList := []string{}
	var resp []FullmatchByEndpointResp

	for _, r := range recv {
		metric := r.Metric
		includeList := r.Include
		excludeList := r.Exclude
		step := 0
		dsType := ""

		if _, exists := config.DEFAULT_METRIC[metric]; exists {
			resp = append(resp, FullmatchByEndpointResp{
				Endpoints: r.Endpoints,
				Metric:    r.Metric,
				Tags:      []string{},
				Step:      config.DEFAULT_STEP,
				DsType:    config.DEFAULT_DSTYPE,
			})

			continue
		}

		for _, endpoint := range r.Endpoints {
			if endpoint == "" {
				logger.Warningf("非法请求: endpoint字段缺失:%v", r)
				continue
			}
			if metric == "" {
				logger.Warningf("非法请求: metric字段缺失:%v", r)
				continue
			}

			metricsItem, exists := cache.EndpointDBObj.GetMetrics(endpoint)
			if !exists {
				logger.Warningf("not found metrics by endpoint:%s", endpoint)
				continue
			}

			if step == 0 || dsType == "" {
				step, dsType, exists = metricsItem.GetMetricStepAndDstype(metric)
				if !exists {
					logger.Warningf("not found step by endpoint:%s metric:%v\n", endpoint, metric)
					continue
				}
			}

			tags, err := cache.EndpointDBObj.QueryCountersByNsMetricXclude(endpoint, metric, includeList, excludeList)
			if err != nil {
				logger.Warning(err)
				continue
			}
			//logger.Debug(endpoint, metric, includeList, excludeList)

			for _, tag := range tags {
				if tag == "" { //过滤掉空字符串
					continue
				}

				if _, exists := tagFilter[tag]; !exists {
					tagList = append(tagList, tag)
					tagFilter[tag] = struct{}{}
				}
			}
		}
		resp = append(resp, FullmatchByEndpointResp{
			Endpoints: r.Endpoints,
			Metric:    r.Metric,
			Tags:      tagList,
			Step:      step,
			DsType:    dsType,
		})

	}

	renderData(c, resp, nil)
}

func DumpIndex(c *gin.Context) {
	err := cache.EndpointDBObj.Persist("normal")
	errors.Dangerous(err)

	renderData(c, "ok", nil)
}
