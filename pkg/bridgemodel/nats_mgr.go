/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package bridgemodel

import (
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"time"
)

var natsConnection *nats.Conn

//takes a comma seperated list of NATS urls form of host:port,host:port
func InitNats(natsUrlList string, connectionName string, timeout time.Duration) error {
	start := time.Now()
	done := false
	var errToReturn error
	var i time.Duration
	for !done {
		i = i + 1
		log.Infof("Connecting to NATS on %s\n", natsUrlList)
		nc, err := nats.Connect(natsUrlList, nats.ClosedHandler(func(_ *nats.Conn) {
				log.Debugf("NATS Connection closed  \n")
			}),
			nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
				if err != nil {
					log.Debugf("Connection disconnect %s  \n", err.Error())
				} else {
					log.Debugf("Connection disconnect no error  \n")
				}
			}),
			nats.ReconnectHandler(func(_ *nats.Conn) {
				log.Debugf("Connection Reconnect  \n")
			}),
			nats.Name(connectionName),
		)

		if err != nil {
			log.Errorf("Error connecting to nats on URL %s  / Error %s \n", natsUrlList, err.Error())
			//increasing sleep longer
			time.Sleep(i * time.Second)
			now := time.Now()
			done = now.Sub(start) >= timeout
			errToReturn = err
		} else {
			log.Infof("Connected to NATS on %s \n", natsUrlList)
			natsConnection = nc
			done = true
		}
		errToReturn = nil
	}
	return errToReturn
}

func GetNatsConnection() *nats.Conn {
	return natsConnection
}
