package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	labelUrl        = "url"
	labelMethod     = "method"
	labelStatusCode = "statusCode"
)

var (
	// Number of request in per second hitting the system
	totalRequestsPerSecond prometheus.Gauge

	totalRequests prometheus.Counter

	// Number of failed  requests (histogram by error code)
	totalFailedRequests *prometheus.CounterVec

	// Number attempts at restricted IPs. counter since app started
	totalRestrictedIPRequests *prometheus.CounterVec

	// Number of attempts at non-restricted IPs. counter since app started
	totalNonRestrictedIPRequests prometheus.Counter
)

func init() {
	totalRequestsPerSecond = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "httpsproxy_total_requests_per_second",
		Help: "the total number of requests per second hitting the system",
	})

	totalFailedRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "httpsproxy_failed_requests",
		Help: "histogram of all the failed requests",
	}, []string{labelStatusCode})

	totalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "httpsproxy_total_requests",
		Help: "The total number of request hitting the system",
	})

	totalRestrictedIPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "httpsproxy_total_restricted_IP_requests",
		Help: "The total number of requests hitting the system for restricted IP/Domains",
	}, []string{labelUrl, labelMethod})

	totalNonRestrictedIPRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "https_total_non_restricted_IP_requests",
		Help: "The total number of requests hitting the system with a valid IP/Domain",
	})

}

func IncTotalRequests() {
	totalRequests.Inc()
}

func IncTotalRestrictedIPRequests(url, method string) {
	totalRestrictedIPRequests.With(
		prometheus.Labels{
			labelUrl:    url,
			labelMethod: method,
		},
	).Inc()
}

func IncTotalFailedRequests(statusCode string) {
	totalFailedRequests.With(
		prometheus.Labels{
			labelStatusCode: statusCode,
		},
	).Inc()
}

func IncTotalNonRestrictedIPRequests() {
	totalNonRestrictedIPRequests.Inc()
}
