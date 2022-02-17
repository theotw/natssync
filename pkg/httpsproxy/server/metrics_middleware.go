package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg/httpsproxy/metrics"
	"github.com/theotw/natssync/pkg/httpsproxy/server/utils"
)

func MetricsMiddleware(c *gin.Context) {
	log.WithFields(
		log.Fields{
			"host":   c.Request.Host,
			"uri":    c.Request.RequestURI,
			"method": c.Request.Method,
		},
	).Debug("metrics middleware info" )

	existingContextWriter := c.Writer
	resWriter := utils.NewResponseWriter(existingContextWriter)
	c.Writer = resWriter

	// every time the metrics middleware is called => 1 request
	metrics.IncTotalRequests()

	c.Next()
	if resWriter.Status() >= 300 {
		metrics.IncTotalFailedRequests(strconv.Itoa(resWriter.GetStatus()))
	}

	//resWriter.WriteDataOut()
}
