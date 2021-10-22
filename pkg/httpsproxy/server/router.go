/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

import (
	"context"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg/httpsproxy"
	models2 "github.com/theotw/natssync/pkg/httpsproxy/models"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var quit chan os.Signal

// Run - configures and starts the web server
func RunHttpProxyServer(test bool) error {

	err := models2.InitNats()
	if err != nil {
		return err
	}

	logLevel := httpproxy.GetEnvWithDefaults("LOG_LEVEL", "debug")
	level, levelerr := log.ParseLevel(logLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)

	nc := models2.GetNatsClient()
	_, err = nc.Subscribe(ResponseForLocationID, func(msg *nats.Msg) {
		locationID := string(msg.Data)
		httpproxy.SetMyLocationID(locationID)
		log.Infof("Using location ID %s", locationID)

	})
	if err != nil {
		log.Fatalf("Unable to talk to NATS %s", err.Error())
	}
	nc.Publish(RequestForLocationID, []byte(""))

	r := newRouter(test)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	log.Infof("Starting server on port 8080")
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	log.Infof("Server Started blocking on channel")
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
	router.Use(loggerMiddleware)
	root := router.Group("/")
	router.NoRoute(func(c *gin.Context) {
		//not sure why I need to do this here and connect is not picked up below, debug it later
		if c.Request.Method == http.MethodConnect {
			connectHandler(c)
		} else {
			c.Status(404)
		}

	})
	router.Handle(http.MethodConnect, "", connectHandler)
	root.Handle(http.MethodGet, "*urlPath", routeHandler)
	root.Handle(http.MethodPost, "*urlPath", routeHandler)
	root.Handle(http.MethodPut, "*urlPath", routeHandler)
	root.Handle(http.MethodPatch, "*urlPath", routeHandler)
	root.Handle(http.MethodHead, "*urlPath", routeHandler)
	root.Handle(http.MethodOptions, "*urlPath", routeHandler)
	root.Handle(http.MethodDelete, "*urlPath", routeHandler)
	root.Handle(http.MethodTrace, "*urlPath", routeHandler)

	return router
}

//router middle ware
func loggerMiddleware(c *gin.Context) {
	clientID := FetchClientIDFromProxyAuthHeader(c)
	log.Debugf("Recieved Request for Client ID:%s of %s %s", clientID, c.Request.Method, c.Request.RequestURI)
	c.Next()
}

const BASIC_AUTH_PREFIX = "Basic "

func FetchClientIDFromProxyAuthHeader(c *gin.Context) string {
	proxyAuth := c.Request.Header.Get("Proxy-Authorization")
	var clientID string
	if strings.HasPrefix(proxyAuth, BASIC_AUTH_PREFIX) {
		b64authInfo := proxyAuth[len(BASIC_AUTH_PREFIX):]
		authInfo, err := base64.StdEncoding.DecodeString(b64authInfo)

		if err != nil {
			log.Errorf("Error decoding auth %s", err.Error())
		} else {
			tmp := string(authInfo)
			split := strings.Split(tmp, ":")
			if len(split) > 0 {
				clientID = split[0]
			}
		}
	}
	return clientID
}
