/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import (
	"bytes"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"io/ioutil"
	"net/http"
	"os"
)

var Quit = make(chan os.Signal, 1)

// Run - configures and starts the web server
func RunBridgeClientRestAPI(test bool) error {

	listenString := pkg.GetEnvWithDefaults("LISTEN_STRING", ":8080")

	r := newRouter(test)
	srv := &http.Server{
		Addr:    listenString,
		Handler: r,
	}

	go func() {
		// service connections
		log.Infof("Starting REST API Server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

	return nil
}

func newRouter(test bool) *gin.Engine {
	router := gin.Default()

	root := router.Group("/bridge-client/")
	root.Handle("GET", "/metrics", metricGetHandlers)
	if test {
		root.Handle("GET", "/kill", func(c *gin.Context) {
			log.Infof("Kill command received, interrupting...")
			Quit <- os.Interrupt
		})
	}
	v1 := router.Group("/bridge-client/1", routeMiddleware)
	v1.Handle("GET", "/about", aboutGetUnversioned)
	v1.Handle("POST", "/unregister", handlePostUnRegister)
	v1.Handle("POST", "/register", handlePostRegister)
	v1.Handle("GET", "/register", registrationGetHandler)
	v1.Handle("GET", "/healthcheck", healthCheckGetUnversioned)
	addUnversionedRoutes(router)
	addOpenApiDefRoutes(router)
	addSwaggerUIRoutes(router)
	addUIRoutes(router)
	return router
}
func addUnversionedRoutes(router *gin.Engine) {
	router.Handle("GET", "/bridge-client/about", aboutGetUnversioned)
	router.Handle("GET", "/bridge-client/healthcheck", healthCheckGetUnversioned)
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
	router.StaticFile("/bridge-client/api/bridge_client_v1.yaml", "openapi/bridge_client_v1.yaml")
	router.StaticFile("/bridge-client/api/swagger.yaml", "openapi/bridge_client_v1.yaml")
}
func addSwaggerUIRoutes(router *gin.Engine) {
	router.Handle("GET", "/bridge-client/api/index.html", swaggerUIGetHandler)
	router.Handle("GET", "/bridge-client/api", swaggerUIGetHandler)
	router.Handle("GET", "/bridge-client/api/", swaggerUIGetHandler)
	swaggerUI := static.LocalFile("third_party/swaggerui/", false)
	webHandler := static.Serve("/bridge-client/api", swaggerUI)
	router.Use(webHandler)
}
func addUIRoutes(router *gin.Engine) {
	//router.Handle("GET", "/onprem-bridge/ui/", uiGetHandler)
	uiFiles := static.LocalFile("webout/", false)
	webHandler := static.Serve("/", uiFiles)
	router.Use(webHandler)
}
