package proxylet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/httpsproxy/metrics"
)

func newRouter() *gin.Engine {
	router := gin.Default()

	router.Handle(http.MethodGet, "/metrics", metricGetHandler)

	log.Info("registered routes: ")
	for _, routeInfo := range router.Routes() {
		log.Infof("%s %s", routeInfo.Method, routeInfo.Path)
	}

	return router
}

func metricGetHandler(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

func RunMetricsServer() {

	srv := &http.Server{
		Addr:    pkg.Config.ListenString,
		Handler: newRouter(),
	}

	metrics.InitProxyletMetrics()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.WithError(err).Fatalf("error from listen and serve")
	}

	log.Info("shutting down metrics server")
}
