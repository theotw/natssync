/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */
package msgs

type MessageFormat interface {
	GeneratePayload(message string, mType string, source string) ([]byte, error)
	ValidateMsgFormat(msg []byte, ceEnabled bool) (bool, error)
}

var msgFormat MessageFormat

func GetMsgFormat() MessageFormat {
	return msgFormat
}

func NewMessageFormat() {
	msgFormat = new(CloudEventsFormat)
	return
}
