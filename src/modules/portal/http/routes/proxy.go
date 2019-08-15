package routes

import (
	"net/http/httputil"
	"net/url"
	"pkg/errors"

	"github.com/open-falcon/falcon-ng/src/modules/portal/config"

	"github.com/gin-gonic/gin"
)

func transferReq(c *gin.Context) {
	target, err := url.Parse(config.Get().Proxy.Transfer)
	errors.Dangerous(err)

	proxy := httputil.NewSingleHostReverseProxy(target)
	c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))

	proxy.ServeHTTP(c.Writer, c.Request)
}

func indexReq(c *gin.Context) {
	target, err := url.Parse(config.Get().Proxy.Index)
	errors.Dangerous(err)

	proxy := httputil.NewSingleHostReverseProxy(target)
	c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))

	proxy.ServeHTTP(c.Writer, c.Request)
}
