package routes

import (
	"github.com/gin-gonic/gin"
)

// Config routes
func Config(r *gin.Engine) {
	sys := r.Group("/api/sender/sys")
	{
		sys.GET("/ping", ping)
		sys.GET("/version", version)
		sys.GET("/pid", pid)
		sys.GET("/addr", addr)
	}

	sender := r.Group("/api/sender")
	{
		sender.POST("/im", im)
		sender.POST("/mail", mail)
		sender.POST("/sms", sms)
		sender.POST("/voice", voice)
	}

}
