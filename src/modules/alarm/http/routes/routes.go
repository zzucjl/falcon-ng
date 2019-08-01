package routes

import (
	"github.com/gin-gonic/gin"
)

// Config routes
func Config(r *gin.Engine) {
	sys := r.Group("/api/alarm/sys")
	{
		sys.GET("/ping", ping)
		sys.GET("/version", version)
		sys.GET("/pid", pid)
		sys.GET("/addr", addr)
	}

}
