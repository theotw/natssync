package msgs

const ENVELOPE_VERSION_1 = 1
const CLOUD_ID = "cloud"

type MessageEnvelope struct {
	EnvelopeVersion int
	RecipientID     string
	SenderID        string
	Message         string
	Signature       string
	MsgKey          string
}
