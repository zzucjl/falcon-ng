package routes

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/net/httplib"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/transfer/backend"
	. "github.com/open-falcon/falcon-ng/src/modules/transfer/config"

	"github.com/gin-gonic/gin"
)

type QueryDataReq struct {
	Start  int64       `json:"start"`
	End    int64       `json:"end"`
	Series []SeriesReq `json:"series"`
}

type Tagkv struct {
	TagK string   `json:"tagk"`
	TagV []string `json:"tagv"`
}

type SeriesReq struct {
	Endpoints []string `json:"endpoints"`
	Metric    string   `json:"metric"`
	Tagkv     []*Tagkv `json:"tagkv"`
}

type SeriesResp struct {
	Dat []Series `json:"dat"`
	Err string   `json:"err"`
}

type Series struct {
	Endpoints []string `json:"endpoints"`
	Metric    string   `json:"metric"`
	Tags      []string `json:"tags"`
	Step      int      `json:"step"`
	DsType    string   `json:"dstype"`
}

func QueryDataForJudge(c *gin.Context) {
	var inputs []dataobj.QueryData

	errors.Dangerous(c.ShouldBindJSON(&inputs))
	resp := backend.FetchData(inputs)
	renderData(c, resp, nil)
}

func QueryData(c *gin.Context) {
	var input QueryDataReq

	errors.Dangerous(c.ShouldBindJSON(&input))

	queryData, err := GetSeries(input.Start, input.End, input.Series)
	if err != nil {
		logger.Error(err, input)
		renderMessage(c, "query err")
		return
	}

	resp := backend.FetchData(queryData)
	renderData(c, resp, nil)
}

func QueryDataForUI(c *gin.Context) {
	var input dataobj.QueryDataForUI

	errors.Dangerous(c.ShouldBindJSON(&input))
	logger.Info("debug: ", input)

	resp := backend.FetchDataForUI(input)
	renderData(c, resp, nil)
}

func GetSeries(start, end int64, req []SeriesReq) ([]dataobj.QueryData, error) {
	var res SeriesResp
	var queryDatas []dataobj.QueryData

	if len(req) < 1 {
		return queryDatas, fmt.Errorf("req err")
	}

	if len(Config.Index.Addrs) < 1 {
		return queryDatas, fmt.Errorf("index addr is nil")
	}

	i := rand.Intn(len(Config.Index.Addrs))
	addr := Config.Index.Addrs[i]

	resp, err := httplib.PostJSON(addr, Config.Index.Timeout, req, nil)
	if err != nil {
		return queryDatas, err
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		logger.Error(string(resp))
		return queryDatas, err
	}

	for _, item := range res.Dat {
		counters := []string{}
		if len(item.Tags) == 0 {
			counters = append(counters, item.Metric)
		} else {
			for _, tag := range item.Tags {
				tagMap, err := dataobj.SplitTagsString(tag)
				if err != nil {
					logger.Warning(err, tag)
					continue
				}
				tagStr := dataobj.SortedTags(tagMap)
				counter := dataobj.PKWithTags(item.Metric, tagStr)
				counters = append(counters, counter)
			}
		}

		queryData := dataobj.QueryData{
			Start:      start,
			End:        end,
			Endpoints:  item.Endpoints,
			Counters:   counters,
			ConsolFunc: "AVERAGE",
			DsType:     item.DsType,
			Step:       item.Step,
		}
		queryDatas = append(queryDatas, queryData)
	}

	return queryDatas, err
}
