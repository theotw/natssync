/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/metrics"
	"github.com/theotw/natssync/pkg/msgs"
	"os"
)

func RunBridgeServerApp(test bool) {
	//hack for when we run as unit tests
	wd,_:=os.Getwd()
	if wd=="/build/tests/apps"{
		os.Chdir("/build")
	}
	log.Info("Starting NATSSync Server")
	log.Infof("Build date: %s", pkg.GetBuildDate())
	level, levelerr := log.ParseLevel(pkg.Config.LogLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)
	metrics.InitMetrics()
	msgs.InitCloudKey()
	log.Info("Init Cache Mgr")
	err := InitCacheMgr()
	if err != nil {
		log.Fatalf("Unable to initialize the cache manager %s", err.Error())
	}
	subjectString := fmt.Sprintf("%s.>", msgs.SB_MSG_PREFIX)
	go RunMsgHandler(subjectString)
	log.Info("Starting Server")
	RunBridgeServer(test)
	log.Info("Server stopped")
}
