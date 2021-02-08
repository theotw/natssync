/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
	"strings"
	"time"
)

var listenForMsgs = true

func StopMessageListener() {
	listenForMsgs = false
}

//Looks for the client ID in the subject string. If not found, return an empty string.
func FindClientID(subject string) string {
	parts := strings.Split(subject, ".")
	var ret string
	if len(parts) > 1 {
		ret = parts[1]
	}
	return ret
}

func RunMsgHandler(subjectString string) {
	for listenForMsgs {

		natsURL := pkg.Config.NatsServerUrl
		log.Infof("Connecting to NATS server %s", natsURL)
		nc, err := nats.Connect(natsURL)
		if err == nil {
			log.Infof("Connected to NATS server %s", natsURL)
			subscription, err := nc.SubscribeSync(subjectString)
			if err == nil {
				for listenForMsgs {
					m, err := subscription.NextMsg(10 * time.Second)
					metrics.IncrementMessageRecieved(1)
					if err == nil {
						if strings.HasSuffix(m.Subject, msgs.ECHO_SUBJECT_BASE) {
							echosub := fmt.Sprintf("%s.bridge-msg-handler", m.Reply)
							nc.Publish(echosub, []byte(time.Now().String()+" message handler"))
						}
						clientID := FindClientID(m.Subject)
						if len(clientID) != 0 {
							cm := new(CachedMsg)
							cm.ClientID = clientID
							plainMsg := new(bridgemodel.NatsMessage)
							plainMsg.Data = m.Data
							plainMsg.Reply = m.Reply
							plainMsg.Subject = m.Subject
							envelope, err2 := msgs.PutObjectInEnvelope(plainMsg, msgs.CLOUD_ID, clientID)
							log.Tracef("Recieved Message with Client ID %s, Subject %s", clientID, plainMsg.Subject)
							if err2 != nil {
								log.Errorf("Error putting message in Envelope client ID:%s error=%s", clientID, err2.Error())
							} else {
								jsonData, marshelError := json.Marshal(&envelope)
								if marshelError != nil {
									log.Errorf("Error marshalling envelope with  clientID:%s error=%s", clientID, marshelError.Error())
								} else {
									cm.Data = string(jsonData)
									GetCacheMgr().PutMessage(cm)
									metrics.IncrementMessagePosted(1)
								}
							}
						}
					} else if strings.Index(err.Error(), "timeout") < 0 {
						log.Infof("Failed to GetNextMsg to NATS server %s, pausing  %d", natsURL, 10)
						time.Sleep(10 * time.Second)
						break
					}
				}
			} else {
				//if not joy, back off and try again in 10 seconds
				log.Infof("Failed to Subscribe to NATS server %s, pausing  %d", natsURL, 10)
				time.Sleep(10 * time.Second)
			}

			nc.Close()
		} else {
			var timeout time.Duration
			timeout = 10
			log.Infof("Failed to connect to NATS server %s, pausing  %d", natsURL, timeout)
			//if not joy, back off and try again in 10 seconds
			time.Sleep(timeout * time.Second)
		}

	}

	return
}
