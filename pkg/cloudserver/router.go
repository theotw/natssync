/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"bytes"
	"context"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/metrics"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

)

var quit chan os.Signal

// Run - configures and starts the web server
func RunBridgeServer(test bool) {
	level, levelerr := log.ParseLevel(pkg.Config.LogLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)
	// Initializes database connection

	r := newRouter(test)
	srv := &http.Server{
		Addr:    pkg.Config.ListenString,
		Handler: r,
	}

	go func() {
		// service connections
		log.Info("In goroutine list and server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
		log.Info("Post In goroutine list and server")
	}()
	log.Info("Web Server running")
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit = make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func newRouter(test bool) *gin.Engine {
	router := gin.Default()
	root := router.Group("/")
	root.Handle("GET", "/metrics", metricGetHandlers)
	bidgeRoot := router.Group("/bridge-server/")
	if test {
		bidgeRoot.Handle("GET", "/kill", func(c *gin.Context) {
			quit <- os.Interrupt
		})
	}
	v1 := router.Group("/bridge-server/1", routeMiddleware)
	v1.Handle("GET", "/about", aboutGetUnversioned)
	v1.Handle("GET", "/healthcheck", healthCheckGetUnversioned)
	v1.Handle("POST", "/register", handlePostRegister)
	v1.Handle("POST", "/message-queue/:premid", handlePostMessage)
	v1.Handle("GET", "/message-queue/:premid", handleGetMessages)
	v1.Handle("POST", "/messages", natsMsgPostHandler)

	addUnversionedRoutes(router)
	addOpenApiDefRoutes(router)
	addSwaggerUIRoutes(router)
	return router
}
func addUnversionedRoutes(router *gin.Engine) {
	router.Handle("GET", "/bridge-server/about", aboutGetUnversioned)
	router.Handle("GET", "/bridge-server/healthcheck", healthCheckGetUnversioned)
}

//router middle ware
func routeMiddleware(c *gin.Context) {
	content := c.Request.Header.Get("Content-Type")

	if content == "application/json" && c.Request.ContentLength < 2048 {
		bits, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Tracef("error reading body %s", err.Error())
		} else {
			log.Tracef("%s", string(bits))
		}

		c.Request.Body = ioutil.NopCloser(bytes.NewReader(bits))
	}

	c.Next()
	if c.Writer != nil {
 		metrics.IncrementHttpResp(c.Writer.Status())
	}
}
func addOpenApiDefRoutes(router *gin.Engine) {
	router.StaticFile("/bridge-server/api/bridge_server_v1.yaml", "openapi/bridge_server_v1.yaml")
	router.StaticFile("/bridge-server/api/swagger.yaml", "openapi/bridge_server_v1.yaml")
}
func addSwaggerUIRoutes(router *gin.Engine) {
	router.Handle("GET", "/bridge-server/api/index.html", swaggerUIGetHandler)
	router.Handle("GET", "/bridge-server/api", swaggerUIGetHandler)
	router.Handle("GET", "/bridge-server/api/", swaggerUIGetHandler)
	swaggerUI := static.LocalFile("third_party/swaggerui/", false)
	webHandler := static.Serve("/bridge-server/api", swaggerUI)
	router.Use(webHandler)
}
