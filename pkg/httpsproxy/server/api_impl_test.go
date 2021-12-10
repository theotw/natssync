package server_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
	utres "github.com/theotw/natssync/pkg/httpsproxy/proxylet/unittestresources"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
	"github.com/theotw/natssync/pkg/httpsproxy/server/utils"
)

func TestServer_RouteHandler_About(t *testing.T) {

	natsConn := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	serverObj := server.NewServerDetailed("dummyLocationID", natsConn, true)

	resWriter := utils.NewResponseWriter(nil)

	dummyGinCxt := &gin.Context{
		Request: &http.Request{
			Method:     http.MethodGet,
			RequestURI: "/about/",
		},
		Writer: resWriter,
	}

	serverObj.RouteHandler(dummyGinCxt)
	assert.Equal(t, resWriter.GetStatus(), http.StatusOK)

	res := &server.AboutResponse{}
	err := json.Unmarshal(resWriter.GetBody(), res)
	assert.Nil(t, err)
	assert.NotEmpty(t, res.AppVersion)
}

func TestServer_RouteHandler_HealthCheck(t *testing.T) {
	natsConn := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	serverObj := server.NewServerDetailed("dummyLocationID", natsConn, true)

	resWriter := utils.NewResponseWriter(nil)

	dummyGinCxt := &gin.Context{
		Request: &http.Request{
			Method:     http.MethodGet,
			RequestURI: "/healthcheck/",
		},
		Writer: resWriter,
	}

	serverObj.RouteHandler(dummyGinCxt)
	assert.Equal(t, resWriter.GetStatus(), http.StatusOK)
}

func TestServer_RouteHandler_Metrics(t *testing.T) {
	natsConn := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	serverObj := server.NewServerDetailed("dummyLocationID", natsConn, true)

	resWriter := utils.NewResponseWriter(nil)

	dummyGinCxt := &gin.Context{
		Request: &http.Request{
			Method:     http.MethodGet,
			RequestURI: "/metrics/",
		},
		Writer: resWriter,
	}

	serverObj.RouteHandler(dummyGinCxt)
	// TODO Update this test
	assert.Equal(t, resWriter.GetStatus(), http.StatusNotImplemented)
}

func TestServer_RouteHandler_BadUrl(t *testing.T) {
	natsConn := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	serverObj := server.NewServerDetailed("dummyLocationID", natsConn, true)

	resWriter := utils.NewResponseWriter(nil)

	dummyGinCxt := &gin.Context{
		Request: &http.Request{
			Method:     http.MethodGet,
			RequestURI: "/foo/",
		},
		Writer: resWriter,
	}

	serverObj.RouteHandler(dummyGinCxt)
	assert.Equal(t, resWriter.GetStatus(), http.StatusServiceUnavailable)
}

func TestServer_RouteHandler(t *testing.T) {
	locationID := "dummyLocationID"
	natsConn := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	serverObj := server.NewServerDetailed(locationID, natsConn, true)

	resWriter := utils.NewResponseWriter(nil)

	headers := http.Header{}
	base64EncodedCred := base64.StdEncoding.EncodeToString([]byte("testLocationID:"))
	proxyAuthValue := fmt.Sprintf("%s%s", server.BASIC_AUTH_PREFIX, base64EncodedCred)
	headers.Add(server.ProxyAuthHeader, proxyAuthValue)

	urlObject, err := url.Parse("http://dummyHost:80/foo")
	assert.Nil(t, err)

	dummyGinCxt := &gin.Context{
		Request: &http.Request{
			Method:     http.MethodGet,
			URL:        urlObject,
			RequestURI: urlObject.Path,
			Host:       urlObject.Host,
			Header:     headers,
			Body:       io.NopCloser(bytes.NewReader([]byte(""))),
		},
		Writer: resWriter,
	}

	httpResponse := server.HttpApiResponseMessage{
		HttpStatusCode: 200,
		RespBody:       "foo",
		RequestID:      "ebae4d09-115f-46eb-b727-bfa8465cd03b",
		Headers:        make(map[string]string),
	}

	respBytes, err := json.Marshal(httpResponse)
	assert.Nil(t, err)
	testResponseSubject := httpproxy.MakeReplyMessageSubject(locationID, true)
	err = natsConn.Publish(testResponseSubject, respBytes)
	assert.Nil(t, err)

	serverObj.RouteHandler(dummyGinCxt)

	// validate that the response was successfully published to Nats
	assert.Equal(t, len(natsConn.Queues[testResponseSubject].Queue), 1)
	assert.Equal(t, respBytes, natsConn.Queues[testResponseSubject].Queue[0].Data)
}
