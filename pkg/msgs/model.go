/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

const ENVELOPE_VERSION_1 = 1
const CLOUD_ID = "cloud-master"
const NB_MSG_PREFIX = "natssync-nb"
const SB_MSG_PREFIX = "natssync-sb"
const ECHOLET_SUFFIX = "echolet"
const ECHO_SUBJECT_BASE = "echo"

type MessageEnvelope struct {
	EnvelopeVersion int
	RecipientID     string
	SenderID        string
	Message         string
	Signature       string
	MsgKey          string
}

type LocationKeyStore interface {
	ReadPrivateKeyData(locationID string) ([]byte, error)
	ReadPublicKeyData(locationID string) ([]byte, error)
	WritePublicKey(locationID string, buf []byte) error
	WritePrivateKey(locationID string, buf []byte) error
}

var keystore LocationKeyStore

func GetKeyStore() LocationKeyStore {
	return keystore
}
func InitLocationKeyStore() error {
	ret, err := NewFileKeyStore()
	keystore = ret
	return err
}
