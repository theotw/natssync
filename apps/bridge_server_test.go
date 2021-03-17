/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package main

import (
	"testing"

	"github.com/theotw/natssync/pkg/cloudserver"
)

func Test_main(t *testing.T) {
	cloudserver.RunBridgeServerApp(false)
}
