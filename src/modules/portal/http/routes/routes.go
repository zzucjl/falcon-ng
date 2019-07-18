package routes

import (
	"github.com/gin-gonic/gin"
)

// Config routes
func Config(r *gin.Engine) {
	self := r.Group("/api/portal")
	{
		self.GET("/ping", ping)
		self.GET("/version", version)
		self.GET("/pid", pid)
	}

}
