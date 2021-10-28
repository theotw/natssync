/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
	"github.com/theotw/natssync/pkg/persistence"
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
	if err := persistence.InitLocationKeyStore(); err != nil {
		log.Fatalf("Error initalizing key store: %s", err)
	}
	store := persistence.GetKeyStore()
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
		clientID := store.LoadLocationID("")
		connection.Publish(bridgemodel.ResponseForLocationID, []byte(clientID))
	})

	var lastClientID string
	var currentSubscription *nats.Subscription
	for true {
		nc := bridgemodel.GetNatsConnection()
		clientID := store.LoadLocationID("")
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
			subj := fmt.Sprintf("%s.>", msgs.NATSSYNC_MESSAGE_PREFIX)
			sub, err := nc.Subscribe(subj, func(msg *nats.Msg) {
				parsedSubject, err2 := msgs.ParseSubject(msg.Subject)
				if err2 == nil {
					log.Tracef("Stored  Client ID %s", clientID)
					log.Tracef("Message Location ID %s", parsedSubject.LocationID)
					//if the target client ID is not this client, push it to the server
					if parsedSubject.LocationID != clientID {
						sendMessageToCloud(msg, serverURL, clientID, pkg.Config.CloudEvents)
					}
				}

			})
			if err != nil {
				log.Fatalf("Error subscribing to %s: %s", subj, err)
				continue
			} else {
				currentSubscription = sub
				log.Infof("Subscribed to %s", subj)
			}
		}

		msglist, err := getMessagesFromCloud(serverURL, clientID)
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
					sendMessageToCloud(&echomsg, serverURL, clientID, pkg.Config.CloudEvents)
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

func isInvalidCertificateError(err error) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("status code %v", pkg.StatusCertificateError))
}

func getMessagesFromCloud(serverURL, clientID string) ([]v1.BridgeMessage, error) {
	url := fmt.Sprintf("%s/bridge-server/1/message-queue/%s", serverURL, clientID)

	httpclient := bridgemodel.NewHttpClient()
	var msglist []v1.BridgeMessage

	for true {
		ac := msgs.NewAuthChallenge("")
		err := httpclient.SendAuthorizedRequestWithBodyAndResp(http.MethodGet, url, ac, &msglist)
		if err != nil {
			if isInvalidCertificateError(err) {
				if certRotationErr := NewCertRotationHandler(serverURL, clientID).HandleCertRotation(); certRotationErr != nil {
					return nil, fmt.Errorf("failed to rotate certificates: %v : %v", certRotationErr, err)
				}
				// certificates rotated successfully, retry the original request
				continue
			}
			return nil, err
		}
		break
	}

	return msglist, nil
}

func sendMessageToCloud(msg *nats.Msg, serverURL string, clientID string, ceEnabled bool) {
	log.Debugf("Sending Msg NB %s", msg.Subject)

	msgFormat := msgs.GetMsgFormat()
	status, err := msgFormat.ValidateMsgFormat(msg.Data, ceEnabled)
	if err != nil {
		log.Errorf("Error validating the cloud event message: %s", err.Error())
		return
	}
	if !status {
		log.Errorf("Cloud event message validation failed, ignoring the message...")
		return
	}

	url := fmt.Sprintf("%s/bridge-server/1/message-queue/%s", serverURL, clientID)
	natmsg := bridgemodel.NatsMessage{Reply: msg.Reply, Subject: msg.Subject, Data: msg.Data}
	envelope, enverr := msgs.PutObjectInEnvelope(natmsg, clientID, pkg.CLOUD_ID)
	if enverr != nil {
		log.Errorf("Error putting msg in envelope %s", enverr.Error())
		return
	}

	jsonbits, jsonerr := json.Marshal(envelope)
	if jsonerr != nil {
		log.WithError(err).Errorf("Error encoding envelope to json bits")
		return
	}
	bmsg := v1.BridgeMessage{ClientID: clientID, MessageData: string(jsonbits), FormatVersion: "1"}

	for true {
		fullPostReq := v1.BridgeMessagePostReq{
			AuthChallenge: *msgs.NewAuthChallenge(""),
			Messages:      []v1.BridgeMessage{bmsg},
		}

		postMsgBits, bmsgerr := json.Marshal(&fullPostReq)
		if bmsgerr != nil {
			log.WithError(bmsgerr).Errorf("Error marshaling bridge message to json bits")
			return
		}

		r := bytes.NewReader(postMsgBits)
		resp, postErr := http.DefaultClient.Post(url, "application/json", r)
		if postErr != nil {
			log.WithError(postErr).Errorf("Error sending message to server.  Dropping the message %s", msg.Subject)
			return
		}

		if resp.StatusCode == pkg.StatusCertificateError {
			if certRotationErr := NewCertRotationHandler(serverURL, clientID).HandleCertRotation(); certRotationErr != nil {
				log.Errorf("failed to rotate certificates")
				return
			}

			// cert rotation successful retry the original request
			continue
		}

		if resp.StatusCode >= 300 {
			log.Errorf("Error sending message to server.  Dropping the message %s  status %s", msg.Subject, resp.Status)
			return
		}
	}

}
