/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	_ "github.com/theotw/natssync/tests/unit"
)

type testCase struct {
	name string
	run func(t *testing.T, keystore *FileKeyStore)
}

func TestFileKeyStore(t *testing.T) {
	parentDir := os.TempDir()
	keystoreDir, err := ioutil.TempDir(parentDir, "keystoretest")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(keystoreDir) // clean up

	store, err := NewFileKeyStore(keystoreDir)
	if err != nil {
		log.Fatal(err)
	}

	tests := []testCase{
		{"Write Keypair", testFileKeystoreWriteKeyPair},
		{"Read Keypair", testFileKeystoreReadKeyPair},
		{"Get LocationID", testFileKeystoreGetLocationID},
		{"Remove Keypair", testFileKeystoreRemoveKeyPair},
		{"Write Location", testFileKeyStoreWriteLocation},
		{"Read Location", testFileKeyStoreReadLocation},
		{"List Clients", testFileKeyStoreListKnownClients},
		{"Remove Location", testFileKeystoreRemoveLocation},
		{"Remove Cloud Master Data", testFileKeystoreRemoveCloudMasterData},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.run(t, store)
		})
	}
}

func testFileKeystoreWriteKeyPair(t *testing.T, keystore *FileKeyStore) {
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("id-%d", i)
		pubkey := fmt.Sprintf("pubkey%d", i)
		privkey := fmt.Sprintf("privkey%d", i)
		err := keystore.WriteKeyPair(id, []byte(pubkey), []byte(privkey))
		assert.Nil(t, err)
	}
}

func testFileKeystoreReadKeyPair(t *testing.T, keystore *FileKeyStore) {
	pubkey, privkey, err := keystore.ReadKeyPair()
	assert.Equal(t, "pubkey2", string(pubkey))
	assert.Equal(t, "privkey2", string(privkey))
	assert.Nil(t, err)
}

func testFileKeystoreGetLocationID(t *testing.T, keystore *FileKeyStore) {
	id := keystore.GetLocationID()
	assert.Equal(t, "id-2", id)
}

func testFileKeystoreRemoveKeyPair(t *testing.T, keystore *FileKeyStore) {
	err := keystore.RemoveKeyPair()
	assert.Nil(t, err)
}

func testFileKeyStoreWriteLocation(t *testing.T, keystore *FileKeyStore) {
	key := "This is definitely a key"
	data := "metadata"
	err := keystore.WriteLocation("foo", []byte(key), data)
	assert.Nil(t, err)
}

func testFileKeyStoreReadLocation(t *testing.T, keystore *FileKeyStore) {
	key, data, err := keystore.ReadLocation("foo")
	assert.Nil(t, err)
	assert.Equal(t, "This is definitely a key", string(key))
	assert.Equal(t, "metadata", data)
	key, data, err = keystore.ReadLocation("foo2")
	assert.Error(t, err)
	assert.Nil(t, key)
	assert.Equal(t, "", data)
}

func testFileKeyStoreListKnownClients(t *testing.T, keystore *FileKeyStore) {
	expectedClients := []string{"foo"}
	clients, err := keystore.ListKnownClients()
	assert.Nil(t, err)
	assert.Equal(t, expectedClients, clients)
}

func testFileKeystoreRemoveLocation(t *testing.T, keystore *FileKeyStore) {
	err := keystore.RemoveLocation("foo")
	assert.Nil(t, err)
	err = keystore.RemoveLocation(CLOUD_ID)
	assert.Error(t, err)
}

func testFileKeystoreRemoveCloudMasterData(t *testing.T, keystore *FileKeyStore) {
	err := keystore.WriteLocation(CLOUD_ID, []byte("somekey"), "metadata")
	assert.Nil(t, err)
	err = keystore.RemoveCloudMasterData()
	assert.Nil(t, err)
	key, data, err := keystore.ReadLocation(CLOUD_ID)
	assert.Error(t, err)
	assert.Nil(t, key)
	assert.Equal(t, "", data)
}
