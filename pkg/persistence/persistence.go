package persistence

import (
	"context"
	"fmt"
	"github.com/theotw/natssync/pkg/persistence/configmap"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/persistence/file"
	"github.com/theotw/natssync/pkg/persistence/mongo"
	types "github.com/theotw/natssync/pkg/types"
)

const (
	fileKeyStoreTypePrefix      = "file://"
	mongoKeyStoreTypePrefix     = "mongodb://"
	configmapKeyStoreTypePrefix = "configmap://"
)

type LocationKeyStore interface {
	// ReadKeyPair if keyID is empty read the latest key
	ReadKeyPair(KeyID string) (*types.LocationData, error)
	WriteKeyPair(locationData *types.LocationData) error
	// RemoveKeyPair if keyID is empty remove the latest key
	RemoveKeyPair(KeyID string) error
	// LoadLocationID if keyID is empty load the location from latest key
	LoadLocationID(KeyID string) string
	WriteLocation(locationData types.LocationData) error
	ReadLocation(locationID string) (*types.LocationData, error)
	RemoveLocation(locationID string) error
	RemoveCloudMasterData() error
	ListKnownClients() ([]string, error)
}

var keystore LocationKeyStore

func GetKeyStore() LocationKeyStore {
	return keystore
}

func parseKeystoreUrl(keystoreUrl string) (string, string, error) {
	log.Debugf("Parsing keystore URL: %s", keystoreUrl)
	ksTypeUrl := strings.SplitAfterN(keystoreUrl, "://", 2)
	if len(ksTypeUrl) != 2 {
		return "", "", fmt.Errorf("unable to parse url '%s'", keystoreUrl)
	}
	ksType := ksTypeUrl[0]
	ksUrl := ksTypeUrl[1]
	return ksType, ksUrl, nil
}

func CreateLocationKeyStore(keystoreUrl string) (LocationKeyStore, error) {
	keystoreType, keystoreUri, err := parseKeystoreUrl(keystoreUrl)
	if err != nil {
		log.Fatal(err)
	}

	switch keystoreType {
	case fileKeyStoreTypePrefix:
		fileKeyStore, err := file.NewFileKeyStore(keystoreUri)
		if err == nil {
			newReaper(fileKeyStore).RunCleanupJob(context.TODO())
		}
		return fileKeyStore, err

	case mongoKeyStoreTypePrefix:
		mongoKeyStore, err := mongo.NewMongoKeyStore(keystoreUri)
		if err == nil {
			newReaper(mongoKeyStore).RunCleanupJob(context.TODO())
		}
		return mongoKeyStore, err

	case configmapKeyStoreTypePrefix:
		configmapKeyStore, err := configmap.NewConfigmapKeyStore(keystoreUri)
		if err == nil {
			newReaper(configmapKeyStore).RunCleanupJob(context.TODO())
		}
		return configmapKeyStore, err
	}

	return nil, fmt.Errorf("unsupported keystore types %s", keystoreType)
}

func InitLocationKeyStore() error {
	var err error
	keystore, err = CreateLocationKeyStore(pkg.Config.KeystoreUrl)
	return err
}
