/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"flag"
	"fmt"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/msgs"
)

type Arguments struct {
	message  *string
	clientID *string
	natsURL  *string
}

func getArguments() Arguments {
	args := Arguments{
		flag.String("msg", "", "Message to send to the client"),
		flag.String("id", pkg.Config.PremId, "ID of the receiving client"),
		flag.String("url", pkg.Config.NatsServerUrl, "URL to connect to NATS"),
	}
	flag.Parse()
	return args
}

//Echo is a special message type.
//its the only message the system looks at and responds to.
//Instead of a single reply, it is expecting multiple replies.
//All the replies will have the "reply" subject as its root, and then a last string
//that indicates which part of the journey has been hit.
//the loop ends when it sees echolet.
func main() {
	args := getArguments()

	log.Printf("Connecting to NATS Server %s \n", *args.natsURL)
	nc, err := nats.Connect(*args.natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	subject := fmt.Sprintf("%s.%s.%s", msgs.SB_MSG_PREFIX, *args.clientID, msgs.ECHO_SUBJECT_BASE)
	replySubject := fmt.Sprintf("%s.%s.%s", msgs.NB_MSG_PREFIX, msgs.CLOUD_ID, bridgemodel.MakeRandomString())
	replyListenSub := fmt.Sprintf("%s.*", replySubject)
	sync, err := nc.SubscribeSync(replyListenSub)
	if err != nil {
		log.Fatalf("Error subscribing: %e", err)
	}
	log.Printf("Subscribed to %s", replyListenSub)

	if err = nc.PublishRequest(subject, replySubject, []byte(*args.message)); err != nil {
		log.Fatalf("Error publishing message: %e", err)
	}
	log.Printf("Published message: %s", *args.message)

	if err = nc.Flush(); err != nil {
		log.Fatalf("Error flushing NATS connection: %e", err)
	}

	for {
		msg, err := sync.NextMsg(5 * time.Minute)
		if err != nil {
			log.Printf("Got Error %s \n", err.Error())
			break
		} else {
			log.Printf("Message received [%s]: %s \n", msg.Subject, string(msg.Data))
			if strings.HasSuffix(msg.Subject, msgs.ECHOLET_SUFFIX) {
				break
			}
		}
	}
}
