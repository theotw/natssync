package utils

import (
	log "github.com/sirupsen/logrus"

	httpproxy "github.com/theotw/natssync/pkg/httpsproxy"
)

const (
	logLevelEnvVariable = "LOG_LEVEL"
)

func InitLogging() {
	logLevel := httpproxy.GetEnvWithDefaults(logLevelEnvVariable, log.DebugLevel.String())
	level, levelerr := log.ParseLevel(logLevel)
	if levelerr != nil {
		log.Infof("No valid log level from ENV, defaulting to debug level was: %s", level)
		level = log.DebugLevel
	}
	log.Infof("Using log level 2 %s / %d", logLevel, level)
	log.SetLevel(level)
}
