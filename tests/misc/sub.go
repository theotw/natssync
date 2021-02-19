/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	nc1, err := nats.Connect("localhost:4222,localhost:4322", nats.ClosedHandler(func(_ *nats.Conn) {
		fmt.Printf("Connection close  \n")
	}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			fmt.Printf("Connection disconnect %s  \n", err.Error())
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			fmt.Printf("Connection Reconnect  \n")

		}),
	)

	if err != nil {
		log.Fatalf("Error connecting %s", err.Error())
	}
	defer nc1.Close()
	sub, err := nc1.QueueSubscribeSync("testub", "bob")
	//sub, err := nc1.SubscribeSync("testub")
	if err != nil {
		log.Fatalf("Error on subscribe %s \n", err.Error())
	}
	for true {
		msg, err := sub.NextMsg(10 * time.Second)
		if err != nil {
			fmt.Printf("Error on read next %s \n", err.Error())
		} else {
			fmt.Printf("Got a message %s \n", string(msg.Data))
		}
	}
}
