package main

import (
	"flag"
	"log"

	"github.com/nats-io/nats.go"
)

type arguments struct {
	url       *string
	msg       *string
	subject   *string
	subscribe *string
}

func getArguments() arguments {
	args := arguments{
		flag.String("u", "nats://localhost:4222", "URL to use to connect to NATS"),
		flag.String("m", "", "Message to send (requires -s option)"),
		flag.String("s", "", "Subject to send the message (requires -m option)"),
		flag.String("r", "", "Subject to receive messages"),
	}
	flag.Parse()
	if (*args.msg == "" && *args.subject != "") || (*args.msg != "" && *args.subject == "") {
		log.Fatal("Message and subject should both be set, not one or the other.")
	}
	return args
}

func main() {
	args := getArguments()

	nc, err := nats.Connect(*args.url)
	if err != nil {
		log.Fatalf("Unable to connect to NATS: %s", err)
	}
	defer nc.Close()
	log.Printf("Connected to nats at %s", *args.url)

	var sub *nats.Subscription = nil
	var replySubject string

	if *args.subscribe != "" {
		sub, err = nc.Subscribe(*args.subscribe, func(msg *nats.Msg) {
			log.Printf("Received message [%s]: %s\n",msg.Subject, string(msg.Data))
		})
		if err != nil {
			log.Fatalf("Error subscribing: %s", err)
		}
		log.Printf("Subscribed to %s\n", *args.subscribe)
		replySubject = *args.subscribe
	} else {
		replySubject = nats.NewInbox()
	}

	if *args.subject != "" {
		if err = nc.PublishRequest(*args.subject, replySubject, []byte(*args.msg)); err != nil {
			log.Fatalf("Error publishing message: %s", err)
		}
		log.Println("Message published")
	}

	if sub != nil {
		for sub.IsValid() {
			continue
		}
	}
}
