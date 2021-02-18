/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
	"os"
	"time"
)

func RunBridgeServerApp(test bool) {
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
		os.Chdir("/build")
	}

	log.Info("Starting NATSSync Server")

	//if we cannot get to NATS, then we are worthless and should stop
	nats_err := bridgemodel.InitNats(pkg.Config.NatsServerUrl, "NatsSyncServer Master", 1*time.Minute)
	if nats_err != nil {
		log.Errorf("Unable to connect to NATS.  Ending app %s \n", nats_err.Error())
		return
	}
	key_error := msgs.InitCloudKey()
	if key_error != nil {
		log.Errorf("Unable to initialize the key manager.  Ending the app %s \n", key_error.Error())
		return
	}
	sub_error := InitSubscriptionMgr()
	if sub_error != nil {
		log.Errorf("Unable to initialize the subscription manager.  Ending the app %s \n", sub_error.Error())
		return
	}

	metrics.InitMetrics()
	log.Info("Starting Server")
	RunBridgeServer(test)
	log.Info("Server stopped")
}
