/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import "os"

type BiDiMessageHandler interface {
	StartMessageHandler(clientID string) error
	StopMessageHandler()
}


func NewBidiMessageHandler(serverURL string) BiDiMessageHandler{
	var ret BiDiMessageHandler
	if os.Getenv("TRANSPORTPROTO") == "websocket" {
		ret=NewWebSocketMessageHandler(serverURL)
	}else{
		//default to REST
		ret=NewRestMessageHandler(serverURL)
	}
	return ret
}