/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/msgs"
)

//The client/south side echo proxylet.  Answers echo calls
var clientID string
var activeSub *nats.Subscription

func main() {
	log.Infof("Version %s", pkg.VERSION)
	level, levelerr := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	clientID = "*"
	natsURL := pkg.Config.NatsServerUrl
	log.Infof("Connecting to NATS server %s", natsURL)
	err := bridgemodel.InitNats(natsURL, "echo Main", 1*time.Minute)
	if err != nil {
		log.Errorf("Unable to connect to NATS, exiting %s", err.Error())
		os.Exit(2)

	}
	msgs.InitMessageFormat()
	msgFormat := msgs.GetMsgFormat()
	if msgFormat == nil {
		log.Fatalf("Unable to get the message format")
	}

	nc := bridgemodel.GetNatsConnection()
	nc.Subscribe(bridgemodel.ResponseForLocationID, func(msg *nats.Msg) {
		clientID = string(msg.Data)
		if activeSub != nil {
			activeSub.Unsubscribe()
		}
		subj := fmt.Sprintf("%s.%s.%s", msgs.NATSSYNC_MESSAGE_PREFIX, clientID, msgs.ECHO_SUBJECT_BASE)
		activeSub, err = nc.Subscribe(subj, msgHandler)
		if err != nil {
			log.Errorf("Unabel to scubscript %s, %s", err.Error(), subj)
		} else {
			log.Infof("Subscribed for %s ", clientID)
		}

	})
	nc.Publish(bridgemodel.RequestForLocationID, []byte(""))

	runtime.Goexit()
}
func msgHandler(msg *nats.Msg) {
	log.Infof("msgHandler: Got message %s : %s  %s", msg.Subject, msg.Reply, msg.Data)
	tmpstring := time.Now().Format("20060102-15:04:05.000")
	echoMsg := fmt.Sprintf("%s | %s %s %s", tmpstring, "echoproxylet", clientID, string(msg.Data))
	replysub := fmt.Sprintf("%s.%s", msg.Reply, msgs.ECHOLET_SUFFIX)
	mType := msg.Subject
	mSource := "urn:theotw:astra:echolet"
	msgFormat := msgs.GetMsgFormat()
	cvMessage, err := msgFormat.GeneratePayload(echoMsg, mType, mSource)
	if err != nil {
		log.Errorf("Failed to generate cloud events payload: %s", err.Error())
		return
	}
	log.Infof("msgHandler: geenrated payload cvMessage %v", cvMessage)
	nc := bridgemodel.GetNatsConnection()
	log.Infof("publishing %v to %s", cvMessage, replysub)
	if err = nc.Publish(replysub, cvMessage); err != nil {
		log.Errorf("Error publishing to %s: %s", replysub, err)
	}
	_ = nc.Flush()
}
