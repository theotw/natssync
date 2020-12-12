/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package cloudclient

import (
	"bytes"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
)

var quit chan os.Signal

// Run - configures and starts the web server
func RunBridgeClient(test bool) error {

	r := newRouter(test)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return nil
}

func newRouter(test bool) *gin.Engine {
	router := gin.Default()
	root := router.Group("/onprem-bridge/")
	if test {
		root.Handle("GET", "/kill", func(c *gin.Context) {
			quit <- os.Interrupt
		})
	}
	v1 := router.Group("/onprem-bridge/1", routeMiddleware)
	v1.Handle("GET", "/about", aboutGetUnversioned)
	v1.Handle("POST", "/register", handlePostRegister)
	v1.Handle("GET", "/register", handleGetRegister)
	addUnversionedRoutes(router)
	addOpenApiDefRoutes(router)
	addSwaggerUIRoutes(router)
	addUIRoutes(router)
	return router
}
func addUnversionedRoutes(router *gin.Engine) {
	router.Handle("GET", "/onprem-bridge/about", aboutGetUnversioned)
	router.Handle("GET", "/onprem-bridge/healthcheck", healthCheckGetUnversioned)
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
	router.StaticFile("/onprem-bridge/api/onprem_openapi_v1.yaml", "openapi/onprem_openapi_v1.yaml")
	router.StaticFile("/onprem-bridge/api/swagger.yaml", "openapi/onprem_openapi_v1.yaml")
}
func addSwaggerUIRoutes(router *gin.Engine) {
	router.Handle("GET", "/onprem-bridge/api/index.html", swaggerUIGetHandler)
	router.Handle("GET", "/onprem-bridge/api", swaggerUIGetHandler)
	router.Handle("GET", "/onprem-bridge/api/", swaggerUIGetHandler)
	swaggerUI := static.LocalFile("third_party/swaggerui/", false)
	webHandler := static.Serve("/onprem-bridge/api", swaggerUI)
	router.Use(webHandler)
}
func addUIRoutes(router *gin.Engine) {
	//router.Handle("GET", "/onprem-bridge/ui/", uiGetHandler)
	uiFiles := static.LocalFile("webout/", false)
	webHandler := static.Serve("/", uiFiles)
	router.Use(webHandler)
}
