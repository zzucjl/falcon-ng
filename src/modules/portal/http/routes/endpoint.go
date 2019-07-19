package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"

	"github.com/open-falcon/falcon-ng/src/model"
)

func endpointGets(c *gin.Context) {
	limit := queryInt(c, "limit", 20)
	query := queryStr(c, "query", "")
	batch := queryStr(c, "batch", "")
	field := queryStr(c, "field", "ident")

	if !(field == "ident" || field == "alias") {
		errors.Bomb("field invalid")
	}

	total, err := model.EndpointTotal(query, batch, field)
	errors.Dangerous(err)

	list, err := model.EndpointGets(query, batch, field, limit, offset(c, limit, total))
	errors.Dangerous(err)

	renderData(c, gin.H{
		"list":  list,
		"total": total,
	}, nil)
}

type endpointImportForm struct {
	Endpoints []string `json:"endpoints"`
}

func endpointImport(c *gin.Context) {
	var f endpointImportForm
	errors.Dangerous(c.ShouldBind(&f))
	renderMessage(c, model.EndpointImport(f.Endpoints))
}
