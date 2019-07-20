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

	self := r.Group("/api/portal/self").Use(middleware.GetCookieUser())
	{
		self.GET("/profile", selfProfileGet)
		self.PUT("/profile", selfProfilePut)
		self.PUT("/password", selfPasswordPut)
	}

	user := r.Group("/api/portal/user").Use(middleware.GetCookieUser())
	{
		user.GET("", userListGet)
		user.POST("", userAddPost)
		user.GET("/:id/profile", userProfileGet)
		user.PUT("/:id/profile", userProfilePut)
		user.PUT("/:id/password", userPasswordPut)
		user.DELETE("/:id", userDel)
	}

	team := r.Group("/api/portal/team").Use(middleware.GetCookieUser())
	{
		team.GET("", teamListGet)
		team.POST("", teamAddPost)
		team.PUT("/:id", teamPut)
		team.DELETE("/:id", teamDel)
	}

	endpoint := r.Group("/api/portal/endpoint").Use(middleware.GetCookieUser())
	{
		endpoint.GET("", endpointGets)
		endpoint.POST("", endpointImport)
		endpoint.PUT("/:id", endpointPut)
		endpoint.DELETE("", endpointDel)
		endpoint.GET("/bindings", endpointBindingsGet)
		endpoint.GET("/bynodeids", endpointByNodeIdsGets)
	}

	tree := r.Group("/api/portal/tree").Use(middleware.GetCookieUser())
	{
		tree.GET("", treeGet)
		tree.GET("/search", treeSearchGet)
	}

	node := r.Group("/api/portal/node").Use(middleware.GetCookieUser())
	{
		node.POST("", nodePost)
		node.GET("/search", nodeSearchGet)
		node.PUT("/:id/name", nodeNamePut)
		node.DELETE("/:id", nodeDel)
		node.GET("/leafids", nodeLeafIdsGet)
		node.GET("/pids", nodePidsGet)
		node.GET("/byids", nodesByIdsGets)
		node.GET("/:id/endpoint", endpointsUnder)
		node.POST("/:id/endpoint-bind", endpointBind)
		node.POST("/:id/endpoint-unbind", endpointUnbind)
		node.GET("/:id/maskconf", maskconfGets)
	}

	maskconf := r.Group("/api/portal/maskconf").Use(middleware.GetCookieUser())
	{
		maskconf.POST("", maskconfPost)
		maskconf.PUT("/:id", maskconfPut)
		maskconf.DELETE("/:id", maskconfDel)
	}

}
