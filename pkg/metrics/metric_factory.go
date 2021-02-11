/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package metrics

import "github.com/prometheus/client_golang/prometheus/promauto"
import "github.com/prometheus/client_golang/prometheus"

var totalQueryForMessages prometheus.Counter
var totalMessagesRecieved prometheus.Counter
var totalValidMessagesPosted prometheus.Counter
var totalQueuedMessages prometheus.Gauge
var timeWaitingForMessages prometheus.Histogram
var totalClientRegistrationSuccesses prometheus.Counter
var totalClientRegistrationFailures prometheus.Counter
var timeToPushMessage prometheus.Histogram
var oldestMessageQueued prometheus.Histogram

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

	totalQueuedMessages = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "natssync_queue_messages_total",
		Help: "Total messages queued for retrieval",
	})
	timeWaitingForMessages = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "natssync_retrieve_time",
		Help: "Time waiting for messages",
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
	oldestMessageQueued = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "natssync_time_message_queued",
		Help: "Age of the oldest message in the cache",
	})
}

func IncrementTotalQueries(count int) {
	if totalQueryForMessages != nil {
		totalQueryForMessages.Add(float64(count))
	}
}
func AddCountToWaitTimes(count int) {
	if timeWaitingForMessages != nil {
		timeWaitingForMessages.Observe(float64(count))
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
func SetTotalMessagesQueued(count int) {
	if totalQueuedMessages != nil {
		totalQueuedMessages.Set(float64(count))
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
func RecordAgeOfMessageQueue(count int) {
	if oldestMessageQueued != nil {
		oldestMessageQueued.Observe(float64(count))
	}
}
