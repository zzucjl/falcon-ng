package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
)

// GetCookieUser 从cookie中获取username
func GetCookieUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		value := session.Get("username")
		if value == nil {
			errors.Bomb("login first please")
		}

		username := value.(string)
		if username == "" {
			errors.Bomb("login first please")
		}

		c.Set("username", value)
		c.Next()
	}
}
