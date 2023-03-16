package integration

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	proxyServerHostPort = "localhost:8082"
	httpGetTestURL      = "http://testnginx"
	httpsGetTestURL     = "https://testnginx"
)

func getProxyUrl(locationID string) string {
	return fmt.Sprintf("http://%s:@%s", locationID, proxyServerHostPort)
}

func TestHttp(t *testing.T) {
	// the http proxy's locationID gets set to nats-sync client's locationID
	locationID := GetNatsSyncClientLocationID(t)

	proxyServerUrlString := getProxyUrl(locationID)
	proxyServerUrl, err := url.Parse(proxyServerUrlString)

	assert.Nil(t, err, "failed to parse proxy url: %v", err)
	httpClient := http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyServerUrl)}}

	resp, err := httpClient.Get(httpGetTestURL)
	assert.Nil(t, err, "failed to send http request: %v", err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err, "failed read response body: %v", err)
	log.Infof("Data from HTTP GET %s", string(bodyBytes))
	assert.NotNil(t, bodyBytes)
}

func TestHttps(t *testing.T) {
	// the http proxy's locationID gets set to nats-sync client's locationID
	locationID := GetNatsSyncClientLocationID(t)

	proxyServerUrlString := getProxyUrl(locationID)
	proxyServerUrl, err := url.Parse(proxyServerUrlString)

	assert.Nil(t, err, "failed to parse proxy url: %v", err)

	httpClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyServerUrl),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := httpClient.Get(httpsGetTestURL)
	if !assert.Nil(t, err, "failed to send http request: %v", err) {
		t.Fatalf("Invalid HTTPS Response %s", err.Error())
		return
	}

	assert.Equal(t, resp.StatusCode, http.StatusOK)
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err, "failed read response body: %v", err)

	assert.NotNil(t, bodyBytes)
}
