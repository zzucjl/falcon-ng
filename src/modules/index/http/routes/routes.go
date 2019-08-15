package routes

import (
	"github.com/gin-gonic/gin"
)

// Config routes
func Config(r *gin.Engine) {
	sys := r.Group("/api/index")
	{
		sys.GET("/ping", ping)
		sys.GET("/version", version)
		sys.GET("/pid", pid)
		sys.GET("/addr", addr)

		sys.POST("/metrics", GetMetricsByEndpoint)
		sys.DELETE("/metrics", DeleteMetrics)
		sys.POST("/tagkv", GetTagkvByEndpoint)
		sys.POST("/counter/fullmatch", FullmatchByEndpoint)
		sys.POST("/counter/clude", CludeByEndpoint)
		sys.POST("/dump", DumpIndex)

	}
}
