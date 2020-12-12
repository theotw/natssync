/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package msgs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessageEnvelope(t *testing.T) {
	msg := []byte("Hello World")
	envelope, err := PutMessageInEnvelope(msg, CLOUD_ID, "client1")
	if err != nil {
		t.Fatalf("Error with put in envelope %s", err.Error())
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

func Test_loadMaster(t *testing.T) {
	master, err := loadPrivateKey(CLOUD_ID)
	assert.Nil(t, err)
	assert.NotNil(t, master)
}
func Test_loadMasterPublic(t *testing.T) {
	master, err := LoadPublicKey(CLOUD_ID)
	assert.Nil(t, err)
	assert.NotNil(t, master)
}
func Test_encrpt(t *testing.T) {
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
