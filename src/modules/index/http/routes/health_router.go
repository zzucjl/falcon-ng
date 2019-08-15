package routes

import (
	"fmt"
	"os"
	"strconv"

	"github.com/open-falcon/falcon-ng/src/modules/index/config"

	"github.com/gin-gonic/gin"
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
