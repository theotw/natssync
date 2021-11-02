/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package httpproxy

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/theotw/natssync/pkg/msgs"
	"os"
)

const CLOUD_ID = "cloud-master"
const NATS_MSG_PREFIX = "natssyncmsg"
const HTTP_PROXY_API_ID = "httpproxy"
const HTTPS_PROXY_API_ID = "httpsproxy"
const HTTPS_PROXY_CONNECTION_REQUEST = "httpsproxy-connection"

func GetEnvWithDefaults(envKey string, defaultVal string) string {
	val := os.Getenv(envKey)
	if len(val) == 0 {
		val = defaultVal
	}
	return val
}

//makes a random reploy subject that will route from the client (south side) to the server side
func MakeReplyMessageSubject() string {
	locationID := GetMyLocationID()
	randomClientUUID := uuid.New().String()
	ret := fmt.Sprintf("%s.%s.%s.*", NATS_MSG_PREFIX, locationID, randomClientUUID)
	return ret
}

var locationID string

func GetMyLocationID() string {
	return locationID
}
func SetMyLocationID(id string) {
	locationID = id
}
func MakeMessageSubject(targetLocationID string, appID string) string {
	sub := fmt.Sprintf("%s.%s.%s", NATS_MSG_PREFIX, targetLocationID, appID)
	return sub
}

func MakeHttpsMessageSubject( targetLocationID string, connectionID string) string {
	sub := fmt.Sprintf("%s.%s.%s.%s,%s", NATS_MSG_PREFIX, targetLocationID,msgs.SKIP_ENCRYPTION_FLAG, HTTPS_PROXY_API_ID, connectionID)
	return sub
}
