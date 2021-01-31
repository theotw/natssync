/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package l2

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
)

//tests nats cluster by sending a message to nats0 and listening for it on nats1
func TestNATSClustering(t *testing.T) {

	nats0 := os.Getenv("nats0")
	nats1 := os.Getenv("nats1")
	if !assert.Greater(t, len(nats0), 0, "need an env var called nats0") {
		t.Fail()
	}
	if !assert.Greater(t, len(nats1), 0, "need an env var called nats1") {
		t.Fail()
	}
	nc0, err := nats.Connect(nats0)
	if err != nil {
		t.Fatal(err)
	}
	defer nc0.Close()

	nc1, err := nats.Connect(nats1)
	if err != nil {
		t.Fatal(err)
	}
	defer nc1.Close()

	// Use a WaitGroup to wait for a message to arrive
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Subscribe
	if _, err := nc1.Subscribe("*", func(m *nats.Msg) {
		fmt.Printf("%s.%s", m.Subject, m.Data)
		wg.Done()
	}); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Publishing")
	if err := nc0.Publish("updates", []byte("All is Well")); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Done Publishing")

	wg.Wait()

	fmt.Println("Done wait")
	assert.True(t, true, true)
}
