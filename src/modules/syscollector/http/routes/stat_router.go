package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/open-falcon/falcon-ng/src/modules/syscollector/ports"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/procs"
)

func getStrategy(c *gin.Context) {
	var resp []interface{}

	port := ports.ListPorts()
	for _, p := range port {
		resp = append(resp, p)
	}
	proc := procs.ListProcs()
	for _, p := range proc {
		resp = append(resp, p)
	}

	renderData(c, resp, nil)
}
