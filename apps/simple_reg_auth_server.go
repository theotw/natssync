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
)

//the main for an example of a simple auth server.  Authorizes a requeest if the user ID and secret matches what is set in the env
//env vars to set are:
//USER_ID= the valid user ID defaults to natssync
//SECRET = the valid user secret.  defaults to changeit
func main() {
	natsURL := pkg.Config.NatsServerUrl
	log.Infof("Connecting to NATS server %s", natsURL)
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Errorf("Unable to connect to NATS, exiting %s", err.Error())
		os.Exit(2)
	}
	expectedAuthToken := pkg.GetEnvWithDefaults("AUTH_TOKEN", "42")
	subj := bridgemodel.REGISTRATION_AUTH_SUBJECT
	nc.Subscribe(subj, func(msg *nats.Msg) {
		log.Infof("Got message %s : ", subj, msg.Reply)
		var req bridgemodel.RegistrationRequest
		var resp bridgemodel.RegistrationResponse
		err := json.Unmarshal(msg.Data, &req)
		if err == nil {
			resp.Success = expectedAuthToken == req.AuthToken
		} else {
			resp.Success = false
		}
		respBits, _ := json.Marshal(&resp)
		log.Infof("Reg Request %s from %s success=%t", req.AuthToken, req.LocationID, resp.Success)

		nc.Publish(msg.Reply, respBits)
		nc.Flush()
	})
	runtime.Goexit()
}
