/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import (
	"flag"
	"fmt"
	"github.com/theotw/natssync/pkg/testing"
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

var quitChannel = make(chan os.Signal, 1)

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

	if pkg.Config.SkipTlsValidation {
		log.Warn("SKIP_TLS_VALIDATION was set to true! Don't use this in production!")
		bridgemodel.ConfigureDefaultTransportToSkipTlsValidation()
	}

	if err := RunBridgeClientRestAPI(); err != nil {
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
	if test {
		testing.NotifyOnAppExitMessage(connection, quitChannel)
	}

	var lastClientID string
	var currentMessageHandler BiDiMessageHandler

	// loop around watching for any changes to the client ID which happens if the user re-registers.
	// if we see that happens, tear down the message handler and start a new one
	for true {
		if timeToQuit(quitChannel) {
			log.Info("Quit signal received, exiting app...")
			return
		}

		nc := bridgemodel.GetNatsConnection()
		clientID := store.LoadLocationID("")
		//no client ID yet?  that happens on a new startup before it is registered.  Just hang out and wait for one
		if len(clientID) == 0 {
			log.Infof("No client ID, sleeping and retrying")
			time.Sleep(5 * time.Second)
			continue
		}
		//in case we re-register and the client ID changes, change what we listen for
		if (clientID != lastClientID) && nc != nil {
			if currentMessageHandler != nil {
				currentMessageHandler.StopMessageHandler()
				currentMessageHandler = nil
			}
			lastClientID = clientID
			//announce the cloud ID/location ID at startup and changes
			connection.Publish(bridgemodel.ResponseForLocationID, []byte(clientID))
			currentMessageHandler=NewBidiMessageHandler(serverURL)
			log.Infof("Starting Message Handler of type %s ",currentMessageHandler.GetHandlerType())
			currentMessageHandler.StartMessageHandler(clientID)
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



func timeToQuit(quitChannel chan os.Signal) bool {
	select {
	case <-quitChannel:
		return true
	default:
		return false
	}
}
