/*
 * Copyright (c) The One True Way 2021. Apache License 2.0.
 * The authors accept no liability, 0 nada for the use of this software.
 * It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/msgs"
)

type Arguments struct {
	message *string
	clientID *string
	natsURL *string
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

	fmt.Printf("Using NATS Server %s \n", *args.natsURL)
	nc, err := nats.Connect(*args.natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	randomClientUUID := uuid.New().String()
	reply := fmt.Sprintf("%s.%s.%s", msgs.NB_MSG_PREFIX, msgs.CLOUD_ID, randomClientUUID)

	sub := fmt.Sprintf("%s.%s.%s", msgs.SB_MSG_PREFIX, *args.clientID, msgs.ECHO_SUBJECT_BASE)
	replyListenSub := fmt.Sprintf("%s.%s.%s.*", msgs.NB_MSG_PREFIX, msgs.CLOUD_ID, randomClientUUID)
	sync, err := nc.SubscribeSync(replyListenSub)
	if err != nil {
		log.Fatalf("Error subscribing: %e", err)
	}

	if err = nc.PublishRequest(sub, reply, []byte(*args.message)); err != nil {
		log.Fatalf("Error publishing message: %e", err)
	}
	if err = nc.Flush(); err != nil {
		log.Fatalf("Error flushing NATS connection: %e", err)
	}

	for {
		msg, err := sync.NextMsg(5 * time.Minute)
		if err != nil {
			fmt.Printf("Got Error %s \n", err.Error())
			break
		} else {
			fmt.Printf(" %s \n", string(msg.Data))
			if strings.HasSuffix(msg.Subject, msgs.ECHOLET_SUFFIX) {
				break
			}
		}
	}
}
