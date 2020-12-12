package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/msgs"
	"log"
	"os"
	"time"
)

func main() {
	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		log.Fatal(err)
	}
	text := "Hello NATS"
	if len(os.Args) > 1 {
		text = os.Args[1]
	}
	defer nc.Close()
	reply := fmt.Sprintf("astra.%s.%s", msgs.CLOUD_ID, uuid.New().String())
	clientID := pkg.GetEnvWithDefaults("PREM_ID", "client1")
	sub := fmt.Sprintf("astra.%s.echo", clientID)
	nc.PublishRequest(sub, reply, []byte(text))
	nc.Flush()
	sync, err := nc.SubscribeSync(reply)
	if err != nil {
		fmt.Printf("Got Error %s", err.Error())
		os.Exit(1)
	}
	msg, err := sync.NextMsg(1 * time.Minute)
	if err != nil {
		fmt.Printf("Got Error %s", err.Error())
	} else {
		fmt.Printf("Got response %s", string(msg.Data))
	}
}
