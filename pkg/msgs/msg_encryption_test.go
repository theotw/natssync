/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/persistence"
	"github.com/theotw/natssync/pkg/types"

	"github.com/stretchr/testify/assert"

	_ "github.com/theotw/natssync/tests/unit"
)

func TestEncryption(t *testing.T) {
	parentDir := os.TempDir()
	keystoreDir, _ := ioutil.TempDir(parentDir, "keystoretest")
	pkg.Config.KeystoreUrl = "file://" + keystoreDir
	metadata := map[string]string{"foo": "bar"}

	if err := InitCloudKey(); err != nil {
		t.Fatal(err)
	}

	store := persistence.GetKeyStore()
	defer func() {
		err := store.RemoveKeyPair("")
		assert.Nil(t, err, "failed to remove key pair: %v", err)
	}()

	locationData, err := store.ReadKeyPair("")
	if err != nil {
		t.Fatal(err)
	}

	writeLocationData, err := types.NewLocationData(pkg.CLOUD_ID, locationData.GetPublicKey(), nil, metadata)
	assert.Nil(t, err)
	if err = store.WriteLocation(*writeLocationData); err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = store.RemoveCloudMasterData()
		assert.Nil(t, err, "failed to remove cloud master data: %v", err)
	}()

	pair, err := GenerateNewKeyPair()
	if err != nil {
		t.Fatalf("Unable to generate new key pair %s", err)
	}
	key, err := encodePublicKeyAsBytes(&pair.PublicKey)
	clientLocationData, err := types.NewLocationData("client1", key, nil, metadata)
	assert.Nil(t, err)
	if err = store.WriteLocation(*clientLocationData); err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := store.RemoveLocation("client1")
		assert.Nil(t, err, "failed to remove location data for 'client1': %v", err)
	}()

	t.Run("Load master private", doTest_loadMasterPrivate)
	t.Run("Load location public", doTest_loadClientPublic)
	t.Run("Test Encrypt", doTest_encrpt)
	t.Run("Test Envelope", doTestMessageEnvelope)
	t.Run("Test v4 Envelope ID", doTestMessageEnvelopev4)
	t.Run("Auth Challenge", doTestAuthChallenge)
	t.Run("Location ID", doTestLocationID)

}
func doTestMessageEnvelope(t *testing.T) {
	msg := []byte("Hello World")
	envelope, err := PutMessageInEnvelopeV3(msg, pkg.CLOUD_ID, pkg.CLOUD_ID)
	if err != nil {
		if !assert.Nil(t, err, "Error with put in envelope") {
			t.Fail()
		}
	}
	if msg == nil {
		t.Fatalf("Error with put in envelope %s", err)
	}

	msg2, err := PullMessageFromEnvelope(envelope)
	if err != nil {
		t.Fatalf("Error with put in envelope %s", err)
	}

	assert.Equal(t, msg, msg2)
}
func doTestMessageEnvelopev4(t *testing.T) {
	msg := []byte("Hello World")
	envelope, err := PutMessageInEnvelopev4(msg, pkg.CLOUD_ID, pkg.CLOUD_ID)
	if err != nil {
		if !assert.Nil(t, err, "Error with put in envelope") {
			t.Fail()
		}
	}
	if msg == nil {
		t.Fatalf("Error with put in envelope %s", err)
	}

	msg2, err := PullMessageFromEnvelope(envelope)
	if err != nil {
		t.Fatalf("Error with put in envelope %s", err)
	}

	assert.Equal(t, msg, msg2)
}

func doTest_loadMasterPrivate(t *testing.T) {
	master, err := LoadPrivateKey("")
	assert.Nil(t, err)
	assert.NotNil(t, master)
}

func doTest_loadClientPublic(t *testing.T) {
	client, err := LoadPublicKey("client1")
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func doTest_encrpt(t *testing.T) {
	plainText := "hello async enc"
	cipher, err := rsaEncrypt([]byte(plainText), pkg.CLOUD_ID)
	if err != nil {
		t.Fatal(err)
	}
	plain2, err := rsaDecrypt(cipher, "")
	if err != nil {
		t.Fatal(err)
	}
	plainText2 := string(plain2)
	assert.Equal(t, plainText, plainText2)
}

//test may seem out of place, but we need to know this works for challenge tests
func doTestLocationID(t *testing.T) {
	unitTestLocation := "unittestlocationID"
	store := persistence.GetKeyStore()
	err := store.WriteKeyPair(unitTestLocation, nil, nil)
	assert.Nil(t, err)
	assert.Nil(t, err, "Not expecting an error for location ID save")
	id := store.LoadLocationID("")
	assert.Equal(t, "unittestlocationID", id)
}

func doTestAuthChallenge(t *testing.T) {
	challenge := NewAuthChallenge("")
	if !assert.NotNil(t, challenge) {
		t.Fatal("Unable to create auth challenge")
	}
	valid := ValidateAuthChallenge(pkg.CLOUD_ID, challenge)
	assert.True(t, valid, "Auth Challenge should be true")
	challenge.AuthChallengeA = "not what we think it should be"
	valid = ValidateAuthChallenge(pkg.CLOUD_ID, challenge)
	assert.False(t, valid, "Auth Challenge should be false")
}
