/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Arguments struct {
	natsURL        *string
	cloudServerURL *string
	cloudEvents    *bool
}

func getClientArguments() Arguments {
	args := Arguments{
		flag.String("u", pkg.Config.NatsServerUrl, "URL to connect to NATS"),
		flag.String("c", pkg.Config.CloudBridgeUrl, "URL to connect to Cloud Server"),
		flag.Bool("ce", pkg.Config.CloudEvents, "Enable CloudEvents messaging format"),
	}
	flag.Parse()
	return args
}
func RunClient(test bool) {
	log.Info("Starting NATSSync Client")
	args := getClientArguments()
	err := bridgemodel.InitNats(*args.natsURL, "echo client", 1*time.Minute)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Build date: %s", pkg.GetBuildDate())
	level, levelerr := log.ParseLevel(pkg.Config.LogLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)
	if err := msgs.InitLocationKeyStore(); err != nil {
		log.Fatalf("Error initalizing key store: %s", err)
	}
	store := msgs.GetKeyStore()
	if store == nil {
		log.Fatalf("Unable to get keystore")
	}
	msgs.InitMessageFormat()
	msgFormat := msgs.GetMsgFormat()
	if msgFormat == nil {
		log.Fatalf("Unable to get the message format")
	}
	if err := RunBridgeClientRestAPI(test); err != nil {
		log.Errorf("Error starting API server %s", err.Error())
		os.Exit(1)
	}

	metrics.InitMetrics()

	serverURL := *args.cloudServerURL

	connection := bridgemodel.GetNatsConnection()
	connection.Subscribe(bridgemodel.RequestForLocationID, func(msg *nats.Msg) {
		clientID := store.LoadLocationID()
		connection.Publish(bridgemodel.ResponseForLocationID, []byte(clientID))
	})

	var lastClientID string
	var currentSubscription *nats.Subscription
	for true {
		nc := bridgemodel.GetNatsConnection()
		clientID := store.LoadLocationID()
		if len(clientID) == 0 {
			log.Infof("No client ID, sleeping and retrying")
			time.Sleep(5 * time.Second)
			continue
		}
		//in case we re-register and the client ID changes, change what we listen for
		if (clientID != lastClientID) && nc != nil {
			if currentSubscription != nil {
				currentSubscription.Unsubscribe()
				currentSubscription = nil
			}
			lastClientID = clientID
			//announce the cloud ID/location ID at startup
			connection.Publish(bridgemodel.ResponseForLocationID, []byte(clientID))
		}
		//same as above, if we re-register, we drop the subscibe and need to resubscribe
		if currentSubscription == nil {
			currentSubscription, err = subscribeToOutboundMessages(serverURL, clientID)
			if err != nil {
				log.Errorf("Error subscribing to messages, will try again %s", err.Error())
			}
		}

		url := fmt.Sprintf("%s/bridge-server/1/message-queue/%s", serverURL, clientID)

		ac := msgs.NewAuthChallenge()
		httpclient := bridgemodel.NewHttpClient()
		var msglist []v1.BridgeMessage
		err := httpclient.SendAuthorizedRequestWithBodyAndResp("GET", url, ac, &msglist)
		if err != nil {
			log.Errorf("Error fetching messages %s", err.Error())
			time.Sleep(2 * time.Second)
			continue
		}
		log.Infof("Received %d messages from server", len(msglist))
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
			status, err := msgFormat.ValidateMsgFormat(natmsg.Data, pkg.Config.CloudEvents)
			if err != nil {
				log.Errorf("Error validating the cloud event message: %s", err.Error())
				return
			}
			if !status {
				log.Errorf("Cloud event message validation failed, ignoring the message...")
				return
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
					sendMessageToCloud(serverURL, clientID, pkg.Config.CloudEvents, &echomsg)
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
	nc := bridgemodel.GetNatsConnection()
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
			keepGoing = err == nats.ErrTimeout
			sendWhatWeHave = len(msgList) > 0
		} else {
			parsedSubject, err2 := msgs.ParseSubject(msg.Subject)
			if err2 == nil {
				log.Tracef("Stored  Client ID %s", clientID)
				log.Tracef("Message Location ID %s", parsedSubject.LocationID)
				//if the target client ID is not this client, push it to the server
				if parsedSubject.LocationID != clientID {
					msgList = append(msgList, msg)
				}
			}
			sendWhatWeHave = len(msgList) > int(maxQueueSize)
		}
		if sendWhatWeHave {
			sendMessageToCloud(serverURL, clientID, false, msgList...)
			msgList = make([]*nats.Msg, 0)
		}
	}
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
		envelope, enverr := msgs.PutObjectInEnvelope(natmsg, clientID, msgs.CLOUD_ID)
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
		messagesToSend=append(messagesToSend,bmsg)

	}
	url := fmt.Sprintf("%s/bridge-server/1/message-queue/%s", serverURL, clientID)

	var fullPostReq v1.BridgeMessagePostReq
	fullPostReq.AuthChallenge = *msgs.NewAuthChallenge()

	fullPostReq.Messages = messagesToSend
	fullPostReq.Messages=messagesToSend
	postMsgBits, bmsgerr := json.Marshal(&fullPostReq)
	if bmsgerr != nil {
		log.Errorf("Error marshaling bridge message to json bits %s", bmsgerr.Error())
		return
	}

	r := bytes.NewReader(postMsgBits)
	resp, posterr := http.DefaultClient.Post(url, "application/json", r)
	if posterr != nil {
		log.Errorf("Error sending message to server.  Dropping the message %d  error was %s", len(messagesToSend), posterr.Error())
		return
	}
	if resp.StatusCode >= 300 {
		log.Errorf("Error sending messages to server.  Dropping the message %d  error was %s", len(messagesToSend), resp.Status)
	}
}
