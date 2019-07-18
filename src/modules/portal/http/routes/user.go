package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"

	"github.com/open-falcon/falcon-ng/src/model"
)

func userListGet(c *gin.Context) {
	limit := queryInt(c, "limit", 20)
	query := c.Query("query")

	total, err := model.UserTotal(query)
	errors.Dangerous(err)

	list, err := model.UserGets(query, limit, offset(c, limit, total))
	errors.Dangerous(err)

	c.JSON(200, gin.H{
		"list":  list,
		"total": total,
	})
}
