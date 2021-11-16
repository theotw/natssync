package nats

import (
	natspkg "github.com/nats-io/nats.go"
	"github.com/theotw/natssync/pkg/testing"
	"os"
)

type MsgHandler func(*Msg)

type ClientInterface interface {
	Subscribe(subj string, cb MsgHandler) (NatsSubscriptionInterface, error)
	Publish(subj string, data []byte) error
	LastError() error
	Flush() error
	SubscribeSync(reply string) (NatsSubscriptionInterface, error)
	PublishRequest(subj string, reply string, data []byte) error
	NotifyOnAppExitMessage(quitChannel chan os.Signal)
}

type natsClient struct {
	natsConn *natspkg.Conn
}

func (n natsClient) PublishRequest(subj string, reply string, data []byte) error {
	return n.natsConn.PublishRequest(subj, reply, data)
}

func (n natsClient) Subscribe(subj string, cb MsgHandler) (NatsSubscriptionInterface, error) {
	pkgCb := func (msg *natspkg.Msg) {
		cb((*Msg)(msg))
	}
	pkgSubscription, err := n.natsConn.Subscribe(subj, pkgCb)
	return newNatsSubscription(pkgSubscription), err
}

func (n natsClient) Publish(subj string, data []byte) error {
	return n.natsConn.Publish(subj, data)
}

func (n natsClient) LastError() error {
	return n.natsConn.LastError()
}

func (n natsClient) Flush() error {
	return n.natsConn.Flush()
}

func (n natsClient) SubscribeSync(subj string) (NatsSubscriptionInterface, error){
	pkgSubscription, err := n.natsConn.SubscribeSync(subj)
	return newNatsSubscription(pkgSubscription), err
}

func (n natsClient) NotifyOnAppExitMessage(quitChannel chan os.Signal) {
	testing.NotifyOnAppExitMessage(n.natsConn, quitChannel)
}

func newNatsClient(natsConn *natspkg.Conn) *natsClient {
	return &natsClient{natsConn: natsConn}
}

func Connect(natsURL string) (ClientInterface, error){
	natsConn, err := natspkg.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	return newNatsClient(natsConn), nil
}
