/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/msgs"
	"log"
	"os"
	"strings"
	"time"
)

//Echo is a secial message type.
//its the only message the system looks at and responds to.
//Instead of a single reply, it is expecting multiple replies.  All the replies will have the "reply" subject as its root, and then a last string
//that indicates which part of the journey has been hit.
// the loop ends when it sees echolet.
func main() {
	naturl := pkg.GetEnvWithDefaults("NATS_SERVER_URL", "nats://127.0.0.1:4222")
	fmt.Printf("Using NATS Server %s \n", naturl)
	nc, err := nats.Connect(naturl)
	if err != nil {
		log.Fatal(err)
	}
	text := "Hello NATS"
	if len(os.Args) < 3 {
		panic("Must have 2 arguments <target localtionID> <message>")
	}
	defer nc.Close()
	randomClientUUID := uuid.New().String()
	reply := fmt.Sprintf("%s.%s.%s", msgs.NB_MSG_PREFIX, msgs.CLOUD_ID, randomClientUUID)
	clientID := os.Args[1]
	text = os.Args[2]
	sub := fmt.Sprintf("%s.%s.%s", msgs.SB_MSG_PREFIX, clientID, msgs.ECHO_SUBJECT_BASE)
	replyListenSub := fmt.Sprintf("%s.%s.%s.*", msgs.NB_MSG_PREFIX, msgs.CLOUD_ID, randomClientUUID)
	sync, err := nc.SubscribeSync(replyListenSub)

	nc.PublishRequest(sub, reply, []byte(text))
	nc.Flush()
	if err != nil {
		fmt.Printf("Got Error %s", err.Error())
		os.Exit(1)
	}
	for true {
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
