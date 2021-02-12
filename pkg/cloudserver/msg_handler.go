/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
)

var listenForMsgs = true

func StopMessageListener() {
	listenForMsgs = false
}

//Looks for the client ID in the subject string. If not found, return an empty string and error.
func FindClientID(subject string) (string, error) {
	parts := strings.Split(subject, ".")
	if len(parts) > 1 && parts[1] != "" {
		return parts[1], nil
	}
	return "", errors.New(fmt.Sprintf("Unable to parse client ID from '%s'", subject))
}

func logNatsConnectionClosed(name string) {
	log.Infof("Connection %s to NATS has been closed", name)
}

func logNatsConnectionDisconnected(name string, err error) {
	msg := fmt.Sprintf("Connection %s disconnected from NATS", name)
	if err != nil {
		log.Errorf("%s in error: %s", msg, err)
		return
	}
	log.Infof(msg)
}

func logNatsConnectionReconnected(name string) {
	log.Infof("Connection %s reconnected to NATS", name)
}

func NewNatsConnection(name string, url string) (*nats.Conn, error) {
	log.Infof("New connection %s to NATS server %s", name, url)
	return nats.Connect(
		url,
		nats.ClosedHandler(func(_ *nats.Conn) {
			logNatsConnectionClosed(name)
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			logNatsConnectionDisconnected(name, err)
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			logNatsConnectionReconnected(name)
		}),
	)
}

func handleEchoRequest(msg *nats.Msg) {
	echoResponseSubject := fmt.Sprintf("%s.bridge-msg-handler", msg.Reply)
	timeNow := time.Now().Format("20060102-15:04:05.000")
	echoResponseMsg := fmt.Sprintf("%s | %s \n", timeNow, "message-handler")

	nc, err := NewNatsConnection("echo-handler", pkg.Config.NatsServerUrl)
	if err != nil {
		log.Errorf("Error connecting to NATS for echo response: %s", err)
		return
	}
	defer nc.Close()

	if err = nc.Publish(echoResponseSubject, []byte(echoResponseMsg)); err != nil {
		log.Errorf("Error attempting to publish echo response: %s", err)
	}
}

func handleSouthBoundMessage(msg *nats.Msg) {
	metrics.IncrementMessageRecieved(1)

	if strings.HasSuffix(msg.Subject, msgs.ECHO_SUBJECT_BASE) {
		go handleEchoRequest(msg)
	}

	clientID, err := FindClientID(msg.Subject)
	if err != nil {
		log.Error(err)
		return
	}

	plainMsg := new(bridgemodel.NatsMessage)
	plainMsg.Data = msg.Data
	plainMsg.Reply = msg.Reply
	plainMsg.Subject = msg.Subject
	envelope, err := msgs.PutObjectInEnvelope(plainMsg, msgs.CLOUD_ID, clientID)
	log.Tracef("Recieved Message with Client ID %s, Subject %s", clientID, plainMsg.Subject)
	if err != nil {
		log.Errorf("Error putting message in Envelope client ID:%s error=%s", clientID, err)
		return
	}

	jsonData, err := json.Marshal(&envelope)
	if err != nil {
		log.Errorf("Error marshalling envelope with  clientID:%s error=%s", clientID, err)
		return
	}

	cm := new(CachedMsg)
	cm.ClientID = clientID
	cm.Timestamp = time.Now()
	cm.Data = string(jsonData)
	if err = GetCacheMgr().PutMessage(cm); err != nil {
		log.Errorf("Error error storing message: %s", err)
		return
	}
	metrics.IncrementMessagePosted(1)
}

func RunMsgHandler(subject string) {
	var err error
	var nc *nats.Conn
	defer nc.Close()  // This is safe because the Close method is idempotent, and will handle a nil receiver.

	for listenForMsgs {
		nc, err = NewNatsConnection("message-handler", pkg.Config.NatsServerUrl)
		if err != nil {
			log.Errorf("Error connecting to NATS: %s", err)
			// Waiting before retry avoids spamming logs and flooding the network
			time.Sleep(time.Second * 5)
			continue
		}
		_, err = nc.Subscribe(subject, handleSouthBoundMessage)
		if err != nil {
			log.Errorf("Error subscribing to %s: %s", subject, err)
			nc.Close()
			continue
		}
		log.Infof("Subscribed to %s", subject)

		for listenForMsgs && !nc.IsClosed() {
			time.Sleep(time.Second)
		}
	}
}
