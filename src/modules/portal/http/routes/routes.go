package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-ng/src/modules/portal/http/middleware"
)

// Config routes
func Config(r *gin.Engine) {
	r.Static("/pub", "./pub")
	r.StaticFile("/favicon.ico", "./pub/favicon.ico")

	sys := r.Group("/api/portal/sys")
	{
		sys.GET("/ping", ping)
		sys.GET("/version", version)
		sys.GET("/pid", pid)
		sys.GET("/addr", addr)
	}

	auth := r.Group("/api/portal/auth")
	{
		auth.POST("/login", login)
		auth.GET("/logout", logout)
	}

	self := r.Group("/api/portal/self").Use(middleware.GetCookie())
	{
		self.GET("/profile", selfProfileGet)
		self.PUT("/profile", selfProfilePut)
		self.PUT("/password", selfPasswordPut)
	}

}
