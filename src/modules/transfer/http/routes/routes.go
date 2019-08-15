package routes

import (
	"github.com/gin-gonic/gin"
)

// Config routes
func Config(r *gin.Engine) {
	sys := r.Group("/api/transfer")
	{
		sys.GET("/ping", ping)
		sys.GET("/version", version)
		sys.GET("/pid", pid)
		sys.GET("/addr", addr)

		sys.POST("/push", PushData)
		sys.POST("/data", QueryDataForJudge)
		sys.POST("/data/ui", QueryDataForUI)
	}

	v2 := r.Group("/api/transfer/v2")
	{
		v2.POST("/data", QueryData)
	}
}
