package routes

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/pager"

	"github.com/open-falcon/falcon-ng/src/model"
)

func urlParamStr(c *gin.Context, field string) string {
	val := c.Param(field)

	if val == "" {
		errors.Bomb("[%s] is blank", field)
	}

	return val
}

func urlParamInt64(c *gin.Context, field string) int64 {
	strval := urlParamStr(c, field)
	intval, err := strconv.ParseInt(strval, 10, 64)
	if err != nil {
		errors.Bomb("cannot convert %s to int64", strval)
	}

	return intval
}

func urlParamInt(c *gin.Context, field string) int {
	return int(urlParamInt64(c, field))
}

func queryStr(c *gin.Context, key string, defaultVal string) string {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}

	return val
}

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

func renderMessage(c *gin.Context, v interface{}) {
	if v == nil {
		c.JSON(200, gin.H{"err": ""})
		return
	}

	switch t := v.(type) {
	case string:
		c.JSON(200, gin.H{"err": t})
	case error:
		c.JSON(200, gin.H{"err": t.Error()})
	}
}

func renderData(c *gin.Context, data interface{}, err error) {
	if err == nil {
		c.JSON(200, gin.H{"dat": data, "err": ""})
		return
	}

	renderMessage(c, err.Error())
}

func loginUsername(c *gin.Context) string {
	username, _ := c.Get("username")
	return username.(string)
}

func loginUser(c *gin.Context) *model.User {
	username := loginUsername(c)

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

func mustUser(id int64) *model.User {
	user, err := model.UserGet("id", id)
	if err != nil {
		errors.Bomb("cannot retrieve user[%d]: %v", id, err)
	}

	if user == nil {
		errors.Bomb("no such user[%d]", id)
	}

	return user
}
