package routes

import (
	"github.com/gin-gonic/gin"
)

// Config routes
func Config(r *gin.Engine) {
	r.Static("/pub", "./pub")
	r.StaticFile("/favicon.ico", "./pub/favicon.ico")

	self := r.Group("/api/portal")
	{
		self.GET("/ping", ping)
		self.GET("/version", version)
		self.GET("/pid", pid)
	}

}
