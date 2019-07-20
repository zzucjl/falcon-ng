package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/open-falcon/falcon-ng/src/model"
)

func treeGet(c *gin.Context) {
	nodes, err := model.NodeGets("")
	renderData(c, nodes, err)
}

func treeSearchGet(c *gin.Context) {
	query := queryStr(c, "query", "")
	nodes, err := model.TreeSearchByPath(query)
	renderData(c, nodes, err)
}
