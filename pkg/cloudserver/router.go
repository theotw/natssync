/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package cloudserver

import (
	"bytes"
	"context"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var quit chan os.Signal

// Run - configures and starts the web server
func RunBridgeServer(test bool) error {
	logLevel := pkg.GetEnvWithDefaults("LOG_LEVEL", "debug")

	level, levelerr := log.ParseLevel(logLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)
	// Initializes database connection

	r := newRouter(test)
	srv := &http.Server{
		Addr:    ":8080",
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
	return nil
}

func newRouter(test bool) *gin.Engine {
	router := gin.Default()
	root := router.Group("/event-bridge/")
	if test {
		root.Handle("GET", "/kill", func(c *gin.Context) {
			quit <- os.Interrupt
		})
	}
	v1 := router.Group("/event-bridge/1", routeMiddleware)
	v1.Handle("GET", "/about", aboutGetUnversioned)
	v1.Handle("POST", "/register", handlePostRegister)
	v1.Handle("POST", "/message-queue/:premid", handlePostMessage)
	v1.Handle("GET", "/message-queue/:premid", handleGetMessages)
	addUnversionedRoutes(router)
	addOpenApiDefRoutes(router)
	addSwaggerUIRoutes(router)
	return router
}
func addUnversionedRoutes(router *gin.Engine) {
	router.Handle("GET", "/event-bridge/about", aboutGetUnversioned)
	router.Handle("GET", "/event-bridge/healthcheck", healthCheckGetUnversioned)
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
}
func addOpenApiDefRoutes(router *gin.Engine) {
	router.StaticFile("/event-bridge/api/cloud_openapi_v1.yaml", "openapi/cloud_openapi_v1.yaml")
	router.StaticFile("/event-bridge/api/swagger.yaml", "openapi/cloud_openapi_v1.yaml")
}
func addSwaggerUIRoutes(router *gin.Engine) {
	router.Handle("GET", "/event-bridge/api/index.html", swaggerUIGetHandler)
	router.Handle("GET", "/event-bridge/api", swaggerUIGetHandler)
	router.Handle("GET", "/event-bridge/api/", swaggerUIGetHandler)
	swaggerUI := static.LocalFile("third_party/swaggerui/", false)
	webHandler := static.Serve("/event-bridge/api", swaggerUI)
	router.Use(webHandler)
}
