/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
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

func main(){
	natsURL := pkg.GetEnvWithDefaults("NATS_SERVER_URL", "nats://127.0.0.1:4222")

	log.Infof("Connecting to NATS server %s", natsURL)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Errorf("Unable to connect to NATS, exiting %s", err.Error())
		os.Exit(2)
	}
	userID:=pkg.GetEnvWithDefaults("USER_ID","natssync")
	password:=pkg.GetEnvWithDefaults("USER_ID","changeit")
	subj:=bridgemodel.REGISTRATION_AUTH_SUBJECT
	nc.Subscribe(subj, func(msg *nats.Msg) {
		log.Infof("Got message %s : ", subj, msg.Reply)
		var req bridgemodel.RegistrationRequest
		var resp bridgemodel.RegistrationResponse
		err:=json.Unmarshal(msg.Data,&req)
		if err == nil{
			resp.Success=userID == req.UserID && password==req.Secret
		}else{
			resp.Success=false
		}
		respBits,_:=json.Marshal(&resp)
		log.Infof("Reg Request %s from %s success=%t",req.UserID, req.LocationID, resp.Success)

		nc.Publish(msg.Reply, respBits)
		nc.Flush()
	})
	runtime.Goexit()
}
