/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg/httpsproxy/nats"
)

// RequestForLocationID send this message to get the location ID
const RequestForLocationID = "natssync.location.request"

// ResponseForLocationID this is the response subject, the data is the location ID, this message can be sent without a request, if the location ID changes
const ResponseForLocationID = "natssync.location.response"

type HttpReqHeader struct {
	Key    string
	Values []string
}
type HttpApiReqMessage struct {
	HttpMethod string
	HttpPath   string
	Body       []byte
	Headers    []HttpReqHeader
	Target     string
}
type HttpApiResponseMessage struct {
	HttpStatusCode int
	RespBody       string
	RequestID      string
	Headers        map[string]string
}

func NewHttpApiReqMessageFromNatsMessage(m *nats.Msg) (*HttpApiReqMessage, error) {
	req := &HttpApiReqMessage{}
	if err := json.Unmarshal(m.Data, &req); err != nil {
		return nil, err
	}
	return req, nil
}

func (req *HttpApiReqMessage) ToHttpRequest() *http.Request {
	url := fmt.Sprintf("http://%s%s", req.Target, req.HttpPath)
	reader := bytes.NewReader(req.Body)
	httpRequest, _ := http.NewRequest(req.HttpMethod, url, reader)
	for _, x := range req.Headers {
		httpRequest.Header.Add(x.Key, x.Values[0])
	}

	return httpRequest
}

func NewHttpApiResponseMessageFromHttpResponse(resp *http.Response) *HttpApiResponseMessage {
	output := &HttpApiResponseMessage{}

	output.HttpStatusCode = resp.StatusCode
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		output.RespBody = string(bodyBytes)
	} else {
		log.WithError(err).Errorf("Got an error reading a resp body from http api")
		output.RespBody = err.Error()
	}
	output.Headers = make(map[string]string)

	for k, v := range resp.Header {
		if len(v) > 0 {
			output.Headers[k] = v[0]
		} else {
			output.Headers[k] = ""
		}
	}

	return output
}

func NewHttpApiResponseMessageFromError(err error) *HttpApiResponseMessage {
	return &HttpApiResponseMessage{
		HttpStatusCode: http.StatusBadGateway,
		RespBody:       err.Error(),
	}
}
