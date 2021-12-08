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
)

type requestValidator struct {
	configFilePath string
	sync.Mutex
	configMap   map[string]bool
	lastUpdated time.Time
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

	return func() error {
		rv.Mutex.Lock()
		defer rv.Mutex.Unlock()

		if rv.lastUpdated.IsZero() || fileInfo.ModTime().After(rv.lastUpdated) {

			newConfigFileData := make(map[string]bool)
			byteData, err := ioutil.ReadFile(rv.configFilePath)
			if err != nil {
				return fmt.Errorf("failed to read file '%v': %v", rv.configFilePath, err)
			}

			if err = json.Unmarshal(byteData, &newConfigFileData); err != nil {
				return fmt.Errorf("failed to unmarshal data in file '%v': %v", rv.configFilePath, err)
			}

			rv.configMap = newConfigFileData
			rv.lastUpdated = fileInfo.ModTime()
		}

		return nil
	}()
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

	func() {
		rv.Mutex.Lock()
		defer rv.Mutex.Unlock()
		targetValue, targetExists = rv.configMap[targetHost]
	}()

	if !targetExists || !targetValue {
		return fmt.Errorf("invalid target")
	}

	return nil
}
