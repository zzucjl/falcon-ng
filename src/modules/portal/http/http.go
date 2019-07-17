package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var srv = &http.Server{
	Addr: ":8080",
}

// Start http server
func Start() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/sleep", func(c *gin.Context) {
		time.Sleep(3 * time.Second)
		c.JSON(200, gin.H{
			"message": "sleep",
		})
	})

	r.GET("/sleep2", func(c *gin.Context) {
		time.Sleep(6 * time.Second)
		c.JSON(200, gin.H{
			"message": "sleep2",
		})
	})

	srv.Handler = r

	go func() {
		// service connections
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
