package nats

import (
	natspkg "github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type MsgHandler func(*Msg)

type ClientInterface interface {
	Subscribe(subj string, cb MsgHandler) (NatsSubscriptionInterface, error)
	QueueSubscribe(subj, queue string, cb MsgHandler) (NatsSubscriptionInterface, error)
	Publish(subj string, data []byte) error
	LastError() error
	Flush() error
	SubscribeSync(reply string) (NatsSubscriptionInterface, error)

	PublishRequest(subj string, reply string, data []byte) error
}

type natsClient struct {
	natsConn *natspkg.Conn
}

func (n natsClient) GetNatsClient() *natspkg.Conn {
	return n.natsConn
}
func (n natsClient) PublishRequest(subj string, reply string, data []byte) error {
	return n.natsConn.PublishRequest(subj, reply, data)
}

func (n natsClient) Subscribe(subj string, cb MsgHandler) (NatsSubscriptionInterface, error) {
	pkgCb := func(msg *natspkg.Msg) {
		cb((*Msg)(msg))
	}
	pkgSubscription, err := n.natsConn.Subscribe(subj, pkgCb)
	return newNatsSubscription(pkgSubscription), err
}
func (n natsClient) QueueSubscribe(subj, queue string, cb MsgHandler) (NatsSubscriptionInterface, error) {
	pkgCb := func(msg *natspkg.Msg) {
		cb((*Msg)(msg))
	}
	pkgSubscription, err := n.natsConn.QueueSubscribe(subj, queue, pkgCb)
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

func (n natsClient) SubscribeSync(subj string) (NatsSubscriptionInterface, error) {
	pkgSubscription, err := n.natsConn.SubscribeSync(subj)
	return newNatsSubscription(pkgSubscription), err
}

func newNatsClient(natsConn *natspkg.Conn) *natsClient {
	return &natsClient{natsConn: natsConn}
}

func Connect(natsURL string) (ClientInterface, error) {

	//natsConn, err := initNats(natsURL, 2*time.Minute)
	natsConn, err := natspkg.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	return newNatsClient(natsConn), nil
}
func initNats(natsUrlList string, timeout time.Duration) (*natspkg.Conn, error) {
	var ret *natspkg.Conn
	userName := os.Getenv("NATS_USER")
	seed := os.Getenv("NATS_SEED")

	start := time.Now()
	done := false
	var errToReturn error
	var i time.Duration
	for !done {
		i = i + 1
		log.Infof("Connecting to NATS on %s", natsUrlList)

		opts := natspkg.Options{
			Url:  natsUrlList,
			Nkey: userName,
		}
		var sigHandler natspkg.SignatureHandler
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
		opts.DisconnectedErrCB = func(_ *natspkg.Conn, err error) {
			if err != nil {
				log.Debugf("Connection disconnect %s", err.Error())
			} else {
				log.Debugf("Connection disconnect no error")
			}
		}
		opts.ClosedCB = func(_ *natspkg.Conn) {
			log.Debugf("NATS Connection closed")
		}
		opts.ReconnectedCB = func(_ *natspkg.Conn) {
			log.Debugf("Connection Reconnect")
		}
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
			ret = nc
			done = true
		}
		errToReturn = nil
	}
	log.Infof("Leaving NATS Init ")
	return ret, errToReturn
}
