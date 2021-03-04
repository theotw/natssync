/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"errors"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg/bridgemodel"
	"github.com/theotw/natssync/pkg/msgs"
)

var mapSync sync.RWMutex
var natsSubscriptions map[string]*nats.Subscription

func InitSubscriptionMgr() error {
	mapSync.Lock()
	defer mapSync.Unlock()
	natsSubscriptions = make(map[string]*nats.Subscription)
	var err error

	nc := bridgemodel.GetNatsConnection()
	if nc == nil {
		return errors.New("uninitialized nats connection")
	}
	_, err = nc.Subscribe(bridgemodel.REGISTRATION_LIFECYCLE_ADDED, handleNewSubscription)
	if err != nil {
		log.Errorf("Error registering for lifecycle add events %s", err)
		return err
	}
	_, err = nc.Subscribe(bridgemodel.REGISTRATION_LIFECYCLE_REMOVED, handleRemovedSubscription)
	if err != nil {
		log.Errorf("Error registering for lifecycle removed events %s", err)
		return err
	}
	_, err = nc.Subscribe(bridgemodel.ACCOUNT_LIFECYCLE_REMOVED, handleRemoveAccount)
	if err != nil {
		log.Errorf("Error registering for account removed events %s", err)
		return err
	}

	knownClients, err := msgs.GetKeyStore().ListKnownClients()
	if err != nil {
		log.Errorf("Unable to list known client, is keystore initialized? %s \n", err)
		return err
	}
	for _, clientID := range knownClients {
		subject := fmt.Sprintf("%s.%s.>", msgs.SB_MSG_PREFIX, clientID)
		//sub, err := nc.SubscribeSync(subject)
		sub, err := nc.QueueSubscribeSync(subject, "natssync-get")
		if err != nil {
			log.Errorf("Unable to subscribe to %s because of %s \n", subject, err.Error())
		} else {
			natsSubscriptions[clientID] = sub
		}
	}
	return nil
}

func handleNewSubscription(msg *nats.Msg) {
	if msg.Data == nil || len(msg.Data) == 0 {
		log.Debugf("Got a new subscription message with no data")
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
	natsSubscriptions[clientID] = sub
	mapSync.Unlock()
}

func handleRemovedSubscription(msg *nats.Msg) {
	if msg.Data == nil || len(msg.Data) == 0 {
		log.Debugf("Got a remove subscription message with no data")
		return
	}
	clientID := string(msg.Data)

	mapSync.Lock()
	defer mapSync.Unlock()
	log.Infof("Removing subscription for clientID %s", clientID)

	sub := natsSubscriptions[clientID]
	if sub == nil {
		log.Warnf("No subscription found for %s", clientID)
		return
	}
	if err := sub.Unsubscribe(); err != nil {
		log.Errorf("Error unsubscribing from %s: %s", clientID, err)
	}
	delete(natsSubscriptions, clientID)
}

func handleRemoveAccount(msg *nats.Msg) {
	if msg.Data == nil || len(msg.Data) == 0 {
		log.Debugf("Got a remove account message with no data")
		return
	}
	clientID := string(msg.Data)
	log.Infof("Removing account for clientID %s", clientID)

	keystore := msgs.GetKeyStore()
	if err := keystore.RemoveLocation(clientID); err != nil {
		log.Error(err)
		return
	}
	nc := bridgemodel.GetNatsConnection()
	log.Tracef("Publishing subscription remove msg for clientID %s", clientID)
	if err := nc.Publish(bridgemodel.REGISTRATION_LIFECYCLE_REMOVED, msg.Data); err != nil {
		log.Error(err)
	}
}

//gets the subscription for the client ID or returns nil
func GetSubscriptionForClient(clientID string) *nats.Subscription {
	var ret *nats.Subscription
	mapSync.RLock()
	ret = natsSubscriptions[clientID]
	mapSync.RUnlock()
	return ret
}
