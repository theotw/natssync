/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package natsmodel

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var natsConnection *nats.Conn

// InitNats takes a comma separated list of NATS urls form of host:port,host:port
func InitNats(natsUrlList string, connectionName string, timeout time.Duration) error {
	userName := os.Getenv("NATS_USER")
	seed := os.Getenv("NATS_SEED")

	start := time.Now()
	done := false
	var errToReturn error
	var i time.Duration
	for !done {
		i = i + 1
		log.Infof("Connecting to NATS on %s", natsUrlList)

		opts := nats.Options{
			Url:  natsUrlList,
			Nkey: userName,
		}
		var sigHandler nats.SignatureHandler
		if len(seed) > 0 {
			sigHandler = func(nonce []byte) ([]byte, error) {
				seedBytes := []byte(seed)
				kp, err := nkeys.FromSeed(seedBytes)
				if err != nil {
					return nil, err
				}
				signature, err := kp.Sign(nonce)
				if err != nil {
					return nil, err
				}
				return signature, nil
			}
		}
		opts.SignatureCB = sigHandler
		opts.DisconnectedErrCB = func(_ *nats.Conn, err error) {
			if err != nil {
				log.Debugf("Connection disconnect %s", err.Error())
			} else {
				log.Debugf("Connection disconnect no error")
			}
		}
		opts.ClosedCB = func(_ *nats.Conn) {
			log.Debugf("NATS Connection closed")
		}
		opts.ReconnectedCB = func(_ *nats.Conn) {
			log.Debugf("Connection Reconnect")
		}
		opts.Name = connectionName
		nc, err := opts.Connect()

		if err != nil {
			log.Errorf("Error connecting to nats on URL %s  / Error %s", natsUrlList, err.Error())
			//increasing sleep longer
			time.Sleep(5 * time.Second)
			now := time.Now()
			done = now.Sub(start) >= timeout
			errToReturn = err
		} else {
			log.Infof("Connected to NATS on %s", natsUrlList)
			natsConnection = nc
			done = true
		}
		errToReturn = nil
	}
	log.Infof("Leaving NATS Init ")
	return errToReturn
}

func GetNatsConnection() *nats.Conn {
	return natsConnection
}
