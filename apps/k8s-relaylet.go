/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg/k8srelay/relaylet"
	"os"
	"os/signal"
)

func main() {
	_, err := relaylet.NewRelaylet()
	if err != nil {
		log.WithError(err).Fatalf("Unable to initialize relaylet %s", err.Error())
	}
	log.Info("Server Started blocking on channel")
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Info("Shutdown Server ...")

	log.Info("Server exiting")
}
