/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"github.com/mediocregopher/radix/v3"
	log "github.com/sirupsen/logrus"
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

func (t *RedisKeyStore) ReadPrivateKeyData(locationID string) ([]byte, error) {
	log.Tracef("redis Get private key %s", locationID)

	var data string

	err := t.Pool.Do(radix.Cmd(&data, "HGET", PRIVATE_HASH_NAME, locationID))
	var ret []byte
	if err == nil {
		ret = []byte(data)
	}
	return ret, err
}
func (t *RedisKeyStore) ReadPublicKeyData(locationID string) ([]byte, error) {
	log.Tracef("redis Get public key %s", locationID)

	var data string

	err := t.Pool.Do(radix.Cmd(&data, "HGET", PUBLIC_HASH_NAME, locationID))
	var ret []byte
	if err == nil {
		ret = []byte(data)
	}
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
