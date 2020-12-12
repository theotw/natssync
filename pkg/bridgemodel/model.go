/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package bridgemodel

//this is a generic message that will be encrypted and decrypted on the bridge.
//Its basicly the NATS data
type NatsMessage struct {
	Subject string
	Reply   string
	Data    []byte
}

type HttpReqHeader struct {
	Key    string
	Values []string
}
type K8SApiReqMessage struct {
	HttpMethod string
	HttpPath   string
	Body       []byte
	Headers    []HttpReqHeader
	Target     string
}
type K8SApiRespMessage struct {
	HttpStatusCode int
	RespBody       string
	RequestID      string
	Headers        map[string]string
}
