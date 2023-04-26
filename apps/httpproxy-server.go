/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
	"github.com/theotw/natssync/utils"
)

func main() {
	utils.InitLogging()

	log.Infof("Version %s", pkg.VERSION)
	log.Debugf("Version debug %s", pkg.VERSION)
	log.Tracef("Version trace %s", pkg.VERSION)

	proxyServer, err := server.NewServer()
	if err != nil {
		log.WithError(err).Fatal("failed to instantiate proxy server")
	}

	proxyServer.RunHttpProxyServer(false)
}
