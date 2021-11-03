/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
	"github.com/theotw/natssync/pkg/httpsproxy/proxylet"
)

func main() {
	log.Infof("Version %s", pkg.VERSION)
	logLevel := httpproxy.GetEnvWithDefaults("LOG_LEVEL", "debug")
	level, levelerr := log.ParseLevel(logLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.SetLevel(level)

	proxyletObject, err := proxylet.NewProxylet()
	if err != nil {
		log.WithError(err).Fatal("Failed to create proxylet object")
	}
	proxyletObject.RunHttpProxylet()
	runtime.Goexit()
}
