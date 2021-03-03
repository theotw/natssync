/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package metrics

import "github.com/prometheus/client_golang/prometheus/promauto"
import "github.com/prometheus/client_golang/prometheus"

var totalQueryForMessages prometheus.Counter
var totalMessagesRecieved prometheus.Counter
var totalValidMessagesPosted prometheus.Counter
var totalClientRegistrationSuccesses prometheus.Counter
var totalClientRegistrationFailures prometheus.Counter
var timeToPushMessage prometheus.Histogram
//sum of all 200 level returns
var httpResp200s prometheus.Counter
//sum of all 400s (including 401 and 404)
var httpResp400s prometheus.Counter
//counter specific for 401 for security
var httpResp401 prometheus.Counter
//counter specific for 404 for health
var httpResp404 prometheus.Counter
var httpResp500 prometheus.Counter

//uses this page https://prometheus.io/docs/guides/go-application/
func InitMetrics() {
	totalQueryForMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_message_query_total",
		Help: "The total number of queries/REST GETs for messages",
	})
	totalMessagesRecieved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_message_received_total",
		Help: "The total number of SB messages received for sending to On Prem clients",
	})
	totalValidMessagesPosted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_message_posted_total",
		Help: "The total number of NB messages posted for sending to Cloud Clients",
	})

	totalClientRegistrationSuccesses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_client_registration_successes",
		Help: "The total number of times client registration succeeded.",
	})
	totalClientRegistrationFailures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_client_registration_failures",
		Help: "The total number of times client registration failed.",
	})
	timeToPushMessage = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "natssync_message_post_time",
		Help: "Time post a message including failed messages",
	})
	httpResp200s = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_http_resp200s",
		Help: "The total number 200 level responses.",
	})
	httpResp400s = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_http_resp400s",
		Help: "The total number of ALL 400 level responses.",
	})
	httpResp401 = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_http_resp401",
		Help: "The total number 401 level responses.",
	})
	httpResp404 = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_http_resp404",
		Help: "The total number 404 level responses.",
	})
	httpResp500 = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_http_resp500s",
		Help: "The total number 500 level responses.",
	})

}

func IncrementTotalQueries(count int) {
	if totalQueryForMessages != nil {
		totalQueryForMessages.Add(float64(count))
	}
}

func IncrementMessageRecieved(count int) {
	if totalMessagesRecieved != nil {
		totalMessagesRecieved.Add(float64(count))
	}
}
func IncrementMessagePosted(count int) {
	if totalValidMessagesPosted != nil {
		totalValidMessagesPosted.Add(float64(count))
	}
}

func IncrementClientRegistrationSuccess(count int) {
	if totalClientRegistrationSuccesses != nil {
		totalClientRegistrationSuccesses.Add(float64(count))
	}
}
func IncrementClientRegistrationFailure(count int) {
	if totalClientRegistrationFailures != nil {
		totalClientRegistrationFailures.Add(float64(count))
	}
}
func RecordTimeToPushMessage(count int) {
	if timeToPushMessage != nil {
		timeToPushMessage.Observe(float64(count))
	}
}
func IncrementHttpResp(statusCode int){
	if statusCode <300{
		httpResp200s.Inc()
	}else if statusCode>399 && statusCode <500{
		httpResp400s.Inc()
		if statusCode ==401{
			httpResp401.Inc()
		}else if statusCode==404{
			httpResp404.Inc()
		}
	}else if statusCode>500{
		httpResp500.Inc()
	}
}
