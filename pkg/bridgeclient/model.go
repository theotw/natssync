/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudclient

import "os"

// BiDiMessageHandler Generic interface for bi directional message handlers
type BiDiMessageHandler interface {

	// StartMessageHandler Starts the handler
	StartMessageHandler(clientID string) error

	// StopMessageHandler stops the handler
	StopMessageHandler()

	// GetHandlerType gets the type of the handler for debug purposes
	GetHandlerType() string
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