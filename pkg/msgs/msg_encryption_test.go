/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"github.com/theotw/natssync/pkg"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/theotw/natssync/tests/unit"
)

func TestEncryption(t *testing.T) {
	parentDir := os.TempDir()
	keystoreDir, _ := ioutil.TempDir(parentDir, "keystoretest")
	pkg.Config.KeystoreUrl = "file://"+keystoreDir

	InitCloudKey()

	t.Run("load master keys", doTest_loadMaster)
	t.Run("Load public key", doTest_loadMasterPublic)
	t.Run("Test Encrpt", doTest_encrpt)
	t.Run("Test Envelope", doTestMessageEnvelope)
	t.Run("Location ID", doTestLocationID)
	t.Run("Auth Challenge", doTestAuthChallenge)

}
func doTestMessageEnvelope(t *testing.T) {

	GenerateAndSaveKey("client1")
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

	master, err := LoadPrivateKey(CLOUD_ID)
	assert.Nil(t, err)
	assert.NotNil(t, master)
}
func doTest_loadMasterPublic(t *testing.T) {
	master, err := LoadPublicKey(CLOUD_ID)
	assert.Nil(t, err)
	assert.NotNil(t, master)
}
func doTest_encrpt(t *testing.T) {
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

const unitestLocation = "unittestlocationID"

//test may seem out of place, but we need to know this works for challenge tests
func doTestLocationID(t *testing.T) {
	store := GetKeyStore()
	err := store.SaveLocationID(unitestLocation)
	assert.Nil(t, err, "Not expecting an error for location ID save")
	id := store.LoadLocationID()
	assert.Equal(t, unitestLocation, id)
}
func doTestAuthChallenge(t *testing.T) {
	store := GetKeyStore()
	err := GenerateAndSaveKey(unitestLocation)
	if !assert.Nil(t, err) {
		t.Fatalf("Unable to create and store keypair for auth challenge test %s", err.Error())
	}
	store.SaveLocationID(unitestLocation)
	challenge := NewAuthChallenge()
	if !assert.NotNil(t, challenge) {
		t.Fatal("Unable to creare auth challenge")
	}
	valid := ValidateAuthChallenge(unitestLocation, challenge)
	assert.True(t, valid, "Auth Challenge should be true")
	challenge.AuthChallengeA = "not what we think it should be"
	valid = ValidateAuthChallenge(unitestLocation, challenge)
	assert.False(t, valid, "Auth Challenge should be false")
}
