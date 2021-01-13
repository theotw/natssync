/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"errors"
	"github.com/mediocregopher/radix/v3"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg"
)

//stores keys in a redis storage.
//It uses 2 hash sets,natssync privte and public keystore
//Each hashset key is the location ID, the value is the data for the pub or private key
type RedisKeyStore struct {
	RedisURL string
	Pool     *radix.Pool
}

const PRIVATE_HASH_NAME = "natssync_private_key_store"
const PUBLIC_HASH_NAME = "natssync_public_key_store"

func NewRedisLocationKeyStore() (*RedisKeyStore, error) {
	ret := new(RedisKeyStore)
	ret.RedisURL = pkg.GetEnvWithDefaults("REDIS_URL", "localhost:6379")
	err := ret.Init()
	return ret, err
}
func (t *RedisKeyStore) Init() error {
	var err error
	addrs := make([]string, 1)
	addrs[0] = t.RedisURL
	t.Pool, err = radix.NewPool("tcp", t.RedisURL, 10)
	if err != nil {
		log.Errorf("Unable to connect to redis")
	}
	return err
}

func (t *RedisKeyStore) ReadPrivateKeyData(locationID string) ([]byte, error) {
	log.Tracef("redis Get private key %s", locationID)

	var data string

	err := t.Pool.Do(radix.Cmd(&data, "HGET", PRIVATE_HASH_NAME, locationID))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("no key found")
	}
	var ret []byte
	ret = []byte(data)
	return ret, err
}
func (t *RedisKeyStore) ReadPublicKeyData(locationID string) ([]byte, error) {
	log.Tracef("redis Get public key %s", locationID)

	var data string

	err := t.Pool.Do(radix.Cmd(&data, "HGET", PUBLIC_HASH_NAME, locationID))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("no key found")
	}
	var ret []byte
	ret = []byte(data)
	return ret, err
}
func (t *RedisKeyStore) WritePrivateKey(locationID string, buf []byte) error {
	log.Tracef("redis write private key %s", locationID)
	data := string(buf)
	err := t.Pool.Do(radix.Cmd(nil, "HSET", PRIVATE_HASH_NAME, locationID, data))
	return err
}
func (t *RedisKeyStore) WritePublicKey(locationID string, buf []byte) error {
	log.Tracef("redis write public key %s", locationID)
	data := string(buf)
	err := t.Pool.Do(radix.Cmd(nil, "HSET", PUBLIC_HASH_NAME, locationID, data))
	return err
}