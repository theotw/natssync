/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */
package proxylet

import (
	"os"
	"sync"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"

	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
	"github.com/theotw/natssync/pkg/httpsproxy/models"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
)

const (
	defaultLocationID = "proxylet"
	locationIDEnvVar  = "DEFAULT_LOCATION_ID"
)

type RequestHandlerInterface interface {
	HttpHandler(m *nats.Msg)
	HttpsHandler(msg *nats.Msg)
	SetLocationID(locationID string)
}

type proxylet struct {
	subMutex          sync.Mutex
	httpSubscription  *nats.Subscription
	httpsSubscription *nats.Subscription
	natsClient        *nats.Conn
	locationID        string
	requestHandler    RequestHandlerInterface
}

func getLocationIDFromEnv() string {
	value, ok := os.LookupEnv(locationIDEnvVar)
	if !ok {
		return defaultLocationID
	}
	return value
}

func NewProxylet() (*proxylet, error) {

	natsClient, err := getInitializedNatsClient()
	if err != nil {
		return nil, err
	}

	defaultLocationID := getLocationIDFromEnv()

	return NewProxyletDetailed(
		natsClient,
		defaultLocationID,
		NewRequestHandler(defaultLocationID, natsClient),
	), nil

}

func NewProxyletDetailed(natsClient *nats.Conn, locationID string, handler RequestHandlerInterface) *proxylet {
	return &proxylet{
		natsClient:     natsClient,
		locationID:     locationID,
		requestHandler: handler,
	}
}

func getInitializedNatsClient() (*nats.Conn, error) {
	if err := models.InitNats(); err != nil {
		return nil, err
	}
	return models.GetNatsClient(), nil
}

func (p *proxylet) setupQueueSubscriptions() {
	p.subMutex.Lock()
	defer p.subMutex.Unlock()

	if p.httpSubscription != nil {
		_ = p.httpSubscription.Unsubscribe()
		p.httpSubscription = nil
	}
	if p.httpsSubscription != nil {
		_ = p.httpsSubscription.Unsubscribe()
		p.httpsSubscription = nil
	}

	log.WithField("locationID", p.locationID).Info("Setting up subscriptions")

	p.requestHandler.SetLocationID(p.locationID)
	subj := httpproxy.MakeMessageSubject(p.locationID, httpproxy.HTTP_PROXY_API_ID)

	p.httpSubscription, _ = p.natsClient.Subscribe(subj, p.requestHandler.HttpHandler)
	log.Printf("Listening on [%s]", subj)

	conReqSubject := httpproxy.MakeMessageSubject(p.locationID, httpproxy.HTTPS_PROXY_CONNECTION_REQUEST)
	p.httpsSubscription, _ = p.natsClient.Subscribe(conReqSubject, p.requestHandler.HttpsHandler)

	if err := p.natsClient.LastError(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on [%s]", conReqSubject)
	_ = p.natsClient.Flush()
}

// configureNatsSyncLocationID: configures the locationID of the natssync client listening on the nats
// that this httpproxy is listening on; this allows configuration of private network to private network
// communication.
func (p *proxylet) configureNatsSyncLocationID() {
	/*
	 * 1. subscribe to queue to get  natssync client's locationID
	 * 2. send out request for natsync client's locationID
	 * when the locationID is received re-subscribe only to recieve messages for that locationID
	 */
	_, err := p.natsClient.Subscribe(server.ResponseForLocationID, func(msg *nats.Msg) {
		locationID := string(msg.Data)
		if locationID != "" {
			p.locationID = locationID
		}

		log.Infof("Using location ID %s", locationID)
		p.setupQueueSubscriptions()

	})
	if err != nil {
		log.WithError(err).Fatalf("Unable to talk to NATS")
	}

	_ = p.natsClient.Publish(server.RequestForLocationID, []byte(""))
}

func (p *proxylet) RunHttpProxylet() {
	p.setupQueueSubscriptions()
	p.configureNatsSyncLocationID()
}
