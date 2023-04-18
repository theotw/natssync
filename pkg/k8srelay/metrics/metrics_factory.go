package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	labelUrl           = "url"
	labelMethod        = "method"
	labelHost          = "host"
	labelStatusCode    = "statusCode"
	serverSystemName   = "httprelayserver"
	proxyletSystemName = "httprelaylet"
)

var (
	// Number of request hitting the system
	totalRequests prometheus.Counter

	// Number of failed  requests (histogram by error code)
	totalFailedRequests *prometheus.CounterVec

	// Number attempts at restricted IPs. counter since app started
	totalRestrictedIPRequests *prometheus.CounterVec

	// Number of attempts at non-restricted IPs. counter since app started
	totalNonRestrictedIPRequests prometheus.Counter
)

func initCommonMetrics(systemName string) {
	totalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_total_requests", systemName),
		Help: "The total number of request hitting the system",
	})

	totalFailedRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_failed_requests", systemName),
		Help: "histogram of all the failed requests",
	}, []string{labelStatusCode})
}

func InitProxyServerMetrics() {
	initCommonMetrics(serverSystemName)
}

func InitProxyletMetrics() {
	initCommonMetrics(proxyletSystemName)

	totalRestrictedIPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_total_restricted_IP_requests", proxyletSystemName),
		Help: "The total number of requests hitting the system for restricted IP/Domains",
	}, []string{labelUrl, labelMethod, labelHost})

	totalNonRestrictedIPRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_total_non_restricted_IP_requests", proxyletSystemName),
		Help: "The total number of requests hitting the system with a valid IP/Domain",
	})
}

func IncTotalRequests() {
	if totalRequests == nil {
		return
	}

	totalRequests.Inc()
}

func IncTotalRestrictedIPRequests(host, url, method string) {
	if totalRestrictedIPRequests == nil {
		return
	}

	totalRestrictedIPRequests.With(
		prometheus.Labels{
			labelUrl:    url,
			labelMethod: method,
			labelHost:   host,
		},
	).Inc()
}

func IncTotalFailedRequests(statusCode string) {
	if totalFailedRequests == nil {
		return
	}

	totalFailedRequests.With(
		prometheus.Labels{
			labelStatusCode: statusCode,
		},
	).Inc()
}

func IncTotalNonRestrictedIPRequests() {
	if totalNonRestrictedIPRequests == nil {
		return
	}

	totalNonRestrictedIPRequests.Inc()
}
