/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package errors

import "fmt"

var errorStrings map[string]map[string]string

func init() {
	errorStrings = make(map[string]map[string]string)
	m := mkUSMap()
	errorStrings["en"] = m
}

func GetErrorString(locale, errcode string) string {
	if _, ok := errorStrings[locale]; !ok {
		locale = "en"
	}
	if _, ok := errorStrings[locale][errcode]; !ok {
		errcode = ERROR_CODE_UNKNOWN
	}
	return errorStrings[locale][errcode]
}
func mkUSMap() map[string]string {
	ret := make(map[string]string)
	ret[fmt.Sprintf("%s.%s", NETWOR_ERROR, NETWORK_ERROR_INVALID_URL)] = "Invalid URL provided"
	ret[fmt.Sprintf("%s.%s", NETWOR_ERROR, NETWORK_ERROR_IP_IN_USE)] = "The IP address is in use"
	ret[fmt.Sprintf("%s.%s", NETWOR_ERROR, NETWORK_ERROR_IP_NOT_IN_SUBNET)] = "The IP address is not in the given subnet"
	ret[fmt.Sprintf("%s.%s", NETWOR_ERROR, NETWORK_ERROR_INVALID_IP_FORMAT)] = "The IP does not appear to be a valid IP v4 or v6 address"
	ret[fmt.Sprintf("%s.%s", NETWOR_ERROR, NETWORK_ERROR_IP_NOT_ACCESSIBLE)] = "The given IP does not appear to be a active host"
	ret[fmt.Sprintf("%s.%s", NETWOR_ERROR, NETWORK_ERROR_CONNECT_FAIL)] = "Failed to connect to external service"
	ret[fmt.Sprintf("%s.%s", BRIDGE_ERROR, ERROR_CODE_UNKNOWN)] = "An unknown error occurred.  Please contact JG"

	return ret
}
