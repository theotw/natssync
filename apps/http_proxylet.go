/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/httpsproxy/proxylet"
	utils2 "github.com/theotw/natssync/utils"
)

func main() {
	utils2.InitLogging()

	log.Infof("Version %s", pkg.VERSION)

	proxyletObject, err := proxylet.NewProxylet()
	if err != nil {
		log.WithError(err).Fatal("Failed to create proxylet object")
	}

	proxyletObject.RunHttpProxylet(false)
}
