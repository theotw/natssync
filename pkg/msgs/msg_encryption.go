/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package msgs

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/theotw/natssync/pkg"
	"io/ioutil"
	"os"
	"path"
)

func loadPrivateKey(clientID string) (*rsa.PrivateKey, error) {
	basePath := pkg.GetEnvWithDefaults("CERT_DIR", "/certs")
	var keyFile string
	if clientID == CLOUD_ID {
		keyFile = "actmaster.pem"
	} else {
		keyFile = fmt.Sprintf("%s.pem", clientID)
	}
	masterPemPath := path.Join(basePath, keyFile)
	f, err := os.Open(masterPemPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	all, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	data, _ := pem.Decode(all)
	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	return privateKeyImported, err
}

//use a blank client ID for the master key
func LoadPublicKey(clientID string) (*rsa.PublicKey, error) {
	all, err := ReadPublicKeyFile(clientID)
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
func GenerateNewKeyPairPOCVersion(clientID string) error {
	pubFilePath := MakePublicKeyFileName(clientID)
	prvFilePath := MakePrivateFileName(clientID)

	pubOut, puberr := os.Create(pubFilePath)
	if puberr != nil {
		return puberr
	}
	defer pubOut.Close()
	privOut, priverr := os.Create(prvFilePath)
	if priverr != nil {
		return priverr
	}
	defer privOut.Close()
	pubBits, err := ReadPublicKeyFile("client1")
	if err != nil {
		return err
	}
	privBits, err := ReadPrivateKeyFile("client1")
	if err != nil {
		return err
	}
	pubOut.Write(pubBits)
	privOut.Write(privBits)
	return nil
}
func ReadPublicKeyFile(clientID string) ([]byte, error) {
	masterPemPath := MakePublicKeyFileName(clientID)
	f, err := os.Open(masterPemPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	all, err := ioutil.ReadAll(f)
	return all, err
}
func ReadPrivateKeyFile(clientID string) ([]byte, error) {
	masterPemPath := MakePublicKeyFileName(clientID)
	f, err := os.Open(masterPemPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	all, err := ioutil.ReadAll(f)
	return all, err
}

func MakePublicKeyFileName(clientID string) string {
	var keyFile string
	if clientID == CLOUD_ID {
		keyFile = "actmaster_public.pem"
	} else {
		keyFile = fmt.Sprintf("%s_public.pem", clientID)
	}

	basePath := pkg.GetEnvWithDefaults("CERT_DIR", "/certs")
	masterPemPath := path.Join(basePath, keyFile)
	return masterPemPath
}
func MakePrivateFileName(clientID string) string {
	var keyFile string
	if clientID == CLOUD_ID {
		keyFile = "actmaster.pem"
	} else {
		keyFile = fmt.Sprintf("%s.pem", clientID)
	}

	basePath := pkg.GetEnvWithDefaults("CERT_DIR", "/certs")
	masterPemPath := path.Join(basePath, keyFile)
	return masterPemPath
}
func PutObjectInEnvelope(ob interface{}, senderID string, recipientID string) (*MessageEnvelope, error) {
	bits, err := json.Marshal(ob)
	if err != nil {
		return nil, err
	}
	return PutMessageInEnvelope(bits, senderID, recipientID)
}
func PutMessageInEnvelope(msg []byte, senderID string, recipientID string) (*MessageEnvelope, error) {
	master, err := loadPrivateKey(senderID)

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

	privkey, err := loadPrivateKey(clientID)
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
