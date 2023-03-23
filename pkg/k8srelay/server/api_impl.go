/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"
)

type JsonError struct {
	Message string `json:"message"`
}

func HandleError(_ *gin.Context, err error) (int, interface{}) {
	ret := JsonError{
		Message: err.Error(),
	}

	return http.StatusInternalServerError, &ret
}

func (s *server) RouteHandler(c *gin.Context) {
	//TODO THIS IS WEIRD NEED TO FIGURE OUT IF WE NEED THIS
	clientID := FetchClientIDFromProxyAuthHeader(c)

	if clientID == "" {

		//normalize out the string
		tmp := c.Request.RequestURI
		if !strings.HasSuffix(tmp, "/") {
			tmp = fmt.Sprintf("%s/", tmp)
		}

		if strings.HasSuffix(tmp, "/about/") {
			aboutGetUnversioned(c)
			return
		}

		if strings.HasSuffix(tmp, "/healthcheck/") {
			healthCheckGetUnversioned(c)
			return
		}

		if strings.HasSuffix(tmp, "/metrics/") {
			metricsHandler(c)
			return
		}

		c.JSON(http.StatusServiceUnavailable, "")
		return
	}

}

type AboutResponse struct {
	AppVersion string `json:"appVersion"`
}

func aboutGetUnversioned(c *gin.Context) {
	var resp AboutResponse
	resp.AppVersion = "1.0.0"

	c.JSON(http.StatusOK, &resp)
}

func healthCheckGetUnversioned(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func metricsHandler(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

func genericHandlerHandler(c *gin.Context) {
	relaylet.DoCall(c)
}
