/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package pkg

import (
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	CLOUD_ID               = "cloud-master"
	StatusCertificateError = 495
)

var Config Configuration

type Configuration struct {
	NatsServerUrl     string
	CloudBridgeUrl    string
	LogLevel          string
	KeystoreUrl       string
	MongodbServer     string
	MongodbPort       string
	MongodbUsername   string
	MongodbPassword   string
	ListenString      string
	ConfigmapName     string
	PodNamespace      string
	CloudEvents       bool
	SkipTlsValidation bool
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
		{&c.MongodbServer, "MONGODB_SERVER", ""},
		{&c.MongodbPort, "MONGODB_PORT", "27017"},
		{&c.MongodbUsername, "MONGODB_USERNAME", ""},
		{&c.MongodbPassword, "MONGODB_PASSWORD", ""},
		{&c.ListenString, "LISTEN_STRING", ":8080"},
		{&c.ConfigmapName, "CONFIGMAP_NAME", ""},
		{&c.CloudEvents, "CLOUDEVENTS_ENABLED", false},
		{&c.SkipTlsValidation, "SKIP_TLS_VALIDATION", false},
	}

	for _, option := range configOptions {
		if reflect.TypeOf(option.defaultValue).Kind() == reflect.Bool {
			*option.value.(*bool) = GetEnvWithDefaultsBool(option.name, option.defaultValue.(bool))
		} else {
			*option.value.(*string) = GetEnvWithDefaults(option.name, option.defaultValue.(string))
		}
	}

	c.PodNamespace = PodNamespace()
}

func PodNamespace() string {
	// See https://github.com/kubernetes/kubernetes/pull/63707. This is currently the best way to get the k8s namespace.
	// This way assumes you've set the POD_NAMESPACE environment variable using the downward API.
	// This check has to be done first for backwards compatibility with the way InClusterConfig was originally set up
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		return ns
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}

	return "default"
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
