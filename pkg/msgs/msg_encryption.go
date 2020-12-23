/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
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
	log "github.com/sirupsen/logrus"
)

func InitCloudKey() error {
	err := InitLocationKeyStore()
	if err != nil {
		return err
	}
	//first checkout the keys
	_, keyErr := LoadPrivateKey(CLOUD_ID)
	if keyErr != nil {
		log.Debugf("Unable to find cloud master private key %s \n", keyErr.Error())
	}
	_, pubKeyErr := LoadPublicKey(CLOUD_ID)
	if pubKeyErr != nil {
		log.Debugf("Unable to find cloud master public key %s \n", pubKeyErr.Error())
	}

	if keyErr != nil || pubKeyErr != nil {
		err = GenerateAndSaveKey(CLOUD_ID)
	}
	return err
}
func GenerateAndSaveKey(locationID string) error {
	pair, err := generateNewKeyPair(locationID)
	if err != nil {
		log.Errorf("Unable to generate new key pair %s \n", err.Error())
		return err
	}
	err = StorePrivateKey(locationID, pair)
	if err != nil {
		return err
	}
	err = StorePublicKey(locationID, &pair.PublicKey)
	if err != nil {
		return err
	}

	return nil
}

func LoadPublicKey(locationID string) (*rsa.PublicKey, error) {
	t := GetKeyStore()
	all, err := t.ReadPublicKeyData(locationID)
	if err != nil {
		return nil, err
	}
	data, _ := pem.Decode(all)
	var ret *rsa.PublicKey
	pubKey, err := x509.ParsePKIXPublicKey(data.Bytes)
	if pubKey != nil && err == nil {
		ret = pubKey.(*rsa.PublicKey)
	}

	return ret, err
}

func LoadPrivateKey(locationID string) (*rsa.PrivateKey, error) {
	t := GetKeyStore()
	all, err := t.ReadPrivateKeyData(locationID)
	if err != nil {
		return nil, err
	}

	data, _ := pem.Decode(all)
	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	return privateKeyImported, err
}

func StorePrivateKey(locationID string, key *rsa.PrivateKey) error {
	t := GetKeyStore()
	fileBits := x509.MarshalPKCS1PrivateKey(key)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: fileBits,
	}
	var buf bytes.Buffer
	err := pem.Encode(&buf, privateKeyBlock)
	if err == nil {
		err = t.WritePrivateKey(locationID, buf.Bytes())
	}
	return err
}
func StorePublicKey(locationID string, key *rsa.PublicKey) error {
	t := GetKeyStore()
	pubFileBits, e := x509.MarshalPKIXPublicKey(key)
	if e != nil {
		return e
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubFileBits,
	}

	var buf bytes.Buffer
	err := pem.Encode(&buf, publicKeyBlock)
	if err == nil {
		err = t.WritePublicKey(locationID, buf.Bytes())
	}

	return err
}
func generateNewKeyPair(clientID string) (*rsa.PrivateKey, error) {
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
	return PutMessageInEnvelope(bits, senderID, recipientID)
}
func PutMessageInEnvelope(msg []byte, senderID string, recipientID string) (*MessageEnvelope, error) {
	master, err := LoadPrivateKey(senderID)

	if err != nil {
		return nil, err
	}

	ret := new(MessageEnvelope)
	msgKey := make([]byte, 16)
	rand.Read(msgKey)
	ret.MsgKey, err = rsaEncrypt(msgKey, recipientID)
	if err != nil {
		return nil, err
	}

	cipherMsg, err := DoAesEncrypt(msg, msgKey)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(cipherMsg)
	sigBits, err := rsa.SignPKCS1v15(rand.Reader, master, crypto.SHA256, hash[:])
	if err != nil {
		return nil, err
	}

	ret.Message = base64.StdEncoding.EncodeToString(cipherMsg)
	ret.EnvelopeVersion = ENVELOPE_VERSION_1
	ret.SenderID = senderID
	ret.RecipientID = recipientID
	ret.Signature = base64.StdEncoding.EncodeToString(sigBits)

	return ret, nil
}
func PullObjectFromEnvelope(ob interface{}, envelope *MessageEnvelope) error {
	bits, err := PullMessageFromEnvelope(envelope)
	if err == nil {
		err = json.Unmarshal(bits, ob)
	}
	return err
}
func PullMessageFromEnvelope(envelope *MessageEnvelope) ([]byte, error) {
	if envelope.EnvelopeVersion == ENVELOPE_VERSION_1 {
		return pullMessageFromEnvelopev1(envelope)
	}
	return nil, errors.New("invalid envelope")
}
func pullMessageFromEnvelopev1(envelope *MessageEnvelope) ([]byte, error) {
	cipherMsgBits, err := base64.StdEncoding.DecodeString(envelope.Message)
	if err != nil {
		return nil, err
	}

	sigBits, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil {
		return nil, err
	}
	msgKey, err := rsaDecrypt(envelope.MsgKey, envelope.RecipientID)
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
	plainMsgBits, err := DoAesDecrypt(cipherMsgBits, msgKey)
	return plainMsgBits, err
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
func rsaDecrypt(cipherText string, clientID string) ([]byte, error) {
	privkey, err := LoadPrivateKey(clientID)
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
