package proxylet_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
	"github.com/theotw/natssync/pkg/httpsproxy/proxylet"
	utres "github.com/theotw/natssync/pkg/httpsproxy/proxylet/unittestresources"
	"github.com/theotw/natssync/pkg/httpsproxy/server"
)


func TestLocationIDSetting(t *testing.T) {
	natsConn := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	natsSyncClientID := "testLocationID"
	err := natsConn.Publish(server.ResponseForLocationID, []byte(natsSyncClientID))
	assert.Nil(t, err)

	locationID := "dummyLocationID"
	requestHandler := utres.NewMockRequestHandler()
	mockProxylet := proxylet.NewProxyletDetailed(natsConn, locationID, requestHandler)
	mockProxylet.RunHttpProxylet()

	assert.Equal(t, natsSyncClientID, requestHandler.GetLocationID())
}

func TestHttpRequest(t *testing.T) {
	natsConn := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	locationID := "dummyLocationID"
	subj := httpproxy.MakeMessageSubject(locationID, httpproxy.HTTP_PROXY_API_ID)
	err := natsConn.Publish(subj, []byte(""))
	assert.Nil(t, err)

	requestHandler := utres.NewMockRequestHandler()

	mockProxylet := proxylet.NewProxyletDetailed(natsConn, locationID, requestHandler)
	mockProxylet.RunHttpProxylet()
	assert.True(t, requestHandler.InvokedHttpHandler())
	assert.False(t, requestHandler.InvokedHttpsHandler())
}

func TestHttpsRequest(t *testing.T) {
	natsConn := utres.NewMockNats(utres.NewDefaultMockNatsInput())
	locationID := "dummyLocationID"
	subj := httpproxy.MakeMessageSubject(locationID, httpproxy.HTTPS_PROXY_CONNECTION_REQUEST)
	err := natsConn.Publish(subj, []byte(""))
	assert.Nil(t, err)

	requestHandler := utres.NewMockRequestHandler()

	mockProxylet := proxylet.NewProxyletDetailed(natsConn, locationID, requestHandler)
	mockProxylet.RunHttpProxylet()
	assert.True(t, requestHandler.InvokedHttpsHandler())
	assert.False(t, requestHandler.InvokedHttpHandler())

}
