package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/str"

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

type endpointForm struct {
	Alias string `json:"alias"`
}

func endpointPut(c *gin.Context) {
	var f endpointForm
	errors.Dangerous(c.ShouldBind(&f))

	id := urlParamInt64(c, "id")
	endpoint, err := model.EndpointGet("id", id)
	errors.Dangerous(err)

	if endpoint == nil {
		errors.Bomb("no such endpoint, id: %d", id)
	}

	endpoint.Alias = f.Alias
	renderMessage(c, endpoint.Update("alias"))
}

func endpointDel(c *gin.Context) {
	ids := str.IdsInt64(queryStr(c, "ids", ""))
	renderMessage(c, model.EndpointDel(ids))
}
