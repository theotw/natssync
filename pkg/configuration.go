/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package pkg

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var Config Configuration

type Configuration struct {
	NatsServerUrl  string
	CloudBridgeUrl string
	LogLevel       string
	RedisUrl       string
	CacheMgr       string
	KeystoreUrl    string
	ListenString   string
}

type configOption struct {
	value        *string
	name         string
	defaultValue string
}

func (c *Configuration) LoadValues() {
	var configOptions = []configOption{
		{&c.NatsServerUrl, "NATS_SERVER_URL", "nats://127.0.0.1:4222"},
		{&c.CloudBridgeUrl, "CLOUD_BRIDGE_URL", "http://localhost:8081"},
		{&c.LogLevel, "LOG_LEVEL", "debug"},
		{&c.RedisUrl, "REDIS_URL", "localhost:6379"}, // TODO: Remove in favor of CacheMgrUrl
		{&c.CacheMgr, "CACHE_MGR", "redis"}, // TODO: Convert to CacheMgrUrl
		{&c.KeystoreUrl, "KEYSTORE_URL", "redis://localhost:6379"},
		{&c.ListenString, "LISTEN_STRING", ":8080"},
	}

	for _, option := range configOptions {
		*option.value = GetEnvWithDefaults(option.name, option.defaultValue)
	}
}

func GetEnvWithDefaults(envKey string, defaultVal string) string {
	val := os.Getenv(envKey)
	if len(val) == 0 {
		val = defaultVal
	} else {
		log.Debugf("Environment variable %s is set to '%s'", envKey, val)
	}
	return val
}

func NewConfiguration() Configuration {
	config := Configuration{}
	config.LoadValues()
	return config
}

func init() {
	Config = NewConfiguration()
}
