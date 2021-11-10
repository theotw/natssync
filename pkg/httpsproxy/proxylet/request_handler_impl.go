/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */
package proxylet

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
	"github.com/theotw/natssync/pkg/httpsproxy/models"
	"github.com/theotw/natssync/pkg/httpsproxy/nats"
	"github.com/theotw/natssync/pkg/httpsproxy/net"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
)

type httpClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

type tcpClientInterface interface {
	DialTimeout(address string, timeout time.Duration) (io.ReadWriteCloser, error)
}

type requestHandler struct {
	counter    int
	httpClient httpClientInterface
	tcpClient  tcpClientInterface
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

	tcpClient := net.NewTcpClient()

	return NewRequestHandlerDetailed(0, httpClient, tcpClient, natsClient, LocationID)
}

func NewRequestHandlerDetailed(counter int, httpClient httpClientInterface, tcpClient tcpClientInterface, natsClient nats.ClientInterface, locationID string) *requestHandler {
	return &requestHandler{
		counter:    counter,
		httpClient: httpClient,
		tcpClient:  tcpClient,
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

// HttpsHandler handles a connection request to a service in the internal network
// We attempt to establish a socket to the target host:port.  If we can connect, we then setup two channels.
// One listener for packets to write to a socket and then a publisher to send packets read from the socket
func (rh *requestHandler) HttpsHandler(msg *nats.Msg) {
	var connectionMsg models.TCPConnectRequest
	var connectionResp models.TCPConnectResponse
	err := json.Unmarshal(msg.Data, &connectionMsg)
	if err != nil {
		log.WithError(err).Errorf("Error deicing connection request")
		connectionResp.State = "failed"
		connectionResp.StateDetails = err.Error()
	} else {
		targetSocket, err := rh.tcpClient.DialTimeout(connectionMsg.Destination, 10*time.Second)

		if err != nil {
			log.WithError(err).
				WithField("destination", connectionMsg.Destination).
				Errorf("Error dialing connection")
			connectionResp.State = "failed"
			connectionResp.StateDetails = err.Error() + " @ " + connectionMsg.Destination
		} else {
			connectionResp.State = "ok"
			outBoundSubject := httpproxy.MakeHttpsMessageSubject(
				connectionMsg.ProxyLocationID,
				connectionMsg.ConnectionID,
			)
			inBoundSubject := httpproxy.MakeHttpsMessageSubject(rh.locationID, connectionMsg.ConnectionID)

			go func() {
				err := models.StartBiDiNatsTunnel(
					rh.natsClient,
					outBoundSubject,
					inBoundSubject,
					connectionMsg.ConnectionID,
					targetSocket,
				)
				if err != nil {
					log.WithError(err).
						WithFields(
							log.Fields{
								"outBoundSubject": outBoundSubject,
								"inBoundSubject":  inBoundSubject,
								"ConnectionID":    connectionMsg.ConnectionID,
							},
						).Errorf("Failed to start bidiNatsTunnel")
				}
			}()

		}
	}

	respbits, jsonerr := json.Marshal(&connectionResp)
	if jsonerr == nil {
		if err = rh.natsClient.Publish(msg.Reply, respbits); err != nil {
			log.WithError(err).
				WithField("subject", msg.Reply).
				Errorf("failed to publish connection response")
		}
		_ = rh.natsClient.Flush()

	} else {
		log.WithError(jsonerr).Error("Error marshaling connection resp message")
	}
}
