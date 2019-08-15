package routes

import (
	"fmt"
	"os"
	"strconv"

	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/funcs"
)

func ping(c *gin.Context) {
	c.String(200, "pong")
}

func version(c *gin.Context) {
	c.String(200, strconv.Itoa(config.Version))
}

func addr(c *gin.Context) {
	c.String(200, c.Request.RemoteAddr)
}

func pid(c *gin.Context) {
	c.String(200, fmt.Sprintf("%d", os.Getpid()))
}

func pushData(c *gin.Context) {
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

	funcs.Push(metricValues)

	if msg != "" {
		renderMessage(c, msg)
		return
	}

	renderData(c, "ok", nil)
	return
}

func getCollect(c *gin.Context) {

}
