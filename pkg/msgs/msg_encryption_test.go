/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package msgs

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestEncryption(t *testing.T) {
	dirName, err := ioutil.TempDir("", "encrptTest")
	if !assert.Nil(t, err, "Error setting up test dir") {
		t.Fail()
	} else {
		os.Setenv("CERT_DIR", dirName)
		t.Run("load master keys", doTest_loadMaster)
		t.Run("Load public key", doTest_loadMasterPublic)
		t.Run("Test Encrpt", doTest_encrpt)
		t.Run("Test Envelope", doTestMessageEnvelope)
		os.RemoveAll(dirName)
	}

}
func doTestMessageEnvelope(t *testing.T) {
	InitCloudKey()
	generateAndSaveKey("client1")
	msg := []byte("Hello World")
	envelope, err := PutMessageInEnvelope(msg, CLOUD_ID, "client1")
	if err != nil {
		if !assert.Nil(t, err, "Error with put in envelope") {
			t.Fail()
		}

	}
	if msg == nil {
		t.Fatalf("Error with put in envelope %s", err.Error())
	}

	msg2, err := PullMessageFromEnvelope(envelope)
	if err != nil {
		t.Fatalf("Error with put in envelope %s", err.Error())
	}

	assert.Equal(t, msg, msg2)
}

func doTest_loadMaster(t *testing.T) {
	InitCloudKey()
	master, err := loadPrivateKey(CLOUD_ID)
	assert.Nil(t, err)
	assert.NotNil(t, master)
}
func doTest_loadMasterPublic(t *testing.T) {
	InitCloudKey()
	master, err := LoadPublicKey(CLOUD_ID)
	assert.Nil(t, err)
	assert.NotNil(t, master)
}
func doTest_encrpt(t *testing.T) {
	InitCloudKey()
	plainText := "hello async enc"
	cipher, err := rsaEncrypt([]byte(plainText), CLOUD_ID)
	if err != nil {
		t.Fatalf("Error %s", err.Error())
	}
	plain2, err := rsaDecrypt(cipher, CLOUD_ID)
	if err != nil {
		t.Fatalf("Error %s", err.Error())
	}
	plainText2 := string(plain2)
	assert.Equal(t, plainText, plainText2)
}
