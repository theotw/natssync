package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg/httpsproxy/metrics"
	"github.com/theotw/natssync/pkg/httpsproxy/server/utils"
)

type metricsCollector struct {
	requestCounter float64
}

func isRestrictedHost(host string) bool {
	return false
}

// TODO connect and test this middleware function to verify that the metrics are being collected accurately.
func MetricsMiddleware(c *gin.Context) {
	existingContextWriter := c.Writer
	resWriter := utils.NewResponseWriter()
	c.Writer = resWriter

	// every time the metrics middleware is called => 1 request
	metrics.IncTotalRequests()

	if isRestrictedHost(c.Request.Host) {
		metrics.IncTotalRestrictedIPRequests(c.Request.URL.String(), c.Request.Method)
	} else {
		metrics.IncTotalNonRestrictedIPRequests()
	}

	c.Next()
	metrics.IncTotalFailedRequests(strconv.Itoa(resWriter.GetStatus()))

	// add back the headers
	for key, values := range resWriter.Header() {
		for _, value := range values {
			existingContextWriter.Header().Add(key, value)
		}
	}

	// write the headers
	existingContextWriter.WriteHeader(resWriter.GetStatus())

	// write the body
	if _, err := existingContextWriter.Write(resWriter.GetBody()); err != nil {
		log.WithError(err).Errorf("Metrics Middleware: failed to write body to gin response writer")
	}
}
