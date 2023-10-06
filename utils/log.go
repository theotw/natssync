package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	logLevelEnvVariable = "LOG_LEVEL"
	logLevelDefault     = log.InfoLevel
)

func InitLogging() {
	level := logLevelDefault
	logLevel := os.Getenv(logLevelEnvVariable)

	if logLevel != "" {
		newLevel, err := log.ParseLevel(logLevel)
		if err != nil {
			log.Warnf("Could not parse log level \"%s\": %v", err, logLevel)
		} else {
			level = newLevel
		}
	}

	log.Infof("Using log level %s", level)
	log.SetLevel(level)
}
