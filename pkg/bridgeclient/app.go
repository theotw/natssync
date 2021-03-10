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
	"strings"
	"time"
)

type Arguments struct {
	natsURL        *string
	cloudServerURL *string
}

func getClientArguments() Arguments {
	args := Arguments{
		flag.String("u", pkg.Config.NatsServerUrl, "URL to connect to NATS"),
		flag.String("c", pkg.Config.CloudBridgeUrl, "URL to connect to Cloud Server"),
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
	if err := RunBridgeClientRestAPI(test); err != nil {
		log.Errorf("Error starting API server %s", err.Error())
		os.Exit(1)
	}

	metrics.InitMetrics()

	serverURL := *args.cloudServerURL

	var lastClientID string
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
			nc.Close()
			nc = nil
			lastClientID = clientID
		}
		subj := fmt.Sprintf("%s.%s.>", msgs.NB_MSG_PREFIX, msgs.CLOUD_ID)
		_, err = nc.Subscribe(subj, func(msg *nats.Msg) {
			sendMessageToCloud(msg, serverURL, clientID)
		})
		if err != nil {
			log.Fatalf("Error subscribing to %s: %s", subj, err)
		}
		log.Infof("Subscribed to %s", subj)

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

			log.Infof("Received message: %s", string(natmsg.Data))

			if len(natmsg.Reply) > 0 {
				if strings.HasSuffix(natmsg.Subject, msgs.ECHO_SUBJECT_BASE) {
					var echomsg nats.Msg
					echomsg.Subject = fmt.Sprintf("%s.bridge-client", natmsg.Reply)
					startpost := time.Now()
					tmpstring := startpost.Format("20060102-15:04:05.000")
					echoMsg := fmt.Sprintf("%s | %s", tmpstring, "message-client")
					echomsg.Data = []byte(echoMsg)
					sendMessageToCloud(&echomsg, serverURL, clientID)
					endpost := time.Now()
					metrics.RecordTimeToPushMessage(int(math.Round(endpost.Sub(startpost).Seconds())))
				}

				if err := nc.PublishRequest(natmsg.Subject, natmsg.Reply, natmsg.Data); err != nil {
					log.Errorf("Error publising request: %s", err)
				}
			} else {
				if err := nc.Publish(natmsg.Subject, natmsg.Data); err != nil {
					log.Errorf("Error publising request: %s", err)
				}
			}
		}
	}
}

func sendMessageToCloud(msg *nats.Msg, serverURL string, clientID string) {
	log.Debugf("Sending Msg NB %s", msg.Subject)
	url := fmt.Sprintf("%s/bridge-server/1/message-queue/%s", serverURL, clientID)
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
	var fullPostReq v1.BridgeMessagePostReq
	fullPostReq.AuthChallenge = *msgs.NewAuthChallenge()
	fullPostReq.Messages = make([]v1.BridgeMessage, 1)
	fullPostReq.Messages[0] = bmsg
	postMsgBits, bmsgerr := json.Marshal(&fullPostReq)
	if bmsgerr != nil {
		log.Errorf("Error marshaling bridge message to json bits %s", jsonerr.Error())
		return
	}

	r := bytes.NewReader(postMsgBits)
	resp, posterr := http.DefaultClient.Post(url, "application/json", r)
	if posterr != nil {
		log.Errorf("Error sending message to server.  Dropping the message %s  error was %s", msg.Subject, posterr.Error())
		return
	}
	if resp.StatusCode >= 300 {
		log.Errorf("Error sending message to server.  Dropping the message %s  error was %s", msg.Subject, resp.Status)
	}
}
