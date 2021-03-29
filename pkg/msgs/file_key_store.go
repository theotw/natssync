/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

const publicKeySuffix = "_public.pem"
const metadataFileSuffix = ".meta"

const privateKeyFileName = "private.pem"
const publicKeyFileName = "public.pem"
const locationIDFileName = "locationID.txt"

type FileKeyStore struct {
	basePath string
}

func NewFileKeyStore(basePath string) (*FileKeyStore, error) {
	ret := new(FileKeyStore)
	ret.basePath = basePath
	return ret, nil
}

func (t *FileKeyStore) WriteKeyPair(locationID string, publicKey []byte, privateKey []byte) error {
	var err error
	if err = t.writeFile(privateKeyFileName, privateKey); err != nil {
		return err
	}
	if err = t.writeFile(publicKeyFileName, publicKey); err != nil {
		return err
	}
	if err = t.writeFile(locationIDFileName, []byte(locationID)); err != nil {
		return err
	}
	return nil
}

func (t *FileKeyStore) ReadKeyPair() ([]byte, []byte, error) {
	publicKey, err := t.readFile(publicKeyFileName)
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := t.readFile(privateKeyFileName)
	if err != nil {
		return nil, nil, err
	}

	return publicKey, privateKey, nil
}

func (t *FileKeyStore) RemoveKeyPair() error {
	var err error
	if err = t.removeFile(privateKeyFileName); err != nil {
		return err
	}
	if err = t.removeFile(publicKeyFileName); err != nil {
		return err
	}
	if err = t.removeFile(locationIDFileName); err != nil {
		return err
	}
	return nil
}

func (t *FileKeyStore) GetLocationID() string {
	locationID, err := t.readFile(locationIDFileName)
	if err != nil {
		return ""
	}
	return string(locationID)
}

func (t *FileKeyStore) WriteLocation(locationID string, buf []byte, metadata map[string]string) error {
	var err error
	locationFile := t.makePublicKeyFileName(locationID)
	if err = t.writeFile(locationFile, buf); err != nil {
		return err
	}
	metaFile := t.makeMetaDataFileName(locationID)

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	if err = t.writeFile(metaFile, metadataBytes); err != nil {
		return err
	}
	return nil
}

func (t *FileKeyStore) ReadLocation(locationID string) ([]byte, map[string]string, error) {
	locationFile := t.makePublicKeyFileName(locationID)
	publicKey, err := t.readFile(locationFile)
	if err != nil {
		return nil, nil, err
	}

	metaFile := t.makeMetaDataFileName(locationID)
	metadataBytes, err := t.readFile(metaFile)
	if err != nil {
		return nil, nil, err
	}

	metadata := make(map[string]string)
	if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, nil, err
	}

	return publicKey, metadata, nil
}

func (t *FileKeyStore) removeLocationData(locationID string, allowCloudMaster bool) error {
	var err error

	if locationID == CLOUD_ID && !allowCloudMaster {
		log.Errorf("Removing default cloud location ID")
		err = fmt.Errorf("unable to remove cloud master location")
		return err
	}

	filename := t.makePublicKeyFileName(locationID)
	if err = t.removeFile(filename); err != nil {
		return err
	}

	metafile := t.makeMetaDataFileName(locationID)
	if err = t.removeFile(metafile); err != nil {
		return err
	}

	return nil
}

func (t *FileKeyStore) RemoveLocation(locationID string) error {
	return t.removeLocationData(locationID, false)
}

func (t *FileKeyStore) RemoveCloudMasterData() error {
	return t.removeLocationData(CLOUD_ID, true)
}

func (t *FileKeyStore) ListKnownClients() ([]string, error) {
	ret := make([]string, 0)
	dir, err := ioutil.ReadDir(t.basePath)
	if err != nil {
		return nil, err
	}

	for _, f := range dir {
		if strings.HasSuffix(f.Name(), publicKeySuffix) {
			limit := len(f.Name()) - len(publicKeySuffix)
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

func (t *FileKeyStore) makePublicKeyFileName(locationID string) string {
	return fmt.Sprintf("%s%s", locationID, publicKeySuffix)
}

func (t *FileKeyStore) makeMetaDataFileName(locationID string) string {
	return fmt.Sprintf("%s%s", locationID, metadataFileSuffix)
}
