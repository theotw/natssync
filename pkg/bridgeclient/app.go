/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/msgs"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func RunClient(test bool) {
	level, levelerr := log.ParseLevel(pkg.Config.LogLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)
	msgs.InitLocationKeyStore()
	err := RunBridgeClient(test)
	if err != nil {
		log.Errorf("Error starting API server %s", err.Error())
		os.Exit(1)
	}

	serverURL := pkg.Config.CloudBridgeUrl
	clientID := pkg.Config.PremId
	if len(clientID) == 0 {
		log.Errorf("No client ID, exiting")
		os.Exit(2)
	}
	var nc *nats.Conn

	for true {
		if nc == nil {
			nc, err = nats.Connect(pkg.Config.NatsServerUrl)
			if err != nil {
				log.Errorf("Unable to connect to NATS, retrying... error: %s", err.Error())
				nc = nil
				time.Sleep(10 * time.Second)
				continue
			} else {
				subj := fmt.Sprintf("%s.%s.>", msgs.NB_MSG_PREFIX, msgs.CLOUD_ID)
				nc.Subscribe(subj, func(msg *nats.Msg) {
					log.Debugf("Sending msg to cloud %s", msg.Subject)
					sendMessageToCloud(msg, serverURL, clientID)
				})
			}
		} else {
			url := fmt.Sprintf("%s/event-bridge/1/message-queue/%s", serverURL, clientID)
			resp, err := http.DefaultClient.Get(url)
			if err != nil {
				log.Errorf("Error fetching messages %s", err.Error())
				time.Sleep(2 * time.Second)
				continue
			}
			if resp.StatusCode >= 300 {
				log.Errorf("Error code fetching messages %s", resp.Status)
				time.Sleep(2 * time.Second)
				continue
			}
			bits, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("Error reading messages %s", err.Error())
				continue
			}
			var msglist []v1.BridgeMessage
			err = json.Unmarshal(bits, &msglist)
			if err != nil {
				log.Errorf("Error unmarshalling messages %s", err.Error())
				continue
			}
			for _, m := range msglist {
				var env msgs.MessageEnvelope
				err = json.Unmarshal([]byte(m.MessageData), &env)
				if err != nil {
					log.Errorf("Error unmarshalling envelope %s", err.Error())
					continue
				}

				var natmsg bridgemodel.NatsMessage
				err := msgs.PullObjectFromEnvelope(&natmsg, &env)
				if err != nil {
					log.Errorf("Error decoding envelope %s", err.Error())
					continue
				}
				if len(natmsg.Reply) > 0 {
					if strings.HasSuffix(natmsg.Subject, msgs.ECHO_SUBJECT_BASE) {
						var echomsg nats.Msg
						echomsg.Subject = fmt.Sprintf("%s.bridge-client", natmsg.Reply)
						echomsg.Data = []byte(time.Now().String() + " bridge client")
						go sendMessageToCloud(&echomsg, serverURL, clientID)
					}

					nc.PublishRequest(natmsg.Subject, natmsg.Reply, natmsg.Data)
				} else {
					nc.Publish(natmsg.Subject, natmsg.Data)
				}
				fmt.Println(natmsg)
			}
		}

	}

}

func sendMessageToCloud(msg *nats.Msg, serverURL string, clientID string) {
	url := fmt.Sprintf("%s/event-bridge/1/message-queue/%s", serverURL, clientID)
	natmsg := bridgemodel.NatsMessage{Reply: msg.Reply, Subject: msg.Subject, Data: msg.Data}
	envelope, enverr := msgs.PutObjectInEnvelope(natmsg, clientID, msgs.CLOUD_ID)
	if enverr != nil {
		log.Errorf("Error putting msg in envelope %s", enverr.Error())
		return
	}
	jsonbits, jsonerr := json.Marshal(envelope)
	if jsonerr != nil {
		log.Errorf("Error encoding envelope to json bits %s", jsonerr.Error())
		return
	}
	bmsg := v1.BridgeMessage{ClientID: clientID, MessageData: string(jsonbits), FormatVersion: "1"}
	bmsgBits, bmsgerr := json.Marshal(bmsg)
	if bmsgerr != nil {
		log.Errorf("Error marshaling bridge message to json bits %s", jsonerr.Error())
		return
	}

	r := bytes.NewReader(bmsgBits)
	resp, posterr := http.DefaultClient.Post(url, "application/json", r)
	if posterr != nil {
		log.Errorf("Error sending message to server.  Dropping the message %s  error was %s", msg.Subject, posterr.Error())
		return
	}
	if resp.StatusCode >= 300 {
		log.Errorf("Error sending message to server.  Dropping the message %s  error was %s", msg.Subject, resp.Status)
	}
}
