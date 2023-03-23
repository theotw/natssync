package nats

import (
	"time"

	natspkg "github.com/nats-io/nats.go"
)


type Msg natspkg.Msg

type NatsSubscriptionInterface interface {
	AutoUnsubscribe(max int) error
	NextMsg(duration time.Duration) (*Msg, error)
	Unsubscribe() error
}

type natsSubscription struct {
	subscription *natspkg.Subscription
}

func (n *natsSubscription) Unsubscribe() error {
	return n.subscription.Unsubscribe()
}

func (n *natsSubscription) AutoUnsubscribe(max int) error {
	return n.subscription.AutoUnsubscribe(max)
}

func (n *natsSubscription) NextMsg(duration time.Duration) (*Msg, error) {
	pkgMsg, err := n.subscription.NextMsg(duration)
	return (*Msg)(pkgMsg), err
}

func newNatsSubscription(subscription *natspkg.Subscription) NatsSubscriptionInterface {
	return &natsSubscription{
		subscription: subscription,
	}
}



