/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/cloudserver"
)

func main() {
	logLevel := pkg.GetEnvWithDefaults("LOG_LEVEL", "debug")

	level, levelerr := log.ParseLevel(logLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)

	fmt.Println("Init Cache Mgr")
	err := cloudserver.InitCacheMgr()
	if err != nil {
		log.Fatalf("Unable to initialize the cache manager %s", err.Error())
	}
	go cloudserver.RunMsgHandler("astra.>")
	fmt.Println("Starting Server")
	cloudserver.RunBridgeServer(false)
	fmt.Println("Server stopped")

}