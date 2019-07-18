package routes

import (
	"github.com/gin-gonic/gin"
)

// Config routes
func Config(r *gin.Engine) {
	self := r.Group("/self")
	{
		self.GET("/ping", ping)
		self.GET("/version", version)
		self.GET("/pid", pid)
	}

}
