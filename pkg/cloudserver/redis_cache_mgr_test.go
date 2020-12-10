package cloudserver

import (
	"github.com/stretchr/testify/assert"
	"onprembridge/pkg"
	"testing"
)

func TestRedisCacheMgr(t *testing.T) {
	url := pkg.GetEnvWithDefaults("REDISURL", "localhost:6370")
	mgr := RedisCacheMgr{RedisURL: url}
	err := mgr.Init()
	if err != nil {
		t.Fatalf("Got init error %s", err.Error())
	}
	mgr.PutMessage(&CachedMsg{ClientID: "cl1", Data: "hello 1"})
	mgr.PutMessage(&CachedMsg{ClientID: "cl1", Data: "hello 2"})
	mgr.PutMessage(&CachedMsg{ClientID: "cl1", Data: "hello 3"})

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
