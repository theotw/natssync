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
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"io/ioutil"
	"net/http"
)

func GenerateUUID() string {
	return uuid.New().String()
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
	log.Tracef("Sending Auth Req\n")
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

type CloudEventsPayload struct {
	Source 		string		`json:"source"`
	Type		string		`json:"type"`
	SpecVersion	string		`json:"specversion"`
	ID			string		`json:"id"`
	Data		interface{}	`json:"data"`
}

func GenerateCloudEventsPayload(message string, mType string, source string) ([]byte, error) {
	var cloudEventsPayload CloudEventsPayload
	cloudEventsPayload.SpecVersion = "1.0"
	cloudEventsPayload.Type = mType
	cloudEventsPayload.ID = GenerateUUID()
	cloudEventsPayload.Source = source
	cloudEventsPayload.Data = message

	reqBytes := new(bytes.Buffer)
	err := json.NewEncoder(reqBytes).Encode(cloudEventsPayload)
	if err != nil {
		return nil, err
	}

	return reqBytes.Bytes(), nil
}

func ValidateCloudEventsMsgFormat(msg []byte, ceEnabled bool) (bool, error){
	if !ceEnabled {
		log.Info("Cloud Events disabled, skipping message validation")
		return true, nil
	}
	var cvMsg CloudEventsPayload
	var err error
	err = json.Unmarshal(msg, &cvMsg)
	if err != nil {
		log.Errorf("Failed to unmarshal json: %s", err.Error())
		return false, err
	}
	if cvMsg.SpecVersion != "1.0" {
		errMsg := fmt.Sprintf("Invalid ID for cloud event, expected 1.0, got %s", cvMsg.ID)
		err = errors.New(errMsg)
		return false, err
	}
	if cvMsg.Source == "" {
		errMsg := fmt.Sprintf("Source not set for cloud event")
		err = errors.New(errMsg)
		return false, err
	}
	if cvMsg.Type == "" {
		errMsg := fmt.Sprintf("Type not set for cloud event")
		err = errors.New(errMsg)
		return false, err
	}
	if cvMsg.ID == "" {
		errMsg := fmt.Sprintf("ID not set for cloud event")
		err = errors.New(errMsg)
		return false, err
	}
	return true, nil
}