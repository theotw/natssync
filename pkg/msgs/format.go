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
