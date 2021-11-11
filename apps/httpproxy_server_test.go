/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func Test_httpproxy_server(t *testing.T) {
	log.Infof("Version %s",pkg.VERSION)
	if err := server.RunHttpProxyServer(true); err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}
