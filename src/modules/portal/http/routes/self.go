package routes

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-ng/src/modules/portal/config"
)

func ping(c *gin.Context) {
	c.String(200, "pong")
}

func version(c *gin.Context) {
	c.String(200, "%d", config.Version)
}

func pid(c *gin.Context) {
	c.String(200, "%d", os.Getpid())
}

func addr(c *gin.Context) {
	c.String(200, c.Request.RemoteAddr)
}
