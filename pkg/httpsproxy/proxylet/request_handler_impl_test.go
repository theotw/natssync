package proxylet_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theotw/natssync/pkg/httpsproxy/models"
	"github.com/theotw/natssync/pkg/httpsproxy/nats"
	"github.com/theotw/natssync/pkg/httpsproxy/proxylet"
	utres "github.com/theotw/natssync/pkg/httpsproxy/proxylet/unittestresources"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
)

func TestNewRequestHandler(t *testing.T) {
	locationID := "dummyLocationID"
	natsClient := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	requestHandler := proxylet.NewRequestHandler(locationID, natsClient)
	assert.NotNil(t, requestHandler)
}

func TestHttpRequestHandler(t *testing.T) {
	locationID := "dummyLocationID"
	natsClient := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	httpClient := utres.NewMockHttpClient()
	tcpClient := utres.NewMockTcpClient()
	requestHandler := proxylet.NewRequestHandlerDetailed(0, httpClient, tcpClient, natsClient, locationID)

	req := server.HttpApiReqMessage{
		HttpMethod: "GET",
		HttpPath:   "/foo/bar",
		Body:       []byte("ok"),
		Headers: []server.HttpReqHeader{{
			Key:    "x-Dummy",
			Values: []string{"dummy"},
		}},
		Target: "localhost:80",
	}

	reqBytes, err := json.Marshal(req)
	assert.Nil(t, err)

	replyChannel := "testReplyChannel"
	natsMsg := &nats.Msg{
		Data:  reqBytes,
		Reply: replyChannel,
	}

	requestHandler.HttpHandler(natsMsg)

	_, err = natsClient.Subscribe(
		replyChannel,
		func(msg *nats.Msg) {
			data := &server.HttpApiResponseMessage{}
			err = json.Unmarshal(msg.Data, data)
			assert.Nil(t, err)
			assert.Equal(t, data.HttpStatusCode, http.StatusOK)
			assert.Equal(t, data.RespBody, "ok")
		},
	)
	assert.Nil(t, err)
}

func TestHttpsProxyConnectRequestResponse(t *testing.T) {
	locationID := "dummyLocationID"
	natsClient := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	httpClient := utres.NewMockHttpClient()
	tcpClient := utres.NewMockTcpClient()
	requestHandler := proxylet.NewRequestHandlerDetailed(0, httpClient, tcpClient, natsClient, locationID)

	serverLocationID := "dummyServerLocationID"
	tcpDestinationAddress := "localhost:8080"
	connectionID := "8c4f27d0-4a59-4507-9a91-1f46d3fad6eb"

	req := models.TCPConnectRequest{
		Destination:     tcpDestinationAddress,
		ProxyLocationID: serverLocationID,
		ConnectionID:    connectionID,
	}

	reqBytes, err := json.Marshal(req)
	assert.Nil(t, err)

	replyChannel := "testReplyChannel"
	natsMsg := &nats.Msg{
		Data:  reqBytes,
		Reply: replyChannel,
	}
	gotConnectionResponse := false
	requestHandler.HttpsHandler(natsMsg)

	_, err = natsClient.Subscribe(
		replyChannel,
		func(msg *nats.Msg) {
			gotConnectionResponse = true
			data := &models.TCPConnectResponse{}
			err = json.Unmarshal(msg.Data, data)
			assert.Nil(t, err)
			assert.Equal(t, data.State, "ok")
		},
	)
	_ = gotConnectionResponse
	assert.Nil(t, err)
	assert.True(t, gotConnectionResponse, "failed to get connection response")
}
