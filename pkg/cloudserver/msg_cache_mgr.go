/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cloudserver

import (
	"github.com/theotw/natssync/pkg"
	"sync"
	"time"
)

type CachedMsg struct {
	Timestamp time.Time
	ClientID  string
	Data      string
}
type MsgCacheManager interface {
	GetMessages(clientID string) ([]*CachedMsg, error)
	PutMessage(message *CachedMsg) error
	//depth of each queue per client/location ID
	GetQueueDepths() map[string]int
	GetAgeOfOldestTimestamp() time.Duration
	Init() error
}

var cacheMgr MsgCacheManager

func InitCacheMgr() error {
	mgrToUse := pkg.Config.CacheMgr
	var err error
	switch mgrToUse {
	case "mem":
		{
			m := new(InMemMessageCache)
			err = m.Init()
			cacheMgr = m
			break
		}
	case "redis":
		{
			m := new(RedisCacheMgr)
			m.RedisURL = pkg.Config.RedisUrl
			err = m.Init()
			cacheMgr = m
		}
	default:
		{
			m := new(RedisCacheMgr)
			err = m.Init()
			cacheMgr = m
		}
	}
	if err != nil {
		cacheMgr = nil
	}
	return err
}
func GetCacheMgr() MsgCacheManager {
	return cacheMgr
}

type InMemMessageCache struct {
	messages map[string][]*CachedMsg
	mapSync  sync.RWMutex
}

//depth of each queue per client/location ID
func (t *InMemMessageCache) GetQueueDepths() map[string]int {
	ret := make(map[string]int)
	for k, v := range t.messages {
		ret[k] = len(v)
	}
	return ret
}

//oldest message timestamp in the cache
func (t *InMemMessageCache) GetAgeOfOldestTimestamp() time.Duration {
	oldest, start := time.Now(), time.Now()
	for _, v := range t.messages {
		for _, m := range v {
			if m.Timestamp.Before(oldest) {
				oldest = m.Timestamp
			}
		}
	}
	return start.Sub(oldest)
}

func (t *InMemMessageCache) Init() error {
	t.mapSync.Lock()
	t.messages = make(map[string][]*CachedMsg)
	t.mapSync.Unlock()
	return nil
}
func (t *InMemMessageCache) GetMessages(clientID string) ([]*CachedMsg, error) {
	t.mapSync.Lock()
	msgs, ok := t.messages[clientID]

	var ret []*CachedMsg
	if ok {
		ret = msgs
	} else {
		ret = make([]*CachedMsg, 0)
	}
	t.messages[clientID] = make([]*CachedMsg, 0)
	t.mapSync.Unlock()
	return ret, nil
}

func (t *InMemMessageCache) PutMessage(msg *CachedMsg) error {
	t.mapSync.Lock()

	msgs, ok := t.messages[msg.ClientID]
	if !ok {
		msgs = make([]*CachedMsg, 0)
	}
	msgs = append(msgs, msg)
	t.messages[msg.ClientID] = msgs

	t.mapSync.Unlock()
	return nil
}
