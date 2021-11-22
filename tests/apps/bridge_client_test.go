/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package apps

import (
	"github.com/prometheus/common/log"
	"github.com/theotw/natssync/pkg"
	cloudclient "github.com/theotw/natssync/pkg/bridgeclient"
	"testing"
)

func TestBridgeClient(t *testing.T) {
	log.Infof("Version %s", pkg.VERSION)
	cloudclient.RunClient(true)
}
