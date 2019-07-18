package routes

import (
	"github.com/gin-gonic/gin"
)

func profileGet(c *gin.Context) {
	renderData(c, loginUser(c), nil)
}

func profilePut(c *gin.Context) {
}

func passwordPut(c *gin.Context) {
}
