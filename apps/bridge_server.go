/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/cloudserver"
	"github.com/theotw/natssync/pkg/msgs"
)

//The main app for the server that runs in the cloud
func main() {
	logLevel := pkg.GetEnvWithDefaults("LOG_LEVEL", "debug")

	level, levelerr := log.ParseLevel(logLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)

	msgs.InitCloudKey()
	fmt.Println("Init Cache Mgr")
	err := cloudserver.InitCacheMgr()
	if err != nil {
		log.Fatalf("Unable to initialize the cache manager %s", err.Error())
	}
	subjectString := fmt.Sprintf("%s.>", msgs.SB_MSG_PREFIX)
	go cloudserver.RunMsgHandler(subjectString)
	fmt.Println("Starting Server")
	cloudserver.RunBridgeServer(false)
	fmt.Println("Server stopped")

}
