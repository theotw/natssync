/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
)

const (
	logLevelEnvVariableServer = "LOG_LEVEL"
)

func main() {

	logLevel := httpproxy.GetEnvWithDefaults(logLevelEnvVariableServer, log.DebugLevel.String())

	level, levelErr := log.ParseLevel(logLevel)
	if levelErr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)

	log.Infof("Version %s", pkg.VERSION)

	proxyServer, err := server.NewServer()
	if err != nil {
		log.WithError(err).Fatal("failed to instantiate proxy server")
	}
	proxyServer.RunHttpProxyServer()
}
