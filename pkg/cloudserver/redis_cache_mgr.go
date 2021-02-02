/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"fmt"
	"time"
)

import (
	"github.com/mediocregopher/radix/v3"
	log "github.com/sirupsen/logrus"
)

const LIST_PREFIX = "astramsg"

type RedisCacheMgr struct {
	RedisURL string
	Pool     *radix.Pool
}

//depth of each queue per client/location ID
func (t *RedisCacheMgr) GetQueueDepths() map[string]int {
	ret := make(map[string]int)
	var keys []string
	err := t.Pool.Do(radix.Cmd(&keys, "KEYS", LIST_PREFIX+"*"))
	if err != nil {
		log.Errorf("Error getting keys in get queue depths %s \n", err.Error())
		return ret
	}
	for _, key := range keys {
		var data int
		err := t.Pool.Do(radix.Cmd(&data, "LLEN", key))
		if err != nil {
			log.Errorf("Error getting queue len for key %s in get queue depths %s \n", key, err.Error())
			return ret
		} else {
			ret[key] = data
		}
	}
	return ret
}

//oldest message timestamp in the cache
func (t *RedisCacheMgr) GetAgeOfOldestTimestamp() time.Duration {
	var ret time.Duration
	oldest, start := time.Now(), time.Now()
	var keys []string
	err := t.Pool.Do(radix.Cmd(&keys, "KEYS", LIST_PREFIX+"*"))
	if err != nil {
		log.Errorf("Error getting keys for the oldest timestamp %s \n", err.Error())
		return ret
	}

	for _, key := range keys {
		var data []time.Time
		err := t.Pool.Do(radix.Cmd(&data,"HMGET", key, "timestamp"))
		if err != nil {
			log.Errorf("Error getting timestamp for key %s in get oldest timestamp %s \n", key, err.Error())
			return ret
		} else {
			for _, s := range data {
				if s.Before(oldest) {
					oldest = s
				}
			}
			ret = start.Sub(oldest)
		}
	}
    return ret
}

func (t *RedisCacheMgr) GetMessages(clientID string) ([]*CachedMsg, error) {
	log.Tracef("redis Get message %s", clientID)
	listName := mkListName(clientID)
	var data string

	err := t.Pool.Do(radix.Cmd(&data, "LPOP", listName))
	ret := make([]*CachedMsg, 0)
	if err == nil && len(data) > 0 {
		p := new(CachedMsg)
		p.ClientID = clientID
		p.Data = data
		ret = append(ret, p)
	}
	log.Tracef("redis Get message client %s got %d messages", clientID, len(ret))
	if err != nil {
		log.Errorf("Got error fetching messages from redis %s", err.Error())
	}
	return ret, err
}
func (t *RedisCacheMgr) PutMessage(message *CachedMsg) error {
	log.Tracef("redis pushing message %s", message.ClientID)
	listName := mkListName(message.ClientID)
	err := t.Pool.Do(radix.Cmd(nil, "RPUSH", listName, message.Data))
	if err != nil {
		log.Errorf("Got error putting messages from redis %s", err.Error())
	}
	return err
}
func (t *RedisCacheMgr) Init() error {
	var err error
	addrs := make([]string, 1)
	addrs[0] = t.RedisURL
	t.Pool, err = radix.NewPool("tcp", t.RedisURL, 10)
	if err != nil {
		log.Errorf("Unable to connect to redis")
	}
	return err
}
func mkListName(clientID string) string {
	return fmt.Sprintf("%s.%s", LIST_PREFIX, clientID)
}
