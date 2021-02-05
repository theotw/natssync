/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"io/ioutil"
	"os"
	"path"
)

type FileKeyStore struct {
	basePath string
}

func NewFileKeyStore() (*FileKeyStore, error) {
	ret := new(FileKeyStore)
	ret.basePath = pkg.Config.CertDir
	return ret, nil
}

func (t *FileKeyStore) WritePublicKey(locationID string, buf []byte) error {
	fileName := t.makePublicKeyFileName(locationID)
	return t.writeKeyFile(fileName, buf)
}
func (t *FileKeyStore) WritePrivateKey(locationID string, buf []byte) error {
	fileName := t.makePrivateFileName(locationID)
	return t.writeKeyFile(fileName, buf)
}

func (t *FileKeyStore) writeKeyFile(fileName string, buf []byte) error {
	keyFile, err := os.Create(fileName)
	if err != nil {
		log.Errorf("Unable to open key file %s \n", err.Error())
		return err
	}
	defer keyFile.Close()
	keyFile.Write(buf)
	return nil
}

func (t *FileKeyStore) ReadPrivateKeyData(locationID string) ([]byte, error) {
	keyPath := t.makePrivateFileName(locationID)
	f, err := os.Open(keyPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	all, err := ioutil.ReadAll(f)
	return all, err
}
func (t *FileKeyStore) ReadPublicKeyData(locationID string) ([]byte, error) {
	keyPath := t.makePublicKeyFileName(locationID)
	f, err := os.Open(keyPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	all, err := ioutil.ReadAll(f)
	return all, err
}

func (t *FileKeyStore) makePublicKeyFileName(locationID string) string {
	var keyFile string
	keyFile = fmt.Sprintf("%s_public.pem", locationID)

	masterPemPath := path.Join(t.basePath, keyFile)
	return masterPemPath
}
func (t *FileKeyStore) makePrivateFileName(locationID string) string {
	var keyFile string
	keyFile = fmt.Sprintf("%s.pem", locationID)

	masterPemPath := path.Join(t.basePath, keyFile)
	return masterPemPath
}
