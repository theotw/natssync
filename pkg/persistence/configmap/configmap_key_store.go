/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package configmap

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/types"
	"github.com/theotw/natssync/pkg/utils"
	k8sTypes "k8s.io/apimachinery/pkg/types"
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
	existingKeyInfo, err := c.getExistingKeyMetadata()
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

func (c *ConfigmapKeyStore) getExistingKeyMetadata() ([]*KeyMetadata, error) {
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
			existingKeys = append(existingKeys, NewKeyMetadata(mountPath, timestamp, id))
		}
	}
	return existingKeys, nil
}

func (c *ConfigmapKeyStore) getLatestKeyMetadata() (*KeyMetadata, error) {
	existingKeys, err := c.getExistingKeyMetadata()
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
	if keyID == "" {
		var err error
		keyMeta, err := c.getLatestKeyMetadata()
		if err != nil {
			return nil, err
		}
		keyID = keyMeta.uuid.String()
	}

	locationData := &types.LocationData{}
	serviceKeyFileName := fmt.Sprintf("%s/%s", c.configmapMountPath, c.makeServiceKeyFileName(keyID))
	locationDataBytes, err := c.readFile(serviceKeyFileName)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(locationDataBytes, locationData); err != nil {
		return nil, err
	}

	return locationData, nil
}

func (c *ConfigmapKeyStore) RemoveKeyPair(keyID string) error {
	if keyID == "" {
		keyMeta, err := c.getLatestKeyMetadata()
		if err != nil {
			return err
		}
		keyID = keyMeta.uuid.String()
	}

	serviceKeyFileName := c.makeServiceKeyFileName(keyID)
	err := c.removeConfigmapKey(serviceKeyFileName)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigmapKeyStore) LoadLocationID(keyID string) string {
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

	filename := c.makeLocationDataFileName(locationID)
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
	dir, err := ioutil.ReadDir(c.configmapName)
	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		if strings.HasSuffix(f.Name(), locationDataKeyFileSuffix) {
			split := strings.Split(f.Name(), "_")
			//id := split[0]
			id, err := utils.ParseUUIDv1(split[0])
			if err != nil {
				log.WithError(err).Error("failed to parse keyID from key file name")
			}
			ret = append(ret, id.String())
		}
	}

	return ret, nil
}

func (c *ConfigmapKeyStore) removeConfigmapKey(key string) error {
	k8sClient, err := c.getK8sClientset()
	if err != nil {
		log.Errorf("Unable to initialize kubernetes client.\n%s", err.Error()) //todo duplicated
		return err
	}

	payloadBytes := []byte(fmt.Sprintf("[{\"op\": \"remove\", \"path\": \"/data/%s\"}]", key))
	_, err = k8sClient.CoreV1().ConfigMaps("").Patch(context.TODO(), c.configmapName, k8sTypes.JSONPatchType, payloadBytes, metav1.PatchOptions{})
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
		log.Errorf("Unable to initialize kubernetes client.\n%s", err.Error()) //todo duplicated
		return err
	}

	payloadBytes := []byte(fmt.Sprintf("{\"data\":{\"%s\": \"%s\"}}", key, string(value)))
	_, err = k8sClient.CoreV1().ConfigMaps("").Patch(context.TODO(), c.configmapName, k8sTypes.MergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		log.Errorf("Unable to patch configmap.\n%s", err.Error())
		return err
	}
	return nil
}

func (c *ConfigmapKeyStore) readFile(fileName string) ([]byte, error) {
	pathToFile := path.Join(c.configmapMountPath, fileName)
	log.Tracef("Reading from file %s", pathToFile)
	f, err := os.Open(pathToFile)
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
	secs := time.Now().Unix()
	return fmt.Sprintf("_%d", secs)
}

func (c *ConfigmapKeyStore) makeServiceKeyFileName(keyID string) string {
	return fmt.Sprintf("%s%s%s", keyID, serviceKeyFileNameSuffix, c.getTimestampSuffix())
}

func (c *ConfigmapKeyStore) getK8sClientset() (*kubernetes.Clientset, error) {
	// Use the k8s service account attached to this pod
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// Create the client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
