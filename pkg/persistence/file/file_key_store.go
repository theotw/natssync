/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/types"
	"github.com/theotw/natssync/pkg/utils"
)

const (
	locationDataKeyFileSuffix = "_locationData.json"
	serviceKeyFileNameSuffix  = "_serviceKeyData.json"
)

type FileKeyStore struct {
	cleanupInterval time.Duration
	cleanupTTL      time.Duration

	basePath string
}

func NewFileKeyStore(basePath string) (*FileKeyStore, error) {
	ret := new(FileKeyStore)
	ret.basePath = basePath
	return ret, nil
}

func (t *FileKeyStore) WriteKeyPair(locationData *types.LocationData) error {

	locationDataBytes, err := json.Marshal(locationData)
	if err != nil {
		return err
	}
	serviceKeyFileName := t.makeServiceKeyFileName(locationData.KeyID)
	if err = t.writeFile(serviceKeyFileName, locationDataBytes); err != nil {
		return err
	}

	return nil
}

func (t *FileKeyStore) GetExistingKeys() ([]*utils.UUIDv1, error) {
	existingKeys := make([]*utils.UUIDv1, 0)
	dir, err := ioutil.ReadDir(t.basePath)
	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		if strings.HasSuffix(f.Name(), serviceKeyFileNameSuffix) {
			limit := len(f.Name()) - len(serviceKeyFileNameSuffix)
			idString := f.Name()[:limit]
			id, err := utils.ParseUUIDv1(idString)
			if err != nil {
				log.WithError(err).Error("failed to parse keyID from key file name")
			}
			existingKeys = append(existingKeys, id)
		}
	}
	return existingKeys, nil
}

func (t *FileKeyStore) GetLatestKeyID() (string, error) {
	existingKeys, err := t.GetExistingKeys()
	if err != nil {
		return "", err
	}

	if len(existingKeys) == 0 {
		return "", fmt.Errorf("existing keys not found")
	}

	maxID := existingKeys[0]
	for _, key := range existingKeys {
		if maxID.GetCreationTime().Before(key.GetCreationTime()) {
			maxID = key
		}
	}
	return maxID.String(), nil
}

func (t *FileKeyStore) ReadKeyPair(keyID string) (*types.LocationData, error) {
	log.Infof("FileKeyStore ReadKeyPair: keyID %s", keyID)
	if keyID == "" {
		var err error
		if keyID, err = t.GetLatestKeyID(); err != nil {
			return nil, err
		}
	}

	locationData := &types.LocationData{}
	serviceKeyFileName := t.makeServiceKeyFileName(keyID)
	locationDataBytes, err := t.readFile(serviceKeyFileName)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(locationDataBytes, locationData); err != nil {
		return nil, err
	}

	return locationData, nil
}

func (t *FileKeyStore) RemoveKeyPair(keyID string) error {
	if keyID == "" {
		var err error
		if keyID, err = t.GetLatestKeyID(); err != nil {
			return err
		}
	}

	var err error
	serviceKeyFileName := t.makeServiceKeyFileName(keyID)
	if err = t.removeFile(serviceKeyFileName); err != nil {
		return err
	}

	return nil
}

func (t *FileKeyStore) LoadLocationID(keyID string) string {

	if keyID == "" {
		var err error
		if keyID, err = t.GetLatestKeyID(); err != nil {
			log.WithError(err).Error("failed get latest keyID")
			return ""
		}
	}

	locationData, err := t.ReadKeyPair(keyID)
	if err != nil {
		log.WithError(err).Error("failed to read key pair")
		return ""
	}
	return locationData.GetLocationID()
}

func (t *FileKeyStore) WriteLocation(locationData types.LocationData) error {

	locationFile := t.makeLocationDataFileName(locationData.GetLocationID())

	data, err := json.Marshal(locationData)
	if err != nil {
		return err
	}
	return t.writeFile(locationFile, data)
}

func (t *FileKeyStore) ReadLocation(locationID string) (*types.LocationData, error) {
	locationFile := t.makeLocationDataFileName(locationID)
	locationDataBytes, err := t.readFile(locationFile)
	if err != nil {
		return nil, err
	}

	locationData := &types.LocationData{}
	if err = json.Unmarshal(locationDataBytes, locationData); err != nil {
		return nil, err
	}

	return locationData, nil
}

func (t *FileKeyStore) removeLocationData(locationID string, allowCloudMaster bool) error {
	var err error

	if locationID == pkg.CLOUD_ID && !allowCloudMaster {
		log.Errorf("Removing default cloud location ID")
		err = fmt.Errorf("unable to remove cloud master location")
		return err
	}

	filename := t.makeLocationDataFileName(locationID)
	if err = t.removeFile(filename); err != nil {
		return err
	}

	return nil
}

func (t *FileKeyStore) RemoveLocation(locationID string) error {
	return t.removeLocationData(locationID, false)
}

func (t *FileKeyStore) RemoveCloudMasterData() error {
	return t.removeLocationData(pkg.CLOUD_ID, true)
}

func (t *FileKeyStore) ListKnownClients() ([]string, error) {
	ret := make([]string, 0)
	dir, err := ioutil.ReadDir(t.basePath)
	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		if strings.HasSuffix(f.Name(), locationDataKeyFileSuffix) {
			limit := len(f.Name()) - len(locationDataKeyFileSuffix)
			id := f.Name()[:limit]
			ret = append(ret, id)
		}
	}

	return ret, nil
}

func (t *FileKeyStore) writeFile(fileName string, buf []byte) error {
	pathToFile := path.Join(t.basePath, fileName)
	log.Tracef("Writing to file %s", pathToFile)
	keyFile, err := os.Create(pathToFile)
	if err != nil {
		log.Errorf("Unable to open key file %s \n", err.Error())
		return err
	}

	defer keyFile.Close()
	_, err = keyFile.Write(buf)
	if err != nil {
		log.Errorf("Unable to write key file %s \n", err.Error())
		return err
	}

	return nil
}

func (t *FileKeyStore) readFile(fileName string) ([]byte, error) {
	pathToFile := path.Join(t.basePath, fileName)
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

func (t *FileKeyStore) removeFile(fileName string) error {
	pathToFile := path.Join(t.basePath, fileName)
	log.Tracef("Removing file %s", pathToFile)
	return os.Remove(pathToFile)
}

func (t *FileKeyStore) makeLocationDataFileName(locationID string) string {
	return fmt.Sprintf("%s%s", locationID, locationDataKeyFileSuffix)
}

func (t *FileKeyStore) makeServiceKeyFileName(keyID string) string {
	return fmt.Sprintf("%s%s", keyID, serviceKeyFileNameSuffix)
}
