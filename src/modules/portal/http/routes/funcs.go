package routes

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/pager"

	"github.com/open-falcon/falcon-ng/src/model"
)

func queryInt(c *gin.Context, key string, defaultVal int) int {
	strv := c.Query(key)
	if strv == "" {
		return defaultVal
	}

	intv, err := strconv.Atoi(strv)
	if err != nil {
		errors.Bomb("cannot convert [%s] to int", strv)
	}

	return intv
}

func queryInt64(c *gin.Context, key string, defaultVal int64) int64 {
	strv := c.Query(key)
	if strv == "" {
		return defaultVal
	}

	intv, err := strconv.ParseInt(strv, 10, 64)
	if err != nil {
		errors.Bomb("cannot convert [%s] to int64", strv)
	}

	return intv
}

func offset(c *gin.Context, limit int, total interface{}) int {
	return pager.NewPaginator(c.Request, limit, total).Offset()
}

func renderMessage(c *gin.Context, msg string) {
	c.JSON(200, gin.H{"err": msg})
}

func renderError(c *gin.Context, err error) {
	if err != nil {
		renderMessage(c, err.Error())
		return
	}

	renderMessage(c, "")
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

func loginRoot(c *gin.Context) *model.User {
	user := loginUser(c)
	if user.IsRoot == 0 {
		errors.Bomb("no privilege")
	}

	return user
}
