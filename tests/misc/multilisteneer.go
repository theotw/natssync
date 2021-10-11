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
	nc1, err := nats.Connect("localhost:30220", nats.ClosedHandler(func(_ *nats.Conn) {
		fmt.Printf("Connection close  \n")
	}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("Connection disconnect %s  \n", err.Error())
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			fmt.Printf("Connection Reconnect  \n")

		}),
	)

	if err != nil {
		log.Fatalf("Error connecting %s", err.Error())
	}
	defer nc1.Close()
	_, err = nc1.QueueSubscribe("testpub.1.>", "bob1", func(msg *nats.Msg) {
		fmt.Printf("Got Queue 1 message %s \n", string(msg.Data))
	})

	_, err = nc1.QueueSubscribe("testpub.1.>", "bob1", func(msg *nats.Msg) {
		fmt.Printf("Got Queue 2 message %s \n", string(msg.Data))
	})

	_, err = nc1.Subscribe("testpub.1.>", func(msg *nats.Msg) {
		fmt.Printf("Got non-queue message %s \n", string(msg.Data))
	})

	//sub, err := nc1.SubscribeSync("testub")
	if err != nil {
		log.Fatalf("Error on subscribe %s \n", err.Error())
	}

	_, err = nc1.Subscribe("testpub.>", func(msg *nats.Msg) {
		fmt.Printf("Got Bulk message %s \n", string(msg.Data))
	})

	for i := 0; i < 20; i++ {
		channel := i % 2
		subject := fmt.Sprintf("testpub.%d.appdata", channel)
		nc1.Publish(subject, []byte(fmt.Sprintf("%d", i)))
		nc1.Flush()
	}
	time.Sleep(5 * time.Second)

}
