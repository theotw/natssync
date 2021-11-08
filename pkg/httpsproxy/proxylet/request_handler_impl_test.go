package proxylet_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

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
	requestHandler := proxylet.NewRequestHandlerDetailed(0, httpClient, natsClient, locationID)

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
