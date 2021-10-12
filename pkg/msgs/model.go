/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"errors"
	"fmt"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
)

const ENVELOPE_VERSION_1 = 1 //EBC AES
const ENVELOPE_VERSION_2 = 2 // CBC AES
const CLOUD_ID = "cloud-master"
const ECHOLET_SUFFIX = "echolet"
const ECHO_SUBJECT_BASE = "echo"
const NATSSYNC_MESSAGE_PREFIX = "natssyncmsg"

type MessageEnvelope struct {
	EnvelopeVersion int
	RecipientID     string
	SenderID        string
	Message         string
	Signature       string
	MsgKey          string
}

type LocationKeyStore interface {
	WriteKeyPair(locationID string, publicKey []byte, privateKey []byte) error
	ReadKeyPair() ([]byte, []byte, error)
	RemoveKeyPair() error
	LoadLocationID() string
	WriteLocation(locationID string, buf []byte, metadata map[string]string) error
	ReadLocation(locationID string) ([]byte, map[string]string, error)
	RemoveLocation(locationID string) error
	RemoveCloudMasterData() error
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

func CreateLocationKeyStore(keystoreUrl string) (ret LocationKeyStore, err error) {
	keystoreType, keystoreUri, err := parseKeystoreUrl(keystoreUrl)
	if err != nil {
		log.Fatal(err)
	}

	switch keystoreType {
	case "file://":
		{
			ret, err = NewFileKeyStore(keystoreUri)
			break
		}
	case "mongodb://":
		{
			ret, err = NewMongoKeyStore(keystoreUri)
			break
		}
	default:
		{
			ret, err = nil, fmt.Errorf("unsupported keystore type %s", keystoreType)
		}
	}
	return
}

func InitLocationKeyStore() error {
	var err error
	keystore, err = CreateLocationKeyStore(pkg.Config.KeystoreUrl)
	return err
}

func MakeNBReplySubject() string {
	replySubject := fmt.Sprintf("%s.%s.%s", NATSSYNC_MESSAGE_PREFIX, CLOUD_ID, bridgemodel.GenerateUUID())
	return replySubject
}

func MakeEchoSubject(clientID string) string {
	subject := fmt.Sprintf("%s.%s.%s", NATSSYNC_MESSAGE_PREFIX, clientID, ECHO_SUBJECT_BASE)
	return subject
}
func MakeMessageSubject(locationID string, params string) string {
	if len(params) == 0 {
		return fmt.Sprintf("%s.%s", NATSSYNC_MESSAGE_PREFIX, locationID)
	}
	return fmt.Sprintf("%s.%s.%s", NATSSYNC_MESSAGE_PREFIX, locationID, params)
}

type ParsedSubject struct {
	OriginalSubject string
	LocationID      string
	AppData         []string //dotted strings parts after the location ID
}

func ParseSubject(subject string) (*ParsedSubject, error) {
	parts := strings.Split(subject, ".")
	if len(parts) < 2 || (parts[0] != NATSSYNC_MESSAGE_PREFIX) {
		return nil, errors.New("invalid.message.subject")
	}
	ret := new(ParsedSubject)
	ret.LocationID = parts[1]
	ret.AppData = parts[2:]
	ret.OriginalSubject = subject

	return ret, nil
}
