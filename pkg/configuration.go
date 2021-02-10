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
	Keystore       string
	CertDir        string
	ListenString   string
}

func (c *Configuration) LoadValues() {
	type configOption struct {
		value        *string
		name         string
		defaultValue string
	}

	var configOptions = []configOption{
		{&c.NatsServerUrl, "NATS_SERVER_URL", "nats://127.0.0.1:4222"},
		{&c.CloudBridgeUrl, "CLOUD_BRIDGE_URL", "http://localhost:8081"},
		{&c.LogLevel, "LOG_LEVEL", "debug"},
		{&c.RedisUrl, "REDIS_URL", "localhost:6379"},
		{&c.CacheMgr, "CACHE_MGR", "redis"},
		{&c.Keystore, "KEYSTORE", "redis"},
		{&c.CertDir, "CERT_DIR", "/certs"},
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
