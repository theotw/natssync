/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func newRouter(s *server) *gin.Engine {
	router := gin.Default()
	router.Use(loggerMiddleware)

	router.NoRoute(func(c *gin.Context) {

		if c.Request.Method == http.MethodConnect {
			s.connectHandler(c)
			return
		}

		c.Status(http.StatusNotFound)
	})

	router.Any("*urlPath", s.RouteHandler)

	log.Info("registered routes: ")
	for _, routeInfo := range router.Routes() {
		log.Infof("%s %s", routeInfo.Method, routeInfo.Path)
	}

	return router
}

const (
	BASIC_AUTH_PREFIX = "Basic "
	ProxyAuthHeader   = "Proxy-Authorization"
)

func FetchClientIDFromProxyAuthHeader(c *gin.Context) string {

	proxyAuth := c.Request.Header.Get(ProxyAuthHeader)
	if strings.HasPrefix(proxyAuth, BASIC_AUTH_PREFIX) {
		b64authInfo := proxyAuth[len(BASIC_AUTH_PREFIX):]
		authInfo, err := base64.StdEncoding.DecodeString(b64authInfo)
		if err != nil {
			log.WithError(err).Error("Error decoding auth")
			return ""

		} else {
			split := strings.Split(string(authInfo), ":")
			if len(split) > 0 {
				return split[0]
			}
		}
	}

	log.Errorf("failed to find clientID from '%s' header", ProxyAuthHeader)
	return ""
}
