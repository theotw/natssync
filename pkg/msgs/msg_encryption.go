/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	v1 "github.com/theotw/natssync/pkg/bridgemodel/generated/v1"
	"github.com/theotw/natssync/pkg/persistence"
	"github.com/theotw/natssync/pkg/types"
)

func InitCloudKey() error {
	err := persistence.InitLocationKeyStore()
	if err != nil {
		return err
	}
	t := persistence.GetKeyStore()

	//first checkout the keys
	_, locationDataErr := t.ReadKeyPair("")
	if locationDataErr != nil {
		log.WithError(err).Debugf("Unable to find cloud master keys")
		err = GenerateAndSaveKey(pkg.CLOUD_ID)
	}

	return err
}

func GenerateAndSaveKey(locationID string) error {
	pair, err := GenerateNewKeyPair()
	if err != nil {
		log.Errorf("Unable to generate new key pair %s", err.Error())
		return err
	}

	err = SaveKeyPair(locationID, pair)
	if err != nil {
		return err
	}

	return nil
}

func SaveKeyPair(locationID string, pair *rsa.PrivateKey) error {
	log.Infof("Saving key pair for %s", locationID)

	t := persistence.GetKeyStore()

	locationData, err := GetKeyPairLocationData(locationID, pair)
	if err != nil {
		return err
	}

	if err := t.WriteKeyPair(locationData); err != nil {
		return err
	}

	return nil
}

func GetKeyPairLocationData(locationID string, pair *rsa.PrivateKey) (*types.LocationData, error) {

	publicKey, err := encodePublicKeyAsBytes(&pair.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := encodePrivateKeyAsBytes(pair)
	if err != nil {
		return nil, err
	}
	locationData, err := types.NewLocationData(locationID, publicKey, privateKey, nil)
	if err != nil {
		return nil, err
	}

	return locationData, nil
}

func LoadPublicKey(locationID string) (*rsa.PublicKey, error) {
	t := persistence.GetKeyStore()
	locationData, err := t.ReadLocation(locationID)
	if err != nil {
		return nil, err
	}
	data, _ := pem.Decode(locationData.GetPublicKey())
	var ret *rsa.PublicKey
	pubKey, err := x509.ParsePKIXPublicKey(data.Bytes)
	if pubKey != nil && err == nil {
		ret = pubKey.(*rsa.PublicKey)
	}

	return ret, err
}

func LoadPrivateKey(keyID string) (*rsa.PrivateKey, error) {
	t := persistence.GetKeyStore()
	locationData, err := t.ReadKeyPair(keyID)
	if err != nil {
		return nil, err
	}

	data, _ := pem.Decode(locationData.GetPrivateKey())
	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	return privateKeyImported, err
}

func encodePrivateKeyAsBytes(key *rsa.PrivateKey) ([]byte, error) {
	fileBits := x509.MarshalPKCS1PrivateKey(key)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: fileBits,
	}
	var buf bytes.Buffer
	err := pem.Encode(&buf, privateKeyBlock)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodePublicKeyAsBytes(key *rsa.PublicKey) ([]byte, error) {
	pubFileBits, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubFileBits,
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, publicKeyBlock)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateNewKeyPair() (*rsa.PrivateKey, error) {
	log.Info("Generating new key pair")
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	return privateKey, nil

}

func PutObjectInEnvelope(ob interface{}, senderID string, recipientID string) (*MessageEnvelope, error) {
	bits, err := json.Marshal(ob)
	if err != nil {
		return nil, err
	}
	msg, ok := ob.(*bridgemodel.NatsMessage)
	skipEncrpt := false
	if ok {
		parsedMsgSubject, _ := ParseSubject(msg.Subject)
		skipEncrpt = parsedMsgSubject.SkipEncryption
	}
	log.Tracef("Puting message in Envelope with Encryption=%v", !skipEncrpt)
	if skipEncrpt {
		return PutMessageInEnvelopev4(bits, senderID, recipientID)
	} else {
		return PutMessageInEnvelopeV3(bits, senderID, recipientID)
	}
}

func PutMessageInEnvelopeV3(msg []byte, senderID string, recipientID string) (*MessageEnvelope, error) {
	master, err := LoadPrivateKey("")
	if err != nil {
		return nil, err
	}

	ret := new(MessageEnvelope)
	msgKey := make([]byte, 16)
	if _, err = rand.Read(msgKey); err != nil {
		return nil, err
	}
	ret.MsgKey, err = rsaEncrypt(msgKey, recipientID)
	if err != nil {
		return nil, err
	}

	cipherMsg, err := DoAesCBCEncrypt(msg, msgKey)
	if err != nil {
		return nil, err
	}
	sigBits, err := SignData(cipherMsg, master)
	if err != nil {
		return nil, err
	}

	ret.Message = base64.StdEncoding.EncodeToString(cipherMsg)
	ret.EnvelopeVersion = ENVELOPE_VERSION_3
	ret.SenderID = senderID
	ret.RecipientID = recipientID
	ret.Signature = base64.StdEncoding.EncodeToString(sigBits)

	t := persistence.GetKeyStore()
	locationData, err := t.ReadLocation(recipientID)
	if err != nil {
		return nil, err
	}
	ret.KeyID = locationData.KeyID

	return ret, nil
}
func PutMessageInEnvelopev4(msg []byte, senderID string, recipientID string) (*MessageEnvelope, error) {
	master, err := LoadPrivateKey("")

	if err != nil {
		return nil, err
	}

	ret := new(MessageEnvelope)
	ret.MsgKey = BLANK_KEY

	sigBits, err := SignData(msg, master)
	if err != nil {
		return nil, err
	}

	ret.Message = base64.StdEncoding.EncodeToString(msg)
	ret.EnvelopeVersion = ENVELOPE_VERSION_4
	ret.SenderID = senderID
	ret.RecipientID = recipientID
	ret.Signature = base64.StdEncoding.EncodeToString(sigBits)

	return ret, nil
}

// NewAuthChallengeFromStoredKey Makes a new auth challenge using known stored private location ID
func NewAuthChallengeFromStoredKey() *v1.AuthChallenge {
	return NewAuthChallenge("")
}

// NewAuthChallenge Makes a new auth challenge, if KeyID is blank, it uses the current known key ID
func NewAuthChallenge(KeyID string) *v1.AuthChallenge {
	key, err := LoadPrivateKey(KeyID)
	if err != nil {
		log.Errorf("Unable to load private Key: %s", err.Error())
		return nil
	}
	timeStr := time.Now().String()
	sig, err := SignData([]byte(timeStr), key)
	if err != nil {
		log.Errorf("Error signing data %s", err.Error())
		return nil
	}
	ret := new(v1.AuthChallenge)
	ret.AuthChallengeA = timeStr
	ret.AuthChellengeB = base64.StdEncoding.EncodeToString(sig)
	return ret
}

func ValidateAuthChallenge(locationID string, challenge *v1.AuthChallenge) bool {
	pubKey, err := LoadPublicKey(locationID)
	if err != nil {
		log.Errorf("Error loading public key for location %s error: %s", locationID, err.Error())
		return false
	}
	sigBits, _ := base64.StdEncoding.DecodeString(challenge.AuthChellengeB)
	hash := sha256.Sum256([]byte(challenge.AuthChallengeA))

	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sigBits)
	if err != nil {
		log.Errorf("Signature Verification Failed %s %s", locationID, err.Error())
		return false
	}
	return true
}

func SignData(dataToSigh []byte, master *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.Sum256(dataToSigh)
	sigBits, err := rsa.SignPKCS1v15(rand.Reader, master, crypto.SHA256, hash[:])
	return sigBits, err
}

func PullObjectFromEnvelope(ob interface{}, envelope *MessageEnvelope) error {
	bits, err := PullMessageFromEnvelope(envelope)
	if err == nil {
		err = json.Unmarshal(bits, ob)
	}
	return err
}

func PullMessageFromEnvelope(envelope *MessageEnvelope) ([]byte, error) {
	switch envelope.EnvelopeVersion {
	case ENVELOPE_VERSION_1:
		return pullMessageFromEnvelopev1(envelope)

	case ENVELOPE_VERSION_2:
		return pullMessageFromEnvelopev2(envelope)

	case ENVELOPE_VERSION_3:
		return pullMessageFromEnvelopev3(envelope)
	case ENVELOPE_VERSION_4:
		return pullMessageFromEnvelopev4(envelope)
	}
	return nil, errors.New("invalid envelope")
}

//ok, Pull From Env 1 and 2 look almost the same, dont try to refactor common, let them live apart.
func pullMessageFromEnvelopev1(envelope *MessageEnvelope) ([]byte, error) {
	cipherMsgBits, err := base64.StdEncoding.DecodeString(envelope.Message)
	if err != nil {
		return nil, err
	}

	sigBits, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil {
		return nil, err
	}

	msgKey, err := rsaDecrypt(envelope.MsgKey, envelope.KeyID)
	if err != nil {
		return nil, err
	}

	publicKey, err := LoadPublicKey(envelope.SenderID)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(cipherMsgBits)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], sigBits)
	if err != nil {
		return nil, err
	}

	plainMsgBits, err := DoAesECBDecrypt(cipherMsgBits, msgKey)
	return plainMsgBits, err
}

func pullMessageFromEnvelopev2(envelope *MessageEnvelope) ([]byte, error) {
	cipherMsgBits, err := base64.StdEncoding.DecodeString(envelope.Message)
	if err != nil {
		return nil, err
	}

	sigBits, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil {
		return nil, err
	}

	msgKey, err := rsaDecrypt(envelope.MsgKey, envelope.KeyID)
	if err != nil {
		return nil, err
	}

	publicKey, err := LoadPublicKey(envelope.SenderID)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(cipherMsgBits)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], sigBits)
	if err != nil {
		return nil, err
	}

	plainMsgBits, err := DoAesCBCDecrypt(cipherMsgBits, msgKey)
	return plainMsgBits, err
}

// Used for messages for version 4, which are messages that are signed but not encrypted.  for SSL type traffic
func pullMessageFromEnvelopev4(envelope *MessageEnvelope) ([]byte, error) {
	plainBits, err := base64.StdEncoding.DecodeString(envelope.Message)
	if err != nil {
		return nil, err
	}

	sigBits, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil {
		return nil, err
	}

	publicKey, err := LoadPublicKey(envelope.SenderID)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(plainBits)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], sigBits)
	if err != nil {
		return nil, err
	}

	return plainBits, err
}

func pullMessageFromEnvelopev3(envelope *MessageEnvelope) ([]byte, error) {
	return pullMessageFromEnvelopev2(envelope)
}

func rsaEncrypt(plain []byte, clientID string) (string, error) {
	pubKey, err := LoadPublicKey(clientID)
	if err != nil {
		return "", err
	}
	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, plain)
	if err != nil {
		return "", err
	}
	cipherText := base64.StdEncoding.EncodeToString(cipher)
	return cipherText, nil
}

func rsaDecrypt(cipherText, keyID string) ([]byte, error) {
	privkey, err := LoadPrivateKey(keyID)
	if err != nil {
		return nil, err
	}
	cipher, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, privkey, cipher)
	return plain, err
}
