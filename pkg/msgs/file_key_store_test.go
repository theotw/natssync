/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/theotw/natssync/pkg"
	_ "github.com/theotw/natssync/tests/unit"
	"io/ioutil"
	"os"
	"testing"
)

func TestKeyStore(t *testing.T) {
	parentDir := os.TempDir()
	keystoreDir, err := ioutil.TempDir(parentDir, "keystoretest")
	if err != nil {
		log.Fatal(err)
	}
	pkg.Config.CertDir = keystoreDir
	defer os.RemoveAll(keystoreDir) // clean up

	store, err := NewFileKeyStore(nil)
	if err != nil {
		log.Fatal(err)
	}

	t.Run("Location ID ", func(t *testing.T) {
		testLocationID(store, t)
	})
	t.Run("Save Pub Keys ", func(t *testing.T) {
		testSavePubKey(store, t)
	})

	t.Run("Read Pub Keys", func(t *testing.T) {
		testReadPubKey(store, t)
	})
	t.Run("Save Priv Keys ", func(t *testing.T) {
		testSavePrivKey(store, t)
	})

	t.Run("Read Priv Keys", func(t *testing.T) {
		testReadPrivKey(store, t)
	})

	t.Run("Load IDS ", func(t *testing.T) {
		testLoadIDs(store, t)
	})

}
func testLoadIDs(store LocationKeyStore, t *testing.T) {
	clients, err := store.ListKnownClients()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(clients))

}
func testSavePubKey(store LocationKeyStore, t *testing.T) {
	var err error
	err = store.WritePublicKey("1", []byte("one"))
	assert.Nil(t, err, "Expecting no error on write")
	store.WritePublicKey("2", []byte("two"))
	assert.Nil(t, err, "Expecting no error on write")
	store.WritePublicKey("3", []byte("three"))
	assert.Nil(t, err, "Expecting no error on write")
}
func testReadPubKey(store LocationKeyStore, t *testing.T) {
	data, err := store.ReadPublicKeyData("2")
	assert.Nil(t, err, "Not expecting error on read")
	assert.Equal(t, []byte("two"), data)
}
func testSavePrivKey(store LocationKeyStore, t *testing.T) {
	var err error
	err = store.WritePrivateKey("1", []byte("onep"))
	assert.Nil(t, err, "Expecting no error on write")
	store.WritePublicKey("2", []byte("twop"))
	assert.Nil(t, err, "Expecting no error on write")
	store.WritePublicKey("3", []byte("threep"))
	assert.Nil(t, err, "Expecting no error on write")
}
func testReadPrivKey(store LocationKeyStore, t *testing.T) {
	data, err := store.ReadPublicKeyData("2")
	assert.Nil(t, err, "Not expecting error on read")
	assert.Equal(t, []byte("twop"), data)
}

func testLocationID(store LocationKeyStore, t *testing.T) {
	expectedID := "location"
	err := store.SaveLocationID(expectedID)
	assert.Nil(t, err, "error saving location ID")
	locationID := store.LoadLocationID()
	assert.Equalf(t, expectedID, locationID, "expect the location to match")

}
