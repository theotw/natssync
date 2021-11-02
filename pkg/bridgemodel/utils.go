/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package bridgemodel

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/utils"
)

func GenerateUUID() string {
	return utils.NewUUIDv4().String()
}

func NewHttpClient() *HttpClient {
	http := new(HttpClient)
	http.ServerURL = pkg.Config.CloudBridgeUrl
	return http
}

type HttpClient struct {
	ServerURL string
}

func (t *HttpClient) SendAuthorizedRequestWithBodyAndResp(method string, url string, bodyOb interface{}, respData interface{}) error {
	data, err := json.Marshal(bodyOb)
	if err != nil {
		return err
	}

	resp, err := t.sendAuthorizedRequest(method, url, data)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return errors.New(fmt.Sprintf("status code %s", resp.Status))
	}
	bits, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return e
	}
	if respData != nil {
		e = json.Unmarshal(bits, respData)
	}
	return e
}

func (t *HttpClient) sendAuthorizedRequest(method string, url string, body []byte) (response *http.Response, err error) {
	log.Tracef("Sending Auth Req")
	request, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	request.Header.Add("Accept", "application/json")
	// Send the request
	response, err = http.DefaultClient.Do(request)
	return
}

func ConfigureDefaultTransport() {
	// allows calls to https
	dt := http.DefaultTransport
	switch dt.(type) {
	case *http.Transport:
		if dt.(*http.Transport).TLSClientConfig == nil {
			dt.(*http.Transport).TLSClientConfig = &tls.Config{}
		}
		dt.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
	}
}
