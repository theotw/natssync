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
	ret[fmt.Sprintf("%s.%s", BRIDGE_ERROR, ERROR_CODE_UNKNOWN)] = "An unknown error occurred.  "
	ret[fmt.Sprintf("%s.%s", BRIDGE_ERROR, INVALID_REGISTRATION_REQ)] = "The registration request was rejected by the registration auth system "
	ret[fmt.Sprintf("%s.%s", BRIDGE_ERROR, INVALID_PUB_KEY)] = "The given public key was not valid. "

	return ret
}
