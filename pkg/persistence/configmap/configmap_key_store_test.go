/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

/*
These tests were taken from file_key_store_test.go and adapted to work with the configmap keystore.
These are not unit tests.
These tests are very useful for the developer to run (similar to unit tests) to ensure everything still works.

Note: Changes to the configmap (E.g. using keystore.WriteKeyPair()) are not reflected in the pod immediately. These
	  tests are using a fairly long sleep after writes to ensure everything works. This can be modified to be a
	  polling type wait instead of sleep if someone has the time and motivation to implement it.

How to run these tests:

These tests currently exercise the real K8s API and must be run inside a pod with correct Pod securityContext. It sounds
confusing, but the easy way is to just run the PrivateClusterAgent install, and then run the test pod inside that namespace.

1. Install the PrivateClusterAgent helm chart. See https://github.com/NetApp/privateclusteragent.
2. Build and push the test image:
	a. From the root of the repo:
	   'make baseimage &&  make testimage  IMAGE_TAG=<yourTag> IMAGE_REPO=docker.repo.eng.netapp.com/<yourRepo>'
	b. Push the image you just built. 'docker push <your-image-repo>/natssync-tests:<yourTag>'
3. Edit pkg/persistence/configmap/configmap_test_pod.yaml. Update 'image:' to point to the image you just pushed ^.
4. From the root of the repo:
   'kubectl apply -f pkg/persistence/configmap/configmap_test_pod.yaml -n <private cluster agent namespace>'

At this point, the pod will begin starting. Once it is running you can exec into the pod

5. Exec into the test pod:
	'kubectl exec -it configmap-keystore-test -n <namespace> -- /bin/sh'
6. ls the configmap mount dir. 'ls /<mount-path>' E.g. 'ls /data'
7. Exit the pod with 'exit'
8. Remove each file that was listed from 'f' using:
   kubectl patch configmap cluster-agent-configmap -n <namespace> --type=json -p='[{"op": "remove", "path": "/data/<filename>"}]
   E.g. 'kubectl patch configmap cluster-agent-configmap -n my-namespace --type=json -p='[{"op": "remove", "path": "/data/foo_locationData.json"}]''
9. Exec back into the pod with 'kubectl exec -it configmap-keystore-test -n <namespace> -- /bin/sh'
10. Repeat step 6, make sure the files are now gone. Fyi, the files can take a minute to clear after the last kubectl patch command.
11. Run the tests with './other-tests/configmap_keystore_amd64_linux.test'
*/

package configmap_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/persistence/configmap"
	"github.com/theotw/natssync/pkg/types"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	_ "github.com/theotw/natssync/tests/unit"
)

type testCase struct {
	name string
	run  func(t *testing.T, keystore *configmap.ConfigmapKeyStore)
}

func TestFileKeyStore(t *testing.T) {
	// Set to the configmap mount location
	configmapMountPath := "/configmap-data"

	store, err := configmap.NewConfigmapKeyStore(configmapMountPath)
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
		fmt.Printf("Starting test: %s\n", test.name)
		t.Run(test.name, func(t *testing.T) {
			test.run(t, store)
		})
	}
}

func testFileKeystoreWriteKeyPair(t *testing.T, keystore *configmap.ConfigmapKeyStore) {
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("id-%d", i)
		pubkey := fmt.Sprintf("pubkey%d", i)
		privkey := fmt.Sprintf("privkey%d", i)
		locationData, err := types.NewLocationData(id, []byte(pubkey), []byte(privkey), nil)
		assert.Nil(t, err)
		err = keystore.WriteKeyPair(locationData)
		assert.Nil(t, err)
		time.Sleep(120 * time.Second)
	}
}

func testFileKeystoreReadKeyPair(t *testing.T, keystore *configmap.ConfigmapKeyStore) {
	locationData, err := keystore.ReadKeyPair("")
	assert.Nil(t, err)
	assert.Equal(t, "pubkey2", string(locationData.GetPublicKey()))
	assert.Equal(t, "privkey2", string(locationData.GetPrivateKey()))
	assert.Nil(t, err)
}

func testFileKeystoreGetLocationID(t *testing.T, keystore *configmap.ConfigmapKeyStore) {
	id := keystore.LoadLocationID("")
	assert.Equal(t, "id-2", id)
}

func testFileKeystoreRemoveKeyPair(t *testing.T, keystore *configmap.ConfigmapKeyStore) {
	err := keystore.RemoveKeyPair("")
	assert.Nil(t, err)
}

func testFileKeyStoreWriteLocation(t *testing.T, keystore *configmap.ConfigmapKeyStore) {
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

func testFileKeyStoreReadLocation(t *testing.T, keystore *configmap.ConfigmapKeyStore) {
	time.Sleep(120 * time.Second)
	metadata := map[string]string{"foo": "bar"}
	locationData, err := keystore.ReadLocation("foo")
	assert.Nil(t, err)
	key := locationData.GetPublicKey()
	assert.Equal(t, "This is definitely a key", string(key))
	assert.Equal(t, metadata, locationData.GetMetadata())
	locationData, err = keystore.ReadLocation("foo2")
	assert.Error(t, err)
	assert.Nil(t, locationData)
}

func testFileKeyStoreListKnownClients(t *testing.T, keystore *configmap.ConfigmapKeyStore) {
	expectedClients := []string{"foo"}
	clients, err := keystore.ListKnownClients()
	assert.Nil(t, err)
	assert.Equal(t, expectedClients, clients)
}

func testFileKeystoreRemoveLocation(t *testing.T, keystore *configmap.ConfigmapKeyStore) {
	err := keystore.RemoveLocation("foo")
	assert.Nil(t, err)
	err = keystore.RemoveLocation(pkg.CLOUD_ID)
	assert.Error(t, err)
}

func testFileKeystoreRemoveCloudMasterData(t *testing.T, keystore *configmap.ConfigmapKeyStore) {
	time.Sleep(120 * time.Second)
	locationData := types.LocationData{
		LocationID: pkg.CLOUD_ID,
		PublicKey:  []byte("somekey"),
	}
	err := keystore.WriteLocation(locationData)
	assert.Nil(t, err)
	time.Sleep(120 * time.Second)
	err = keystore.RemoveCloudMasterData()
	assert.Nil(t, err)
	time.Sleep(120 * time.Second)

	lData, err := keystore.ReadLocation(pkg.CLOUD_ID)
	assert.Error(t, err)
	assert.Nil(t, lData)
}
