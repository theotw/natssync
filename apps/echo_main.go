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
func main() {
	var clientID string
	if len(os.Args) != 2 {
		log.Infof("No client ID provided, defaulting to *")
		clientID = "*"

	} else {
		clientID = os.Args[1]
	}
	natsURL := pkg.Config.NatsServerUrl
	log.Infof("Connecting to NATS server %s", natsURL)
	err := bridgemodel.InitNats(natsURL, "echo Main", 1*time.Minute)
	if err != nil {
		log.Errorf("Unable to connect to NATS, exiting %s", err.Error())
		os.Exit(2)

	}
	nc := bridgemodel.GetNatsConnection()

	subj := fmt.Sprintf("%s.%s.%s", msgs.NATSSYNC_MESSAGE_PREFIX, clientID, msgs.ECHO_SUBJECT_BASE)

	msgs.InitMessageFormat()
	msgFormat := msgs.GetMsgFormat()
	if msgFormat == nil {
		log.Fatalf("Unable to get the message format")
	}

	_, err = nc.Subscribe(subj, func(msg *nats.Msg) {
		log.Infof("Got message %s : %s  %s", subj, msg.Reply, msg.Data)
		tmpstring := time.Now().Format("20060102-15:04:05.000")
		echoMsg := fmt.Sprintf("%s | %s %s %s", tmpstring, "echoproxylet", clientID, string(msg.Data))
		replysub := fmt.Sprintf("%s.%s", msg.Reply, msgs.ECHOLET_SUFFIX)
		mType := subj
		mSource := "urn:theotw:astra:echolet"
		msgFormat := msgs.GetMsgFormat()
		cvMessage, err := msgFormat.GeneratePayload(echoMsg, mType, mSource)

		if err != nil {
			log.Errorf("Failed to generate cloud events payload: %s", err.Error())
			return
		}
		if err = nc.Publish(replysub, cvMessage); err != nil {
			log.Errorf("Error publishing to %s: %s", replysub, err)
		}
		_ = nc.Flush()
	})

	if err != nil {
		log.Errorf("Error subscribing to %s: %s", subj, err)
	}

	runtime.Goexit()
}
