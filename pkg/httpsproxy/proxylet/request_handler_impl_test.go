package proxylet_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theotw/natssync/pkg/httpsproxy/models"
	"github.com/theotw/natssync/pkg/httpsproxy/nats"
	"github.com/theotw/natssync/pkg/httpsproxy/proxylet"
	utres "github.com/theotw/natssync/pkg/httpsproxy/proxylet/unittestresources"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
)

func XTestNewRequestHandler(t *testing.T) {
	locationID := "dummyLocationID"
	natsClient := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	requestHandler, err := proxylet.NewRequestHandler(locationID, natsClient)
	assert.Nil(t, err)
	assert.NotNil(t, requestHandler)
}

func XTestHttpRequestHandler(t *testing.T) {
	locationID := "dummyLocationID"
	natsClient := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	httpClient := utres.NewMockHttpClient()
	tcpClient := utres.NewMockTcpClient()
	reqValidator := utres.NewMockRequestValidator()
	requestHandler := proxylet.NewRequestHandlerDetailed(
		0,
		httpClient,
		tcpClient,
		natsClient,
		locationID,
		reqValidator,
	)

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

func XTestHttpsProxyConnectRequestResponse(t *testing.T) {
	locationID := "dummyLocationID"
	natsClient := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	httpClient := utres.NewMockHttpClient()
	tcpClient := utres.NewMockTcpClient()
	reqValidator := utres.NewMockRequestValidator()
	requestHandler := proxylet.NewRequestHandlerDetailed(
		0,
		httpClient,
		tcpClient,
		natsClient,
		locationID,
		reqValidator,
	)

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
			assert.Equal(t, data.State, models.TCPConnectStateOK)
		},
	)
	_ = gotConnectionResponse
	assert.Nil(t, err)
	assert.True(t, gotConnectionResponse, "failed to get connection response")
}

func XTestHttpRequestHandlerValidationError(t *testing.T) {
	locationID := "dummyLocationID"
	natsClient := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	httpClient := utres.NewMockHttpClient()
	tcpClient := utres.NewMockTcpClient()
	reqValidator := utres.NewMockRequestValidator()
	errorMessage := "dummy validation error"
	reqValidator.ValidationError = fmt.Errorf(errorMessage)
	requestHandler := proxylet.NewRequestHandlerDetailed(
		0,
		httpClient,
		tcpClient,
		natsClient,
		locationID,
		reqValidator,
	)

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
			assert.Equal(t, data.HttpStatusCode, http.StatusBadGateway)
			assert.Equal(t, data.RespBody, errorMessage)
		},
	)
	assert.Nil(t, err)
}

func XTestHttpsProxyConnectRequestResponseValidationError(t *testing.T) {
	locationID := "dummyLocationID"
	natsClient := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	httpClient := utres.NewMockHttpClient()
	tcpClient := utres.NewMockTcpClient()
	reqValidator := utres.NewMockRequestValidator()
	errorMessage := "dummy validation error"
	reqValidator.ValidationError = fmt.Errorf(errorMessage)
	requestHandler := proxylet.NewRequestHandlerDetailed(
		0,
		httpClient,
		tcpClient,
		natsClient,
		locationID,
		reqValidator,
	)

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
			assert.Equal(t, data.State, models.TCPConnectStateFailed)
		},
	)
	_ = gotConnectionResponse
	assert.Nil(t, err)
	assert.True(t, gotConnectionResponse, "failed to get connection response")
}
