package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// GetCookie 从cookie中获取username
func GetCookie() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		value := session.Get("username")
		c.Set("username", value)
		c.Next()
	}
}
