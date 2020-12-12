/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package cloudserver

import "fmt"

import (
	"github.com/mediocregopher/radix/v3"
	log "github.com/sirupsen/logrus"
)

const LIST_PREFIX = "astramsg"

type RedisCacheMgr struct {
	RedisURL string
	Pool     *radix.Pool
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
