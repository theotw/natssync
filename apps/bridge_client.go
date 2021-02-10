/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	cloudclient "github.com/theotw/natssync/pkg/bridgeclient"
)

//The main class for the on site (southside) client.
//Env vars needed are:
//NATS_SERVER_URL=nats://127.0.0.1:4222
//CLOUD_BRIDGE_URL=http://somehost:port
//PREM_ID=the location ID from a registration request
func main() {
	cloudclient.RunClient(false)
}
