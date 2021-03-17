/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package errors

import "fmt"

// errors
const (
	ERROR_CODE_UNKNOWN = "unknown"
)

//network errors
const (
	NETWORK_ERROR_INVALID_URL      = "invalidurl"
	NETWORK_ERROR_IP_NOT_IN_SUBNET = "ip.not.in.subnet"
	INVALID_REGISTRATION_REQ       = "invalid.reg.request"
	INVALID_PUB_KEY                = "invalid.pub.key"
	INVALID_LOCATION_ID            = "invalid.location.id"
)

const (
	BRIDGE_ERROR = "bridgeerror"
	NETWOR_ERROR = "network"
)

type InternalError struct {
	Subsystem      string
	SubSystemError string
	Params         map[string]string
}

func (t *InternalError) Error() string {
	return fmt.Sprintf("%s.%s", t.Subsystem, t.SubSystemError)
}
func (t *InternalError) ErrorCode() string {
	return fmt.Sprintf("%s.%s", t.Subsystem, t.SubSystemError)
}
func NewInternalError(subSystem, code string, params map[string]string) *InternalError {
	var x InternalError
	x.Subsystem = subSystem
	x.SubSystemError = code
	x.Params = params
	return &x
}
func NewInternalErrorWithDataParam(subSystem, code string, param string) *InternalError {
	var x InternalError
	x.Subsystem = subSystem
	x.SubSystemError = code
	x.Params = map[string]string{"data": param}
	return &x
}
