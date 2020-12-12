package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"os"
	"runtime"
)

func main() {
	natsURL := pkg.GetEnvWithDefaults("NATS_SERVER_URL", "nats://127.0.0.1:4222")

	log.Infof("Connecting to NATS server %s", natsURL)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Errorf("Unable to connect to NATS, exiting %s", err.Error())
		os.Exit(2)
	}
	clientID := pkg.GetEnvWithDefaults("PREM_ID", "client1")
	subj := fmt.Sprintf("astra.%s.echo", clientID)

	nc.Subscribe(subj, func(msg *nats.Msg) {
		log.Infof("Got message %s : ", subj, msg.Reply)
		echoMsg := fmt.Sprintf("From %s message=%s \n", clientID, string(msg.Data))
		nc.Publish(msg.Reply, []byte(echoMsg))
		nc.Flush()
	})
	runtime.Goexit()
}
