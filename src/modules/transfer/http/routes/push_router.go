package routes

import (
	"fmt"

	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/transfer/backend"
	. "github.com/open-falcon/falcon-ng/src/modules/transfer/config"
)

func PushData(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		renderMessage(c, "blank body")
		return
	}

	recvMetricValues := []*dataobj.MetricValue{}
	metricValues := []*dataobj.MetricValue{}
	errors.Dangerous(c.ShouldBind(&recvMetricValues))

	var msg string
	for _, v := range recvMetricValues {
		logger.Debug("->recv: ", v)
		err := v.CheckValidity()
		if err != nil {
			msg += fmt.Sprintf("recv metric %v err:%v\n", v, err)
			logger.Warningf(msg)
			continue
		}
		metricValues = append(metricValues, v)
	}

	if Config.Tsdb.Enabled {
		backend.Push2TsdbSendQueue(metricValues)
	}

	if msg != "" {
		renderMessage(c, "blank body")
	}

	renderData(c, "ok", nil)
	return
}
