package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
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
	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		log.Fatal(err)
	}
	text := "Hello NATS"
	if len(os.Args) < 3 {
		panic("Must have 2 arguments <target localtionID> <message>")
	}
	defer nc.Close()
	reply := fmt.Sprintf("%s.%s.%s.*", msgs.NB_MSG_PREFIX, msgs.CLOUD_ID, uuid.New().String())
	clientID := os.Args[1]
	text = os.Args[2]
	sub := fmt.Sprintf("%s.%s.%s", msgs.SB_MSG_PREFIX, clientID,msgs.ECHO_SUBJECT_BASE)
	nc.PublishRequest(sub, reply, []byte(text))
	nc.Flush()
	sync, err := nc.SubscribeSync(reply)
	if err != nil {
		fmt.Printf("Got Error %s", err.Error())
		os.Exit(1)
	}
	for ; true; {
		msg, err := sync.NextMsg(1 * time.Minute)
		if err != nil {
			fmt.Printf("Got Error %s", err.Error())
			break
		} else {
			fmt.Printf("Got response %s", string(msg.Data))
			if strings.HasSuffix(msg.Subject,"proxylet"){
				break
			}
		}

	}
}
