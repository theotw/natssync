/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package pkg

import (
	"os"
	"testing"
	"time"
)

func TestGetEnvWithDefaults(t *testing.T) {
	timeNowString := time.Now().String()
	fooTime := GetEnvWithDefaults("foo_time", timeNowString)

	if fooTime != timeNowString {
		t.Errorf("Expected '%s' but got '%s'", timeNowString, fooTime)
	}

	err := os.Setenv("foo_env", "bar")

	if err != nil {
		t.Errorf("Error setting foo_env: %e", err)
	}
	if GetEnvWithDefaults("foo_env", "baz") != "bar" {
		t.Errorf("Failure to get environment variable when set")
	}
}

func TestNewConfiguration(t *testing.T) {
	config := NewConfiguration()

	type envVarDefault struct {
		value    interface{}
		expected interface{}
	}

	var envVarDefaults = []envVarDefault{
		{config.NatsServerUrl, "nats://127.0.0.1:4222"},
		{config.CloudBridgeUrl, "http://localhost:8081"},
		{config.LogLevel, "debug"},
		{config.KeystoreUrl, "file:///tmp"},
		{config.ListenString, ":8080"},
		{config.CloudEvents, false},
	}

	for _, envVar := range envVarDefaults {
		if envVar.value != envVar.expected {
			t.Errorf("Unexpected value: '%s' != '%s'", envVar.value, envVar.expected)
		}
	}
}
