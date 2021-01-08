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

//uses this page https://prometheus.io/docs/guides/go-application/
func InitMetrics() {
	totalQueryForMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_message_query_total",
		Help: "The total number of queries for messages",
	})
	totalMessagesRecieved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_message_received_total",
		Help: "The total number of messages received for sending to clients",
	})
	totalValidMessagesPosted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_message_posted_total",
		Help: "The total number of messages posted for sending to clients",
	})

	totalQueuedMessages = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "natssync_queue_messages_total",
		Help: "Total messages queued for retrieval",
	})
	timeWaitingForMessages = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "natssync_retrieve_time",
		Help: "Time waiting for messages",
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
