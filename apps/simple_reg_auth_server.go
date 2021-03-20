/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"os"
	"runtime"
	"time"
)

//the main for an example of a simple auth server.  Authorizes a request if the user ID and secret matches what is set in the env
//env vars to set are:
//USER_ID= the valid user ID defaults to natssync
//SECRET = the valid user secret. defaults to changeit
func main() {
	natsURL := pkg.Config.NatsServerUrl
	log.Infof("Connecting to NATS server %s", natsURL)

	err := bridgemodel.InitNats(natsURL, "echo client", 1*time.Minute)
	if err != nil {
		log.Errorf("Unable to connect to NATS, exiting %s", err.Error())
		os.Exit(2)

	}
	nc := bridgemodel.GetNatsConnection()

	expectedAuthToken := pkg.GetEnvWithDefaults("AUTH_TOKEN", "42")
	subj := bridgemodel.REGISTRATION_AUTH_WILDCARD
	nc.Subscribe(subj, func(msg *nats.Msg) {
		log.Infof("Got message %s : %s", msg.Subject, msg.Reply)
		var respBits []byte
		if msg.Subject == bridgemodel.REGISTRATION_AUTH_SUBJECT {
			var regReq bridgemodel.RegistrationRequest
			var regResp bridgemodel.RegistrationResponse
			err := json.Unmarshal(msg.Data, &regReq)
			if err == nil {
				regResp.Success = expectedAuthToken == regReq.AuthToken
			} else {
				regResp.Success = false
			}
			respBits, _ = json.Marshal(&regResp)
			log.Infof("Reg Request %s from %s success=%t", regReq.AuthToken, regReq.LocationID, regResp.Success)
		} else if msg.Subject == bridgemodel.UNREGISTRATION_AUTH_SUBJECT {
			var unregReq bridgemodel.UnRegistrationRequest
			var unregResp bridgemodel.UnRegistrationResponse
			err := json.Unmarshal(msg.Data, &unregReq)
			if err == nil {
				unregResp.Success = expectedAuthToken == unregReq.AuthToken
			} else {
				unregResp.Success = false
			}
			respBits, _ = json.Marshal(&unregResp)
			log.Infof("UnReg Request %s from %s success=%t", unregReq.AuthToken, unregReq.LocationID, unregResp.Success)
		} else {
			var regReq bridgemodel.GenericAuthRequest
			var regResp bridgemodel.GenericAuthResponse
			err := json.Unmarshal(msg.Data, &regReq)
			if err == nil {
				regResp.Success = expectedAuthToken == regReq.AuthToken
			} else {
				regResp.Success = false
			}
			respBits, _ = json.Marshal(&regResp)
			log.Infof("Generic AuthRequest %s success=%t", regReq.AuthToken, regResp.Success)

		}
		nc.Publish(msg.Reply, respBits)
		nc.Flush()
	})
	runtime.Goexit()
}
