/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package l2

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const clientPrefix = "http://localhost:8080/bridge-client"
const serverPrefix = "http://localhost:8080/bridge-server"

type test_case struct {
	UrlSuffix      string
	ExpectedStatus int
}

func TestClientURLs(t *testing.T) {
	tests := []test_case{
		{"healthcheck", 200},
		{"1/healthcheck", 200},
		{"about", 200},
		{"1/about", 200},
		{"api/bridge_client_v1.yaml", 200},
		{"api/swagger.yaml", 200},
		{"api/", 200},
	}
	for _, test := range tests {
		t.Run(test.UrlSuffix, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", clientPrefix, test.UrlSuffix)
			status := get_test(url, t)
			assert.Equal(t, test.ExpectedStatus, status)
		})

	}
}
func TestServerURLs(t *testing.T) {
	tests := []test_case{
		{"healthcheck", 200},
		{"1/healthcheck", 200},
		{"about", 200},
		{"1/about", 200},
		{"api/bridge_server_v1.yaml", 200},
		{"api/swagger.yaml", 200},
		{"api/", 200},
	}
	for _, test := range tests {
		t.Run(test.UrlSuffix, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", serverPrefix, test.UrlSuffix)
			status := get_test(url, t)
			assert.Equal(t, test.ExpectedStatus, status)
		})

	}
}

//returns the status code or 0 on error (Error logged)
func get_test(url string, t *testing.T) int {
	resp, err := http.DefaultClient.Get(url)
	ret := 0
	if err != nil {
		logrus.Errorf("Error fetching URL:  %s  error %s", url, err.Error())
	} else {
		ret = resp.StatusCode
	}
	return ret
}
