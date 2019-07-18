package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/toolkits/pkg/errors"
)

func renderMessage(c *gin.Context, msg string) {
	c.JSON(200, gin.H{"err": msg})
}

func renderData(c *gin.Context, data interface{}, err error) {
	if err == nil {
		c.JSON(200, gin.H{"dat": data, "err": ""})
		return
	}

	renderMessage(c, err.Error())
}

func loginUser(c *gin.Context) *model.User {
	username, has := c.Get("username")
	if !has {
		return nil
	}

	if username == "" {
		return nil
	}

	user, err := model.UserGet("username", username)
	errors.Dangerous(err)

	if user == nil {
		errors.Bomb("login first please")
	}

	return user
}
