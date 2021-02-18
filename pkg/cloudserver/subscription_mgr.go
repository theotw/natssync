/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/msgs"
	"sync"
)

var mapSync sync.RWMutex
var nats_subscriptions map[string]*nats.Subscription

func InitSubscriptionMgr() error {
	mapSync.Lock()
	defer mapSync.Unlock()
	nats_subscriptions = make(map[string]*nats.Subscription)

	nc := bridgemodel.GetNatsConnection()
	if nc == nil {
		return errors.New("uninitialized nats connection")
	}
	_, suberr1 := nc.Subscribe(bridgemodel.REGISTRATION_LIFECYCLE_ADDED, handleNewSubscription)
	if suberr1 != nil {
		log.Errorf("Error registering for lifecycle add events %s \n", suberr1)
		return suberr1
	}
	_, suberr2 := nc.Subscribe(bridgemodel.REGISTRATION_LIFECYCLE_ADDED, handleRemovedSubscription)
	if suberr2 != nil {
		log.Errorf("Error registering for lifecycle add events %s \n", suberr2)
		return suberr2
	}

	knownClients, err := msgs.GetKeyStore().ListKnownClients()
	if err != nil {
		log.Errorf("Unable to list known client, is keystore initialized? %s \n", err)
		return err
	}
	for _, clientID := range knownClients {
		subject := fmt.Sprintf("%s.%s.>", msgs.SB_MSG_PREFIX, clientID)
		sub, err := nc.SubscribeSync(subject)
		if err != nil {
			log.Errorf("Unable to subscribe to %s because of %s \n", subject, err.Error())
		} else {
			nats_subscriptions[clientID] = sub
		}
	}
	return nil
}
func handleNewSubscription(msg *nats.Msg) {
	if msg.Data == nil || len(msg.Data) == 0 {
		log.Debugf("Got a new subscription message with no data \n")
		return
	}
	nc := bridgemodel.GetNatsConnection()

	clientID := string(msg.Data)
	subject := fmt.Sprintf("%s.%s.>", msgs.SB_MSG_PREFIX, clientID)
	sub, err := nc.SubscribeSync(subject)
	if err != nil {
		log.Errorf("Error subscribing to subject: %s error: %s \n", subject, err.Error())
		return
	}
	mapSync.Lock()
	nats_subscriptions[clientID] = sub
	mapSync.Unlock()
}
func handleRemovedSubscription(msg *nats.Msg) {
	if msg.Data == nil || len(msg.Data) == 0 {
		log.Debugf("Got a new subscription message with no data \n")
		return
	}
	clientID := string(msg.Data)

	mapSync.Lock()
	delete(nats_subscriptions, clientID)
	mapSync.Unlock()

}

//gets the subscription for the client ID or returns nil
func GetSubscriptionForClient(clientID string) *nats.Subscription {
	var ret *nats.Subscription
	mapSync.RLock()
	ret = nats_subscriptions[clientID]
	mapSync.RUnlock()
	return ret
}
