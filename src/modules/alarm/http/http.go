package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/open-falcon/falcon-ng/src/modules/alarm/config"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/http/middleware"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/http/routes"
)

var srv = &http.Server{
	ReadTimeout:    10 * time.Second,
	WriteTimeout:   10 * time.Second,
	MaxHeaderBytes: 1 << 20,
}

// Start http server
func Start() {
	c := config.GetCfgYml()

	loggerMid := middleware.LoggerWithConfig(middleware.LoggerConfig{})
	recoveryMid := middleware.Recovery()

	if c.Logger.Level != "DEBUG" {
		gin.SetMode(gin.ReleaseMode)
		middleware.DisableConsoleColor()
	}

	r := gin.New()
	r.Use(loggerMid, recoveryMid)

	routes.Config(r)

	srv.Addr = c.HTTP.Listen
	srv.Handler = r

	go func() {
		log.Println("starting http server, listening on:", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listening %s occur error: %s\n", srv.Addr, err)
		}
	}()
}

// Shutdown http server
func Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalln("cannot shutdown http server:", err)
	}

	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Println("shutdown http server timeout of 5 seconds.")
	default:
		log.Println("http server stopped")
	}
}
