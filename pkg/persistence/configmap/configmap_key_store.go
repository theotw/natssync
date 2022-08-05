/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package configmap

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/types"
	"github.com/theotw/natssync/pkg/utils"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	locationDataKeyFileSuffix = "_locationData.json"
	serviceKeyFileNameSuffix  = "_serviceKeyData.json"
)

type ConfigmapKeyStore struct {
	cleanupInterval time.Duration
	cleanupTTL      time.Duration

	configmapName      string
	configmapMountPath string
}

func (c *ConfigmapKeyStore) GetExistingKeys() ([]*utils.UUIDv1, error) {
	keys := make([]*utils.UUIDv1, 0)
	existingKeyInfo, err := c.getExistingKeysMetadata()
	if err != nil {
		return nil, err
	}

	for _, key := range existingKeyInfo {
		keys = append(keys, key.uuid)
	}
	return keys, nil
}

func (c *ConfigmapKeyStore) GetLatestKeyID() (string, error) {
	key, err := c.getLatestKeyMetadata()
	if err != nil {
		return "", err
	}
	return key.uuid.String(), nil
}

type KeyMetadata struct {
	mountPath   string
	createdTime int64
	uuid        *utils.UUIDv1
}

func NewKeyMetadata(mountPath string, createdTime int64, uuid *utils.UUIDv1) *KeyMetadata {
	keyMeta := KeyMetadata{
		mountPath:   mountPath,
		createdTime: createdTime,
		uuid:        uuid,
	}
	return &keyMeta
}

func NewConfigmapKeyStore(configmapMountPath string) (*ConfigmapKeyStore, error) {
	ret := new(ConfigmapKeyStore)
	ret.configmapName = pkg.Config.ConfigmapName
	ret.configmapMountPath = configmapMountPath
	return ret, nil
}

func (c *ConfigmapKeyStore) WriteKeyPair(locationData *types.LocationData) error {

	locationDataBytes, err := json.Marshal(locationData)
	if err != nil {
		return err
	}
	serviceKeyFileName := c.makeServiceKeyFileName(locationData.KeyID)
	if err = c.addConfigmapKeyPair(serviceKeyFileName, locationDataBytes); err != nil {
		return err
	}

	return nil
}

func (c *ConfigmapKeyStore) getExistingKeysMetadata() ([]*KeyMetadata, error) {
	existingKeys := make([]*KeyMetadata, 0)
	dir, err := ioutil.ReadDir(c.configmapMountPath)
	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		if strings.HasSuffix(f.Name(), serviceKeyFileNameSuffix) {
			mountPath := fmt.Sprintf("%s/%s", c.configmapMountPath, f.Name())
			split := strings.Split(f.Name(), "_")

			id, err := utils.ParseUUIDv1(split[0])
			if err != nil {
				log.WithError(err).Error("failed to parse keyID from key file name")
			}

			timestamp, err := strconv.ParseInt(split[1], 10, 64)
			if err != nil {
				log.WithError(err).Error("failed to parse keyID timestamp from key file name")
			}
			keyData := NewKeyMetadata(mountPath, timestamp, id)
			existingKeys = append(existingKeys, keyData)
		}
	}
	return existingKeys, nil
}

func (c *ConfigmapKeyStore) getKeyMetadata(keyID string) (*KeyMetadata, error) {
	existingKeys, err := c.getExistingKeysMetadata()
	if err != nil {
		return nil, err
	}
	for _, key := range existingKeys {
		if key.uuid.String() == keyID {
			return key, nil
		}
	}
	log.WithError(err).Error("failed to find key")
	return nil, fmt.Errorf("keyID %s not found", keyID)
}

func (c *ConfigmapKeyStore) getLatestKeyMetadata() (*KeyMetadata, error) {
	existingKeys, err := c.getExistingKeysMetadata()
	if err != nil {
		return nil, err
	}

	if len(existingKeys) == 0 {
		return nil, fmt.Errorf("existing keys not found")
	}

	currentLatestKey := existingKeys[0]
	for _, key := range existingKeys {
		if key.createdTime > currentLatestKey.createdTime {
			currentLatestKey = key
		}
	}
	return currentLatestKey, nil
}

func (c *ConfigmapKeyStore) ReadKeyPair(keyID string) (*types.LocationData, error) {
	var keyMeta *KeyMetadata
	if keyID == "" {
		var err error
		keyMeta, err = c.getLatestKeyMetadata()
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		keyMeta, err = c.getKeyMetadata(keyID)
		if err != nil {
			return nil, err
		}
	}

	locationData := &types.LocationData{}
	locationDataBytes, err := c.readFile(keyMeta.mountPath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(locationDataBytes, locationData); err != nil {
		return nil, err
	}

	return locationData, nil
}

func (c *ConfigmapKeyStore) RemoveKeyPair(keyID string) error {
	var keyMeta *KeyMetadata
	if keyID == "" {
		var err error
		keyMeta, err = c.getLatestKeyMetadata()
		if err != nil {
			return err
		}
	} else {
		var err error
		keyMeta, err = c.getKeyMetadata(keyID)
		if err != nil {
			return err
		}
	}

	_, filename := filepath.Split(keyMeta.mountPath)
	err := c.removeConfigmapKey(filename)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigmapKeyStore) LoadLocationID(keyID string) string {
	log.Infof("LoadLocationID: keyID %s", keyID)
	if keyID == "" {
		keyMeta, err := c.getLatestKeyMetadata()
		if err != nil {
			log.WithError(err).Error("failed get latest keyID")
			return ""
		}
		keyID = keyMeta.uuid.String()
	}

	locationData, err := c.ReadKeyPair(keyID)
	if err != nil {
		log.WithError(err).Error("failed to read key pair")
		return ""
	}
	return locationData.GetLocationID()
}

func (c *ConfigmapKeyStore) WriteLocation(locationData types.LocationData) error {

	locationFile := c.makeLocationDataFileName(locationData.GetLocationID())

	data, err := json.Marshal(locationData)
	if err != nil {
		return err
	}
	return c.addConfigmapKeyPair(locationFile, data)
}

func (c *ConfigmapKeyStore) ReadLocation(locationID string) (*types.LocationData, error) {
	locationFile := fmt.Sprintf("%s/%s", c.configmapMountPath, c.makeLocationDataFileName(locationID))
	log.Infof("Reading location from %s", locationFile)
	locationDataBytes, err := c.readFile(locationFile)
	if err != nil {
		return nil, err
	}

	locationData := &types.LocationData{}
	if err = json.Unmarshal(locationDataBytes, locationData); err != nil {
		return nil, err
	}

	return locationData, nil
}

func (c *ConfigmapKeyStore) removeLocationData(locationID string, allowCloudMaster bool) error {
	var err error

	if locationID == pkg.CLOUD_ID && !allowCloudMaster {
		log.Errorf("Removing default cloud location ID")
		err = fmt.Errorf("unable to remove cloud master location")
		return err
	}

	filename := fmt.Sprintf("%s", c.makeLocationDataFileName(locationID))
	if err = c.removeConfigmapKey(filename); err != nil {
		return err
	}

	return nil
}

func (c *ConfigmapKeyStore) RemoveLocation(locationID string) error {
	return c.removeLocationData(locationID, false)
}

func (c *ConfigmapKeyStore) RemoveCloudMasterData() error {
	return c.removeLocationData(pkg.CLOUD_ID, true)
}

func (c *ConfigmapKeyStore) ListKnownClients() ([]string, error) {
	ret := make([]string, 0)
	dir, err := ioutil.ReadDir(c.configmapMountPath)
	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		if strings.HasSuffix(f.Name(), locationDataKeyFileSuffix) {
			split := strings.Split(f.Name(), "_")
			id := split[0]
			ret = append(ret, id)
		}
	}

	return ret, nil
}

func (c *ConfigmapKeyStore) removeConfigmapKey(key string) error {
	// Note: "/data" is not a file/dir. This is specific to k8s configmaps.
	escapedKeyBytes, err := json.Marshal(fmt.Sprintf("/data/%s", key))
	if err != nil {
		return err
	}

	k8sClient, err := c.getK8sClientset()
	if err != nil {
		return err
	}

	payloadBytes := []byte(fmt.Sprintf("[{\"op\": \"remove\", \"path\": %s}]", string(escapedKeyBytes)))
	_, err = k8sClient.CoreV1().ConfigMaps(pkg.Config.PodNamespace).Patch(context.TODO(), c.configmapName, k8sTypes.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		log.Errorf("Unable to remove configmap key.\n%s", err.Error())
		return err
	}
	return nil
}

func (c *ConfigmapKeyStore) addConfigmapKeyPair(key string, value []byte) error {
	log.Tracef("Updating key '%s' in configmap '%s'", key, c.configmapName)

	k8sClient, err := c.getK8sClientset()
	if err != nil {
		return err
	}

	escapedKeyBytes, err := json.Marshal(key)
	if err != nil {
		return err
	}

	escapedValueBytes, err := json.Marshal(string(value))
	if err != nil {
		return err
	}

	payloadString := fmt.Sprintf("{\"data\": {%s: %s}}", string(escapedKeyBytes), string(escapedValueBytes))
	payloadBytes := []byte(payloadString)

	_, err = k8sClient.CoreV1().ConfigMaps(pkg.Config.PodNamespace).Patch(context.TODO(), c.configmapName, k8sTypes.MergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		log.Errorf("Unable to patch configmap.\n%s", err.Error())
		return err
	}
	return nil
}

func (c *ConfigmapKeyStore) readFile(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	fileData, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return fileData, nil
}

func (c *ConfigmapKeyStore) makeLocationDataFileName(locationID string) string {
	return fmt.Sprintf("%s%s", locationID, locationDataKeyFileSuffix)
}

func (c *ConfigmapKeyStore) getTimestampSuffix() string {
	nanoSecs := time.Now().UnixNano()
	return fmt.Sprintf("_%d", nanoSecs)
}

func (c *ConfigmapKeyStore) makeServiceKeyFileName(keyID string) string {
	return fmt.Sprintf("%s%s%s", keyID, c.getTimestampSuffix(), serviceKeyFileNameSuffix)
}

func (c *ConfigmapKeyStore) getK8sClientset() (*kubernetes.Clientset, error) {
	// Use the k8s service account attached to this pod
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("Unable to initialize Kubernetes client config.\n%s", err.Error())
		return nil, err
	}
	// Create the client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("Unable to initialize Kubernetes client.\n%s", err.Error())
		return nil, err
	}
	return clientset, nil
}
