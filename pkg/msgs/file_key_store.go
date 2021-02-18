/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const publicKeySuffix = "_public.pem"

type FileKeyStore struct {
	basePath string
}

func NewFileKeyStore(conn *nats.Conn) (*FileKeyStore, error) {
	ret := new(FileKeyStore)
	ret.basePath = pkg.Config.CertDir
	if conn != nil {
		ret.registerForLifecycleEvents(conn)
	}
	return ret, nil
}
func (t *FileKeyStore) registerForLifecycleEvents(conn *nats.Conn) {
	conn.Subscribe(bridgemodel.REGISTRATION_LIFECYCLE_REMOVED, func(msg *nats.Msg) {
		if msg.Data != nil {
			locationID := string(msg.Data)
			log.Infof("Recieved Lifecycle removed event %s", locationID)
			name := t.makePublicKeyFileName(locationID)
			err := os.Remove(name)
			if err != nil {
				log.Errorf("Recieved lifecycle removed but not an error %s", err.Error())
			}
		} else {
			log.Errorf("Recieved lifecycle removed with no data")
		}
	})
}
func (t *FileKeyStore) LoadLocationID() string {
	var ret string
	fileName := path.Join(t.basePath, "locationkey.txt")
	f, err := os.Open(fileName)
	if err == nil {
		defer f.Close()
		all, err := ioutil.ReadAll(f)
		if err == nil {
			ret = string(all)
		}
	}
	return ret
}
func (t *FileKeyStore) SaveLocationID(locationID string) error {
	fileName := path.Join(t.basePath, "locationkey.txt")
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(locationID))
	return err
}

func (t *FileKeyStore) WritePublicKey(locationID string, buf []byte) error {
	fileName := t.makePublicKeyFileName(locationID)
	return t.writeKeyFile(fileName, buf)
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
	keyFile = fmt.Sprintf("%s%s", locationID, publicKeySuffix)

	masterPemPath := path.Join(t.basePath, keyFile)
	return masterPemPath
}
func (t *FileKeyStore) makePrivateFileName(locationID string) string {
	var keyFile string
	keyFile = fmt.Sprintf("%s.pem", locationID)

	masterPemPath := path.Join(t.basePath, keyFile)
	return masterPemPath
}
