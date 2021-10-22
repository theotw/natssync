/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package models

import (
	"errors"
	"github.com/nats-io/nats.go"
	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
	"time"

	log "github.com/sirupsen/logrus"
)

var nc *nats.Conn

func InitNats() error {
	counter := 0
	for {
		counter = counter + 1
		err := connect()
		if err != nil {
			log.Errorf("Unable to init nats, sleeping %s", err.Error())
			time.Sleep(3 * time.Second)
			if counter > 10 {
				return errors.New("timeout waiting for nats")
			}

		} else {
			break
		}
	}
	return nil
}
func connect() error {
	natsURL := httpproxy.GetEnvWithDefaults("NATS_SERVER_URL", "nats://127.0.0.1:4222")
	log.Infof("Connecting to NATS server %s", natsURL)

	tmpnc, err := nats.Connect(natsURL)
	if err != nil {
		log.Errorf("Error connecting to NATS %s", err.Error())

		return err
	}
	nc = tmpnc
	return nil
}
func GetNatsClient() *nats.Conn {
	return nc
}
