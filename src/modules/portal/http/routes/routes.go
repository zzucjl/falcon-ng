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

	users := r.Group("/api/portal/users")
	{
		users.GET("/invite", userInviteGet)
		users.POST("/invite", userInvitePost)
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
	}

	endpoints := r.Group("/api/portal/endpoints").Use(middleware.GetCookieUser())
	{
		endpoints.GET("/bindings", endpointBindingsGet)
		endpoints.GET("/bynodeids", endpointByNodeIdsGets)
	}

	tree := r.Group("/api/portal/tree").Use(middleware.GetCookieUser())
	{
		tree.GET("", treeGet)
		tree.GET("/search", treeSearchGet)
	}

	node := r.Group("/api/portal/node").Use(middleware.GetCookieUser())
	{
		node.POST("", nodePost)
		node.PUT("/:id/name", nodeNamePut)
		node.DELETE("/:id", nodeDel)
		node.GET("/:id/endpoint", endpointsUnder)
		node.POST("/:id/endpoint-bind", endpointBind)
		node.POST("/:id/endpoint-unbind", endpointUnbind)
		node.GET("/:id/maskconf", maskconfGets)
		node.GET("/:id/screen", screenGets)
		node.POST("/:id/screen", screenPost)
	}

	nodes := r.Group("/api/portal/nodes").Use(middleware.GetCookieUser())
	{
		nodes.GET("/search", nodeSearchGet)
		nodes.GET("/leafids", nodeLeafIdsGet)
		nodes.GET("/pids", nodePidsGet)
		nodes.GET("/byids", nodesByIdsGets)
	}

	maskconf := r.Group("/api/portal/maskconf").Use(middleware.GetCookieUser())
	{
		maskconf.POST("", maskconfPost)
		maskconf.PUT("/:id", maskconfPut)
		maskconf.DELETE("/:id", maskconfDel)
	}

	screen := r.Group("/api/portal/screen").Use(middleware.GetCookieUser())
	{
		screen.PUT("/:id", screenPut)
		screen.DELETE("/:id", screenDel)
		screen.GET("/:id/subclass", screenSubclassGets)
		screen.POST("/:id/subclass", screenSubclassPost)
	}

	subclass := r.Group("/api/portal/subclass").Use(middleware.GetCookieUser())
	{
		subclass.PUT("", screenSubclassPut)
		subclass.DELETE("/:id", screenSubclassDel)
		subclass.GET("/:id/chart", chartGets)
		subclass.POST("/:id/chart", chartPost)
	}

	subclasses := r.Group("/api/portal/subclasses").Use(middleware.GetCookieUser())
	{
		subclasses.PUT("/loc", screenSubclassLocPut)
	}

	chart := r.Group("/api/portal/chart").Use(middleware.GetCookieUser())
	{
		chart.PUT("/:id", chartPut)
		chart.DELETE("/:id", chartDel)
	}

	charts := r.Group("/api/portal/charts").Use(middleware.GetCookieUser())
	{
		charts.PUT("/weights", chartWeightsPut)
	}

	tmpchart := r.Group("/api/portal/tmpchart").Use(middleware.GetCookieUser())
	{
		tmpchart.GET("", tmpChartGet)
		tmpchart.POST("", tmpChartPost)
	}

	event := r.Group("/api/portal/event").Use(middleware.GetCookieUser())
	{
		event.GET("/cur", eventCurGets)
		event.GET("/cur/:id", eventCurGetById)
		event.DELETE("/cur/:id", eventCurDel)
		event.GET("/his", eventHisGets)
		event.GET("/his/:id", eventHisGetById)
		event.POST("/curs/claim", eventCurClaim)
	}

	collect := r.Group("/api/portal/collect").Use(middleware.GetCookieUser())
	{
		collect.POST("", collectPost)
		collect.GET("/list", collectsGet)
		collect.GET("", collectGet)
		collect.PUT("", collectPut)
		collect.DELETE("", collectsDel)
		collect.POST("/check", regExpCheck)
	}

	collects := r.Group("/api/portal/collects")
	{
		collects.GET("/:endpoint", collectGetByEndpoint)
	}

	stra := r.Group("/api/portal/stra").Use(middleware.GetCookieUser())
	{
		stra.POST("", straPost)
		stra.PUT("", straPut)
		stra.DELETE("", strasDel)
		stra.GET("", strasGet)
		stra.GET("/:sid", straGet)
	}

	stras := r.Group("/api/portal/stras")
	{
		stras.GET("/effective", effectiveStrasGet)
	}

	transferProxy := r.Group("/api/transfer")
	{
		transferProxy.GET("/data", transferReq)
	}
}
