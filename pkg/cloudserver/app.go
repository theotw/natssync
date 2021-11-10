/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
)

func RunBridgeServerApp(test bool) {
	if test {
		log.Warn("TEST MODE IS ENABLED")
	}

	level, levelerr := log.ParseLevel(pkg.Config.LogLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)

	log.Infof("Build date: %s", pkg.GetBuildDate())
	//hack for when we run as unit tests
	wd, _ := os.Getwd()
	if wd == "/build/tests/apps" {
		if err := os.Chdir("/build"); err != nil {
			log.Fatalf("Error attempting to change directories: %s", err)
		}
	}

	log.Info("Starting NATSSync Server")

	//if we cannot get to NATS, then we are worthless and should stop
	natsErr := bridgemodel.InitNats(pkg.Config.NatsServerUrl, "NatsSyncServer Master", 1*time.Minute)
	if natsErr != nil {
		log.Fatalf("Unable to connect to NATS. Ending app %s", natsErr.Error())
	}
	if keyError := msgs.InitCloudKey(); keyError != nil {
		log.Fatalf("Unable to initialize the key manager. Ending the app %s", keyError.Error())
	}
	if subError := InitSubscriptionMgr(); subError != nil {
		log.Fatalf("Unable to initialize the subscription manager. Ending the app %s", subError.Error())
	}

	metrics.InitMetrics()
	log.Info("Starting Server")
	RunBridgeServer(test)
	log.Info("Server stopped")
}
