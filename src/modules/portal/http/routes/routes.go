package routes

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginAuth struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

// Config routes
func Config(r *gin.Engine) {
	r.GET("/self/ping", pong)

	r.GET("/incr", func(c *gin.Context) {
		session := sessions.Default(c)
		var count int
		v := session.Get("count")
		if v == nil {
			count = 0
		} else {
			count = v.(int)
			count++
		}
		session.Set("count", count)
		session.Save()
		c.JSON(200, gin.H{"count": count})
	})

	r.POST("/auth/login", func(c *gin.Context) {
		var la LoginAuth
		c.Bind(&la)
		c.JSON(200, la)
	})
}
