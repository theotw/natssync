/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import "github.com/theotw/natssync/pkg/cloudserver"

//The main app for the server that runs in the cloud
func main() {
	cloudserver.RunBridgeServerApp(false)
}
