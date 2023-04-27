/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg/httpsproxy/proxylet"
)

func main() {

	proxyletObject, err := proxylet.NewProxylet()
	if err != nil {
		log.WithError(err).Fatal("Failed to create proxylet object")
	}

	go proxylet.RunMetricsServer()

	proxyletObject.RunHttpProxylet(false)
}
