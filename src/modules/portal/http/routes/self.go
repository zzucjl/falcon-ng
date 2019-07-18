package routes

import "github.com/gin-gonic/gin"

func pong(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong3",
	})
}
