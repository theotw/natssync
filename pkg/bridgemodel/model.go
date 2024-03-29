/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package bridgemodel

const REGISTRATION_AUTH_SUBJECT = "natssync.auth.registration"
const NATSPOST_AUTH_SUBJECT = "natssync.auth.natspost"
const REGISTRATION_QUERY_AUTH_SUBJECT = "natssync.auth.queryreg"
const REGISTRATION_AUTH_WILDCARD = "natssync.auth.*"
const UNREGISTRATION_AUTH_SUBJECT = "natssync.auth.unregister"
const REGISTRATION_LIFECYCLE_ADDED = "natssync.registration.lifecyle.added"
const REGISTRATION_LIFECYCLE_REMOVED = "natssync.registration.lifecyle.removed"
const ACCOUNT_LIFECYCLE_REMOVED = "account.lifecycle.removed" // TODO: This should probably be configurable

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

type RegistrationRequest struct {
	//user and secret provided by the client
	AuthToken string `json:"authToken"`
	//for reference, this is generated by the bridge server
	LocationID string `json:"locationID"`
}

type UnRegistrationRequest struct {
	AuthToken  string `json:"authToken"`
	LocationID string `json:"locationID"`
}

type RegistrationResponse struct {
	Success bool `json:"success"`
}

//use this when we just need an auth request that has no add on data
type GenericAuthRequest struct {
	AuthToken string `json:"authToken"`
}
type GenericAuthResponse struct {
	Success bool `json:"success"`
}
type UnRegistrationResponse struct {
	Success bool `json:"success"`
}

// RequestForLocationID send this message to get the location ID
const RequestForLocationID = "natssync.location.request"

// ResponseForLocationID this is the response subject, the data is the location ID, this message can be sent without a request, if the location ID changes
const ResponseForLocationID = "natssync.location.response"
