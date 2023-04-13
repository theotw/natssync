/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func newRouter(s *server) *gin.Engine {
	router := gin.Default()
	router.Use(loggerMiddleware, MetricsMiddleware)

	router.Any("*urlPath", genericHandlerHandler)

	log.Info("registered routes: ")
	for _, routeInfo := range router.Routes() {
		log.Infof("%s %s", routeInfo.Method, routeInfo.Path)
	}

	return router
}
func newInternalRouter(s *server) *gin.Engine {
	router := gin.Default()
	router.GET("/about", aboutGetUnversioned)
	router.GET("/healthcheck", healthCheckGetUnversioned)
	router.GET("/metrics", metricsHandler)

	log.Info("registered internal routes: ")
	for _, routeInfo := range router.Routes() {
		log.Infof("%s %s", routeInfo.Method, routeInfo.Path)
	}
	return router
}
