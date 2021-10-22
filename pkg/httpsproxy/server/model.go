/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

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
