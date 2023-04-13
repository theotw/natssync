package server

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func loggerMiddleware(c *gin.Context) {

	clientID := GetRouteIDFromAuthHeader(c)
	log.WithFields(log.Fields{
		"method":   c.Request.Method,
		"url":      c.Request.RequestURI,
		"clientID": clientID,
	}).Debugf("Received Request")
	c.Next()

}
