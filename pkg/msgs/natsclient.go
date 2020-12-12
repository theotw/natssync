/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package msgs

import (
	"fmt"
	nats "github.com/nats-io/nats.go"
	"log"
	"sync"
)

func ConnectToNats() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Use a WaitGroup to wait for a message to arrive
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Subscribe
	if _, err := nc.Subscribe("*", func(m *nats.Msg) {
		fmt.Printf("%s.%s", m.Subject, m.Data)
		wg.Done()
	}); err != nil {
		log.Fatal(err)
	}

	if err := nc.Publish("updates", []byte("All is Well")); err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}
