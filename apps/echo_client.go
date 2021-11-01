/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg/bridgemodel"
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
	log.Infof("Version %s",pkg.VERSION)
	args := getECArguments()

	log.Printf("Connecting to NATS Server %s", *args.natsURL)
	err := bridgemodel.InitNats(*args.natsURL, "echo client", 1*time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	nc := bridgemodel.GetNatsConnection()
	defer nc.Close()
	sync, err := nc.SubscribeSync(bridgemodel.ResponseForLocationID)
	if err != nil {
		log.Fatalf("Unable to subscript to location subject %s", err.Error())
	}
	nc.Publish(bridgemodel.RequestForLocationID, []byte(""))
	msg, err := sync.NextMsg(10 * time.Second)
	var locationID string
	if err != nil {
		log.Errorf("Unable to get location ID from sync client, using cloud master %s", err.Error())
		locationID = msgs.CLOUD_ID
	} else {
		locationID = string(msg.Data)
	}
	log.Infof("Got Location ID to use for reply %s", locationID)

	subject := msgs.MakeEchoSubject(*args.clientID)

	msgs.InitMessageFormat()
	msgFormat := msgs.GetMsgFormat()
	if msgFormat == nil {
		log.Fatalf("Unable to get the message format")
	}

	i := 0
	done := false
	for !done {
		doping(nc, subject, *args.message, locationID)
		i = i + 1
		done = *args.repeate != -1 && i >= *args.repeate
	}
}

func doping(nc *nats.Conn, subject string, message string, locationID string) {
	var err error
	start := time.Now()
	replySubject := msgs.MakeReplySubject(locationID)
	replyListenSub := fmt.Sprintf("%s.*", replySubject)
	sync, err := nc.SubscribeSync(replyListenSub)
	if err != nil {
		log.Fatalf("Error subscribing: %e", err)
	}

	// Add cloud events
	mType := subject
	mSource := "urn:theotw:astra:echolet"
	msgFormat := msgs.GetMsgFormat()
	cvMessage, err := msgFormat.GeneratePayload(message, mType, mSource)
	if err != nil {
		log.Errorf("Failed to generate cloud events payload: %s", err.Error())
		return
	}

	if err = nc.PublishRequest(subject, replySubject, []byte(cvMessage)); err != nil {
		log.Fatalf("Error publishing message: %e", err)
	}
	log.Infof("Published message: %s", cvMessage)

	if err = nc.Flush(); err != nil {
		log.Fatalf("Error flushing NATS connection: %e", err)
	}

	for {
		msg, err := sync.NextMsg(5 * time.Minute)
		if err != nil {
			log.Printf("Got Error %s", err.Error())
			break
		} else {
			fmt.Printf("Message received [%s]: \n", msg.Subject)
			if strings.HasSuffix(msg.Subject, msgs.ECHOLET_SUFFIX) {
				break
			}
		}
	}
	end := time.Now()
	delta := end.Sub(start)
	fmt.Printf("Total time %d ms", delta.Milliseconds())
}
