/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package pkg

import (
	"os"
	"reflect"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var Config Configuration

type Configuration struct {
	NatsServerUrl  string
	CloudBridgeUrl string
	LogLevel       string
	KeystoreUrl    string
	ListenString   string
	CloudEvents    bool
}

type configOption struct {
	value        interface{}
	name         string
	defaultValue interface{}
}

func (c *Configuration) LoadValues() {
	var configOptions = []configOption{
		{&c.NatsServerUrl, "NATS_SERVER_URL", "nats://127.0.0.1:4222"},
		{&c.CloudBridgeUrl, "CLOUD_BRIDGE_URL", "http://localhost:8081"},
		{&c.LogLevel, "LOG_LEVEL", "debug"},
		{&c.KeystoreUrl, "KEYSTORE_URL", "file:///tmp"},
		{&c.ListenString, "LISTEN_STRING", ":8080"},
		{&c.CloudEvents, "CLOUDEVENTS_ENABLED", false},
	}

	for _, option := range configOptions {
		if reflect.TypeOf(option.defaultValue).Kind() == reflect.Bool {
			*option.value.(*bool) = GetEnvWithDefaultsBool(option.name, option.defaultValue.(bool))
		} else {
			*option.value.(*string) = GetEnvWithDefaults(option.name, option.defaultValue.(string))
		}
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

func GetEnvWithDefaultsBool(envKey string, defaultVal bool) bool {
	val, _ := strconv.ParseBool(os.Getenv(envKey))
	if !val {
		val = defaultVal
	} else {
		log.Debugf("Environment variable %s is set to '%v'", envKey, val)
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
