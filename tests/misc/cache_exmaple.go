/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"github.com/nats-io/nats.go"
	"github.com/theotw/natssync/pkg/bridgemodel"
	"sync"
	"time"
)

func main() {
	bridgemodel.InitNats("nats://nats:4200","example",50*time.Second)
	nc:=bridgemodel.GetNatsConnection()
	cache:=NewCache()
	nc.Subscribe("cachedirty", func(msg *nats.Msg) {
		cache.LoadCache()
	})
	cache.LoadCache()


}

type Cache struct {
	data map[string]string
	dataLock sync.RWMutex
}
func NewCache() *Cache{
	ret:=new (Cache)
	return ret
}
func (t *Cache) getValue(key string) string{
	var ret string
	t.dataLock.RLock()
	ret=t.data[key]
	t.dataLock.RUnlock()
	return ret
}
func (t *Cache) LoadCache(){
	t.dataLock.Lock()
	t.data = fetchCacheData()
	t.dataLock.Unlock()
}

func fetchCacheData()map[string]string{
	//do something interesting
	ret:=make(map[string]string)
	return ret
}








