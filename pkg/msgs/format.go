package msgs

type MessageFormat interface {
	GeneratePayload(message string, mType string, source string) ([]byte, error)
	ValidateMsgFormat(msg []byte, ceEnabled bool) (bool, error)
	GetMsgFormat() MsgPayload
}
