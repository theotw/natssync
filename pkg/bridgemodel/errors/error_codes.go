/*
 * Copyright (c) 2020. NetApp
 */

package errors

import "fmt"

// errors
const (
	ERROR_CODE_UNKNOWN = "unknown"
)

//network errors
const (
	NETWORK_ERROR_INVALID_URL       = "invalidurl"
	NETWORK_ERROR_INVALID_IP_FORMAT = "invalid.ip"
	NETWORK_ERROR_CONNECT_FAIL      = "connectfailed"
	NETWORK_ERROR_IP_IN_USE         = "ip.inuse"
	NETWORK_ERROR_IP_NOT_ACCESSIBLE = "ip.not.accessable"
	NETWORK_ERROR_IP_NOT_IN_SUBNET  = "ip.not.in.subnet"
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
func NewInernalError(subSystem, code string, params map[string]string) *InternalError {
	var x InternalError
	x.Subsystem = subSystem
	x.SubSystemError = code
	x.Params = params
	return &x
}
func NewInernalErrorWithDataParam(subSystem, code string, param string) *InternalError {
	var x InternalError
	x.Subsystem = subSystem
	x.SubSystemError = code
	x.Params = map[string]string{"data": param}
	return &x
}
