/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
)

const ENVELOPE_VERSION_1 = 1 //EBC AES
const ENVELOPE_VERSION_2 = 2 // CBC AES
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
	//loads the location ID for this client.  If not initialized, a blank string is returned
	LoadLocationID() string
	//loads the location ID
	SaveLocationID(locationID string) error
	ReadPrivateKeyData(locationID string) ([]byte, error)
	ReadPublicKeyData(locationID string) ([]byte, error)
	WritePublicKey(locationID string, buf []byte) error
	WritePrivateKey(locationID string, buf []byte) error
	ListKnownClients() ([]string, error)
}

var keystore LocationKeyStore

func parseKeystoreUrl(keystoreUrl string) (string, string, error) {
	log.Debugf("Parsing keystore URL: %s", keystoreUrl)
	ksTypeUrl := strings.SplitAfterN(keystoreUrl, "://", 2)
	if len(ksTypeUrl) != 2 {
		return "", "", fmt.Errorf("unable to parse url '%s'", keystoreUrl)
	}
	ksType := ksTypeUrl[0]
	ksUrl := ksTypeUrl[1]
	return ksType, ksUrl, nil
}

func GetKeyStore() LocationKeyStore {
	return keystore
}

func CreateLocationKeyStore(keystoreUrl string, conn *nats.Conn) (ret LocationKeyStore, err error) {
	keystoreType, keystoreUri, err := parseKeystoreUrl(keystoreUrl)
	if err != nil {
		log.Fatal(err)
	}

	switch keystoreType {
	case "file://":
		{
			ret, err = NewFileKeyStore(keystoreUri, conn)
			break
		}
	case "redis://":
		{
			ret, err = NewRedisLocationKeyStore(keystoreUri)
			break
		}
	case "mongodb://":
		{
			ret, err = NewMongoKeyStore(keystoreUri)
			break
		}
	}
	return
}

func InitLocationKeyStore(conn *nats.Conn) error {
	var err error
	keystore, err = CreateLocationKeyStore(pkg.Config.KeystoreUrl, conn)
	return err
}
