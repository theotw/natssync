/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package metrics

import "github.com/prometheus/client_golang/prometheus/promauto"
import "github.com/prometheus/client_golang/prometheus"

var totalMessagesRecieved prometheus.Counter
var totalValidMessagesPosted prometheus.Counter

//uses this page https://prometheus.io/docs/guides/go-application/
func InitMetrics() {
	totalMessagesRecieved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_message_received_total",
		Help: "The total number of messages received for sending to clients",
	})
	totalValidMessagesPosted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "natssync_message_posted_total",
		Help: "The total number of messages posted for sending to clients",
	})
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
