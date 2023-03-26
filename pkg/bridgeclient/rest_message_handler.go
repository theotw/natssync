/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
	"github.com/theotw/natssync/pkg/natsmodel"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RestMessageHandler A rest based implementation of the BiDiMessageHanadler
type RestMessageHandler struct {
	serverURL           string
	stopFlag            bool
	currentSubscription *nats.Subscription
}

func NewRestMessageHandler(serverURL string) *RestMessageHandler {
	ret := new(RestMessageHandler)
	ret.serverURL = serverURL
	ret.stopFlag = false
	return ret
}
func (t *RestMessageHandler) GetHandlerType() string {
	return "rest"
}
func (t *RestMessageHandler) StartMessageHandler(clientID string) error {
	currentSubscription, err := subscribeToOutboundMessages(t.serverURL, clientID)
	if err != nil {
		log.Errorf("Error subscribing to messages, will try again %s", err.Error())
	}
	t.currentSubscription = currentSubscription
	go t.pullMessageFromCloud(clientID)
	return nil
}
func (t *RestMessageHandler) StopMessageHandler() {
	t.stopFlag = true
	t.currentSubscription.Unsubscribe()
}
func (t *RestMessageHandler) pullMessageFromCloud(clientID string) {
	for !t.stopFlag {
		msglist, err := getMessagesFromCloud(t.serverURL, clientID)
		if err != nil {
			log.Errorf("Error fetching messages %s", err.Error())
			time.Sleep(2 * time.Second)
			continue
		}
		log.Infof("Received %d messages from server", len(msglist))

		for _, m := range msglist {
			nc := natsmodel.GetNatsConnection()
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
			msgFormat := msgs.GetMsgFormat()
			status, err := msgFormat.ValidateMsgFormat(natmsg.Data, pkg.Config.CloudEvents)
			if err != nil {
				log.Errorf("Error validating the cloud event message: %s", err.Error())
				continue
			}
			if !status {
				log.Errorf("Cloud event message validation failed, ignoring the message...")
				continue
			}
			log.Infof("Received message: sub=%s reply=%s", natmsg.Subject, natmsg.Reply)

			if len(natmsg.Reply) > 0 {
				if strings.HasSuffix(natmsg.Subject, msgs.ECHO_SUBJECT_BASE) {
					var echomsg nats.Msg
					echomsg.Subject = fmt.Sprintf("%s.bridge-client", natmsg.Reply)
					startpost := time.Now()
					tmpstring := startpost.Format("20060102-15:04:05.000")
					echoMsg := fmt.Sprintf("%s | %s", tmpstring, "message-client")
					echomsg.Data = []byte(echoMsg)
					mType := echomsg.Subject
					mSource := "urn:theotw:astra:bridge-client"
					cvMessage, err := msgFormat.GeneratePayload(echoMsg, mType, mSource)
					if err != nil {
						log.Errorf("Failed to generate cloud events payload: %s", err.Error())
						return
					}
					echomsg.Data = cvMessage
					sendMessageToCloud(t.serverURL, clientID, pkg.Config.CloudEvents, &echomsg)
					endpost := time.Now()
					metrics.RecordTimeToPushMessage(int(math.Round(endpost.Sub(startpost).Seconds())))
				}

				if err := nc.PublishRequest(natmsg.Subject, natmsg.Reply, natmsg.Data); err != nil {
					log.Errorf("Error publishing request: %s", err)
				}
			} else {
				if err := nc.Publish(natmsg.Subject, natmsg.Data); err != nil {
					log.Errorf("Error publishing request: %s", err)
				}
			}
		}
	}
}

func subscribeToOutboundMessages(serverURL, clientID string) (*nats.Subscription, error) {
	nc := natsmodel.GetNatsConnection()
	subj := fmt.Sprintf("%s.>", msgs.NATSSYNC_MESSAGE_PREFIX)
	sub, err := nc.SubscribeSync(subj)
	if err != nil {
		return nil, err
	}
	go handleOutboundMessages(sub, serverURL, clientID)
	return sub, nil
}

// handleOutboundMessages  This pulls messages off the queue and groups a bunch of them to push them together
// if we have to wait more than N ms for a message, we will go ahead and send what we have
// or if we get more than N messages, we will send them along
func handleOutboundMessages(subscription *nats.Subscription, serverURL, clientID string) {
	timeoutStr := pkg.GetEnvWithDefaults("NATSSYNC_MSG_WAIT_TIMEOUT", "5")
	maxMsgHoldStr := pkg.GetEnvWithDefaults("NATSSYNC__MAX_MSG_HOLD", "512")
	waitTimeout, numErr := strconv.ParseInt(timeoutStr, 10, 16)
	if numErr != nil {
		waitTimeout = 5
	}
	maxQueueSize, numErr := strconv.ParseInt(maxMsgHoldStr, 10, 16)
	if numErr != nil {
		waitTimeout = 512
	}

	msgList := make([]*nats.Msg, 0)
	keepGoing := true
	for keepGoing {
		msg, err := subscription.NextMsg(time.Duration(waitTimeout) * time.Millisecond)
		sendWhatWeHave := false
		//if we get a timeout (or any error) send what we have
		if err != nil {
			sendWhatWeHave = len(msgList) > 0
			// bail if it was not a timeout error
			keepGoing = err == nats.ErrTimeout
		} else {
			parsedSubject, err2 := msgs.ParseSubject(msg.Subject)
			if err2 == nil {
				log.Tracef("Found message to send NB Stored  Client ID=%s, Message Target %s", clientID, parsedSubject.LocationID)
				//if the target client ID is not this client, push it to the server
				if parsedSubject.LocationID != clientID {
					log.Tracef("Adding message to list to send NB")
					msgList = append(msgList, msg)
				} else {
					log.Tracef("Message not meant for NB, dropping")
				}
			}
			sendWhatWeHave = len(msgList) > int(maxQueueSize)
		}
		if sendWhatWeHave {
			sendMessageToCloud(serverURL, clientID, false, msgList...)
			msgList = make([]*nats.Msg, 0)
		}
	}
	log.Infof("Leaving Handle Outbound Messages ")
}
func sendMessageToCloud(serverURL string, clientID string, ceEnabled bool, msgsList ...*nats.Msg) {
	messagesToSend := make([]v1.BridgeMessage, 0)
	for _, msg := range msgsList {
		msgFormat := msgs.GetMsgFormat()
		status, err := msgFormat.ValidateMsgFormat(msg.Data, ceEnabled)
		if err != nil {
			log.Errorf("Error validating the cloud event message: %s", err.Error())
			continue
		}
		if !status {
			log.Errorf("Cloud event message validation failed, ignoring the message...")
			continue
		}

		natmsg := bridgemodel.NatsMessage{Reply: msg.Reply, Subject: msg.Subject, Data: msg.Data}
		envelope, enverr := msgs.PutObjectInEnvelope(natmsg, clientID, pkg.CLOUD_ID)
		if enverr != nil {
			log.Errorf("Error putting msg in envelope %s", enverr.Error())
			continue
		}
		jsonbits, jsonerr := json.Marshal(&envelope)
		if jsonerr != nil {
			log.Errorf("Error encoding envelope to json bits, wkipping message %s", jsonerr.Error())
			continue
		}
		bmsg := v1.BridgeMessage{ClientID: clientID, MessageData: string(jsonbits), FormatVersion: "1"}
		messagesToSend = append(messagesToSend, bmsg)

	}
	url := fmt.Sprintf("%s/bridge-server/1/message-queue/%s", serverURL, clientID)

	for true {
		fullPostReq := v1.BridgeMessagePostReq{
			AuthChallenge: *msgs.NewAuthChallengeFromStoredKey(),
			Messages:      messagesToSend,
		}

		httpclient := bridgemodel.NewHttpClient()
		postErr := httpclient.SendAuthorizedRequestWithBodyAndResp(http.MethodPost, url, fullPostReq, nil)
		//resp, postErr := http.DefaultClient.Post(url, "application/json", r)
		if postErr != nil {
			log.WithError(postErr).Errorf("Error sending message to server.  Dropping the messages ")
			if isInvalidCertificateError(postErr) {
				if certRotationErr := NewCertRotationHandler(serverURL, clientID).HandleCertRotation(); certRotationErr != nil {
					log.Errorf("failed to rotate certificates")
					return
				}

				// cert rotation successful retry the original request
				continue
			}

			return
		}
		break
	}

}
