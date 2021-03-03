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
	repeate  *int
}

func getECArguments() Arguments {
	args := Arguments{
		flag.String("m", "", "Message to send to the client"),
		flag.String("i", "", "ID of the receiving client"),
		flag.String("u", pkg.Config.NatsServerUrl, "URL to connect to NATS"),
		flag.Int("r", 1, "Number of times to repeate, default is 1 -1 means forever"),
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
	args := getECArguments()

	log.Printf("Connecting to NATS Server %s \n", *args.natsURL)
	err := bridgemodel.InitNats(*args.natsURL, "echo client", 1*time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	nc := bridgemodel.GetNatsConnection()
	defer nc.Close()

	subject := fmt.Sprintf("%s.%s.%s", msgs.SB_MSG_PREFIX, *args.clientID, msgs.ECHO_SUBJECT_BASE)

	i := 0
	done := false
	for !done {
		doping(nc, subject, *args.message)
		i = i + 1
		done = *args.repeate != -1 && i >= *args.repeate
	}
}

func doping(nc *nats.Conn, subject string, message string) {
	var err error
	start := time.Now()
	replySubject := fmt.Sprintf("%s.%s.%s", msgs.NB_MSG_PREFIX, msgs.CLOUD_ID, bridgemodel.GenerateUUID())
	replyListenSub := fmt.Sprintf("%s.*", replySubject)
	sync, err := nc.SubscribeSync(replyListenSub)
	if err != nil {
		log.Fatalf("Error subscribing: %e", err)
	}

	if err = nc.PublishRequest(subject, replySubject, []byte(message)); err != nil {
		log.Fatalf("Error publishing message: %e", err)
	}
	log.Printf("Published message: %s", message)

	if err = nc.Flush(); err != nil {
		log.Fatalf("Error flushing NATS connection: %e", err)
	}

	for {
		msg, err := sync.NextMsg(5 * time.Minute)
		if err != nil {
			log.Printf("Got Error %s \n", err.Error())
			break
		} else {
			fmt.Printf("Message received [%s]: %s \n", msg.Subject, string(msg.Data))
			if strings.HasSuffix(msg.Subject, msgs.ECHOLET_SUFFIX) {
				break
			}
		}
	}
	end := time.Now()
	delta := end.Sub(start)
	fmt.Printf("Total time %d ms \n", delta.Milliseconds())
}
