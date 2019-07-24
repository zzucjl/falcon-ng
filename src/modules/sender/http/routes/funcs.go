package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"

	"github.com/open-falcon/falcon-ng/src/modules/sender/config"
)

func renderMessage(c *gin.Context, v interface{}) {
	if v == nil {
		c.JSON(200, gin.H{"err": ""})
		return
	}

	switch t := v.(type) {
	case string:
		c.JSON(200, gin.H{"err": t})
	case error:
		c.JSON(200, gin.H{"err": t.Error()})
	}
}

func auth(c *gin.Context) {
	val := c.GetHeader("x-srv-token")
	if _, exists := config.GetCfgYml().AuthMap[val]; !exists {
		errors.Dangerous("auth failed")
	}
}
