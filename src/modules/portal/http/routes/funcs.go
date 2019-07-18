package routes

import "github.com/gin-gonic/gin"

func renderMessage(c *gin.Context, msg string) {
	c.JSON(200, gin.H{"err": msg})
}
