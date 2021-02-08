/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/msgs"
	"os"
	"runtime"
	"time"
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
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Errorf("Unable to connect to NATS, exiting %s", err.Error())
		os.Exit(2)
	}

	subj := fmt.Sprintf("%s.%s.%s", msgs.SB_MSG_PREFIX, clientID, msgs.ECHO_SUBJECT_BASE)

	nc.Subscribe(subj, func(msg *nats.Msg) {
		log.Infof("Got message %s : ", subj, msg.Reply)
		echoMsg := fmt.Sprintf("%s From %s message=%s \n", time.Now().String(), clientID, string(msg.Data))
		replysub := fmt.Sprintf("%s.%s", msg.Reply, msgs.ECHOLET_SUFFIX)
		nc.Publish(replysub, []byte(echoMsg))
		nc.Flush()
	})
	runtime.Goexit()
}
