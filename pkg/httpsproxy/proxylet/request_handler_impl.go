/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */
package proxylet

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg/httpsproxy/models"
	"github.com/theotw/natssync/pkg/httpsproxy/nats"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
)

type httpClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

type requestHandler struct {
	counter    int
	httpClient httpClientInterface
	natsClient nats.ClientInterface
	locationID string
}

func NewRequestHandler(LocationID string, natsClient nats.ClientInterface) *requestHandler {

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return NewRequestHandlerDetailed(0, httpClient, natsClient, LocationID)
}

func NewRequestHandlerDetailed(counter int, client httpClientInterface, natsClient nats.ClientInterface, locationID string) *requestHandler {
	return &requestHandler{
		counter:    counter,
		httpClient: client,
		natsClient: natsClient,
		locationID: locationID,
	}
}

func (rh *requestHandler) SetLocationID(locationID string) {
	rh.locationID = locationID
}

func (rh *requestHandler) HttpHandler(m *nats.Msg) {
	if string(m.Data) == "" {
		return
	}

	log.Printf("[#%d] Received on [%s]: '%s'", rh.counter, m.Subject, string(m.Data))

	var resp *server.HttpApiResponseMessage

	req, err := server.NewHttpApiReqMessageFromNatsMessage(m)
	if err != nil {
		log.WithError(err).Errorf("Error decoding http message")
		resp = server.NewHttpApiResponseMessageFromError(err)
		if err = rh.natsClient.Publish(m.Reply, []byte("ack")); err != nil {
			log.WithError(err).
				WithField("subject", m.Reply).
				Error("Failed to publish ack response")
		}
		_ = rh.natsClient.Flush()
		return
	}

	httpResp, err := rh.httpClient.Do(req.ToHttpRequest())
	if err != nil {
		log.WithError(err).Errorf("Error decoding http message")
		resp = server.NewHttpApiResponseMessageFromError(err)
	} else {
		resp = server.NewHttpApiResponseMessageFromHttpResponse(httpResp)
	}

	respBytes, err := json.Marshal(&resp)
	if err != nil {
		log.WithError(err).Error("Error marshaling response body")
		return
	}

	if err := rh.natsClient.Publish(m.Reply, respBytes); err != nil {
		log.WithError(err).
			WithField("subject", m.Reply).
			Error("Failed to publish response")
	}
	_ = rh.natsClient.Flush()
}

func (rh *requestHandler) HttpsHandler(msg *nats.Msg) {
	models.HandleConnectionRequest(msg, rh.locationID)
}
