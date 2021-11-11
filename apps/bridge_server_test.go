/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/cloudserver"
	"testing"
)

func Test_bridge_server(t *testing.T) {
	log.Infof("Version %s",pkg.VERSION)
	cloudserver.RunBridgeServerApp(true)
}
