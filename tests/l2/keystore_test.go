/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package l2

import (
	"github.com/stretchr/testify/assert"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/msgs"
	"os"
	"path/filepath"
	"testing"
)

const pubkey = "testpub"
const privkey = "testprivate"

func TestFileKeyStore(t *testing.T) {

	tmpDir := os.TempDir()
	pubFile := filepath.Join(tmpDir, pubkey)
	privFile := filepath.Join(tmpDir, privkey)
	os.Remove(pubFile)
	os.Remove(privFile)
	pkg.Config.CertDir = tmpDir
	doKeyStore(t, "file")
	os.Remove(pubFile)
	os.Remove(privFile)

}
func TestRedisKeyStore(t *testing.T) {
	doKeyStore(t, "redis")
}
func doKeyStore(t *testing.T, ksType string) {
	store, err := msgs.CreateLocationKeyStore(ksType, nil)
	if !assert.Nil(t, err, "error should be nil") {
		t.Fatalf("Unable to initialize key store %s", err.Error())
	}

	testPub := []byte("test pub key")
	store.WritePublicKey(pubkey, testPub)
	testPrivate := []byte("test private key")
	store.WritePrivateKey(privkey, testPrivate)

	pubData, err := store.ReadPublicKeyData(pubkey)
	assert.Nil(t, err)
	assert.Equal(t, testPub, pubData)
	privateData, err := store.ReadPrivateKeyData(privkey)
	assert.Nil(t, err)
	assert.Equal(t, testPrivate, privateData)

	store.SaveLocationID("42")
	id := store.LoadLocationID()
	assert.Equal(t, "42", id)

}
