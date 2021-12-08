package proxylet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultConfigFilePath = "/etc/proxylet/validHostsConfig.json"
	defaultBehaviorBlock  = "block"
	defaultBehaviorAllow  = "allow"
)

type configFileData struct {
	Default string          `json:"defaultBehaviour"`
	Hosts   map[string]bool `json:"hosts"`
}

type requestValidator struct {
	sync.RWMutex
	configFilePath string
	configMap      configFileData
	lastUpdated    time.Time
}

func NewRequestValidator(configFilePaths ...string) (*requestValidator, error) {
	configFilePath := defaultConfigFilePath
	if len(configFilePaths) > 0 {
		configFilePath = configFilePaths[0]
	}

	rv := &requestValidator{
		configFilePath: configFilePath,
	}

	rv.refreshConfigIfRequired()

	return rv, nil
}

func (rv *requestValidator) refreshConfigIfRequired() {
	if err := rv.swapMapIfRequired(); err != nil {
		log.WithError(err).Warning("unable to find config file, defaulting to block all requests")
	}
}

func (rv *requestValidator) swapMapIfRequired() error {

	fileInfo, err := os.Stat(rv.configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("failed to to find config file: %v", rv.configFilePath)
			return fmt.Errorf("unable to list file '%v': %v", rv.configFilePath, err)
		}

		return fmt.Errorf("listing file failed with error: %v", err)
	}

	if rv.lastUpdated.IsZero() || fileInfo.ModTime().After(rv.lastUpdated) {

		newConfigFileData := configFileData{}
		byteData, err := ioutil.ReadFile(rv.configFilePath)
		if err != nil {
			return fmt.Errorf("failed to read file '%v': %v", rv.configFilePath, err)
		}

		if err = json.Unmarshal(byteData, &newConfigFileData); err != nil {
			return fmt.Errorf("failed to unmarshal data in file '%v': %v", rv.configFilePath, err)
		}

		func() {
			rv.RWMutex.Lock()
			defer rv.RWMutex.Unlock()

			rv.configMap = newConfigFileData
			rv.lastUpdated = fileInfo.ModTime()
		}()

	}

	return nil

}

func getTargetHost(target string) string {
	output := strings.Split(target, ":")
	if len(output) > 0 {
		return output[0]
	}
	return target
}

func (rv *requestValidator) IsValidRequest(target string) error {

	targetHost := getTargetHost(target)

	rv.refreshConfigIfRequired()

	targetExists := false
	targetValue := false
	defaultBehavior := defaultBehaviorBlock

	func() {
		rv.RWMutex.RLock()
		defer rv.RWMutex.RUnlock()

		targetValue, targetExists = rv.configMap.Hosts[targetHost]

		defaultBehaviourTmp := rv.configMap.Default
		if defaultBehaviourTmp != "" {
			defaultBehavior = defaultBehaviourTmp
		}

	}()

	if (!targetExists || !targetValue) && defaultBehavior != defaultBehaviorAllow {

		return fmt.Errorf("invalid target")
	}

	return nil
}
