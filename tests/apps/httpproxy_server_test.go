/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package apps

import (
	"testing"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
	utils2 "github.com/theotw/natssync/utils"

	log "github.com/sirupsen/logrus"
)

func TestHttpProxyServer(t *testing.T) {
	utils2.InitLogging()

	log.Infof("Version %s", pkg.VERSION)

	proxyServer, err := server.NewServer()
	if err != nil {
		log.WithError(err).Fatal("failed to instantiate proxy server")
	}

	proxyServer.RunHttpProxyServer(true)
}
