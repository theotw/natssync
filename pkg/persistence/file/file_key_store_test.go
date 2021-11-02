/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package file_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/persistence/file"
	types "github.com/theotw/natssync/pkg/types"
	_ "github.com/theotw/natssync/tests/unit"
)

type testCase struct {
	name string
	run  func(t *testing.T, keystore *file.FileKeyStore)
}

func TestFileKeyStore(t *testing.T) {
	parentDir := os.TempDir()
	keystoreDir, err := ioutil.TempDir(parentDir, "keystoretest")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = os.RemoveAll(keystoreDir)
		assert.Nil(t, err, "failed to remove testing resource '%v': %v", keystoreDir, err)
	}() // clean up

	store, err := file.NewFileKeyStore(keystoreDir)
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

func testFileKeystoreWriteKeyPair(t *testing.T, keystore *file.FileKeyStore) {
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("id-%d", i)
		pubkey := fmt.Sprintf("pubkey%d", i)
		privkey := fmt.Sprintf("privkey%d", i)
		locationData, err := types.NewLocationData(id, []byte(pubkey), []byte(privkey), nil)
		assert.Nil(t, err)
		err = keystore.WriteKeyPair(locationData)
		assert.Nil(t, err)
	}
}

func testFileKeystoreReadKeyPair(t *testing.T, keystore *file.FileKeyStore) {
	locationData, err := keystore.ReadKeyPair("")
	assert.Nil(t, err)
	assert.Equal(t, "pubkey2", string(locationData.GetPublicKey()))
	assert.Equal(t, "privkey2", string(locationData.GetPrivateKey()))
	assert.Nil(t, err)
}

func testFileKeystoreGetLocationID(t *testing.T, keystore *file.FileKeyStore) {
	id := keystore.LoadLocationID("")
	assert.Equal(t, "id-2", id)
}

func testFileKeystoreRemoveKeyPair(t *testing.T, keystore *file.FileKeyStore) {
	err := keystore.RemoveKeyPair("")
	assert.Nil(t, err)
}

func testFileKeyStoreWriteLocation(t *testing.T, keystore *file.FileKeyStore) {
	key := "This is definitely a key"
	metadata := map[string]string{"foo": "bar"}
	locationData := types.LocationData{
		LocationID: "foo",
		PublicKey:  []byte(key),
		Metadata:   metadata,
	}
	err := keystore.WriteLocation(locationData)
	assert.Nil(t, err)
}

func testFileKeyStoreReadLocation(t *testing.T, keystore *file.FileKeyStore) {
	metadata := map[string]string{"foo": "bar"}
	locationData, err := keystore.ReadLocation("foo")
	assert.Nil(t, err)
	key := locationData.GetPublicKey()
	assert.Nil(t, err)
	assert.Equal(t, "This is definitely a key", string(key))
	assert.Equal(t, metadata, locationData.GetMetadata())
	locationData, err = keystore.ReadLocation("foo2")
	assert.Error(t, err)
	assert.Nil(t, locationData)
}

func testFileKeyStoreListKnownClients(t *testing.T, keystore *file.FileKeyStore) {
	expectedClients := []string{"foo"}
	clients, err := keystore.ListKnownClients()
	assert.Nil(t, err)
	assert.Equal(t, expectedClients, clients)
}

func testFileKeystoreRemoveLocation(t *testing.T, keystore *file.FileKeyStore) {
	err := keystore.RemoveLocation("foo")
	assert.Nil(t, err)
	err = keystore.RemoveLocation(pkg.CLOUD_ID)
	assert.Error(t, err)
}

func testFileKeystoreRemoveCloudMasterData(t *testing.T, keystore *file.FileKeyStore) {
	locationData := types.LocationData{
		LocationID: pkg.CLOUD_ID,
		PublicKey:  []byte("somekey"),
	}
	err := keystore.WriteLocation(locationData)
	assert.Nil(t, err)
	err = keystore.RemoveCloudMasterData()
	assert.Nil(t, err)

	lData, err := keystore.ReadLocation(pkg.CLOUD_ID)
	assert.Error(t, err)
	assert.Nil(t, lData)
}
