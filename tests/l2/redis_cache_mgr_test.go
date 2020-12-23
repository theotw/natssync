/*
 * Copyright (c) The One True Way 2020. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package l2

import (
	"github.com/stretchr/testify/assert"
	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/cloudserver"
	"testing"
)

func TestRedisCacheMgr(t *testing.T) {
	url := pkg.GetEnvWithDefaults("REDISURL", "localhost:6370")
	mgr := cloudserver.RedisCacheMgr{RedisURL: url}
	err := mgr.Init()
	if err != nil {
		t.Fatalf("Got init error %s", err.Error())
	}
	mgr.PutMessage(&cloudserver.CachedMsg{ClientID: "cl1", Data: "hello 1"})
	mgr.PutMessage(&cloudserver.CachedMsg{ClientID: "cl1", Data: "hello 2"})
	mgr.PutMessage(&cloudserver.CachedMsg{ClientID: "cl1", Data: "hello 3"})

	messages, err := mgr.GetMessages("cl1")
	assert.Nil(t, err)
	assert.Greater(t, len(messages), 0)
	messages, err = mgr.GetMessages("cl1")
	assert.Nil(t, err)
	assert.Greater(t, len(messages), 0)
	messages, err = mgr.GetMessages("cl1")
	assert.Nil(t, err)
	assert.Greater(t, len(messages), 0)
	messages, err = mgr.GetMessages("cl1")
	assert.Nil(t, err)
	assert.Equal(t, len(messages), 0)
}
