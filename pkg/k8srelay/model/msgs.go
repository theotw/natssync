/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package models

const K8SRelayRequestMessageSubjectSuffix = "k8s-relay-req"

type CallRequest struct {
	// Path is the path part of the URL for this call
	Path string `json:"path"`

	// Method the HTTP method to perform
	Method string `json:"method"`
	// InBody is the input body for the call, which may be nil
	InBody []byte `json:"inBody,omitempty"`
}

type CallResponse struct {
	// Path is the path part of the URL for this call
	Path string `json:"path"`

	Headers map[string]string `json:"headers"`

	StatusCode int `json:"statusCode"`
	// InBody is the input body for the call, which may be nil
	OutBody []byte `json:"inBody,omitempty"`
}

func NewCallResponse() *CallResponse {
	x := new(CallResponse)
	x.Headers = make(map[string]string, 0)
	return x
}
func (t *CallResponse) AddHeader(k, v string) {
	t.Headers[k] = v
}
