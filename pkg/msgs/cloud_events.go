package msgs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/natssync/pkg/bridgemodel"
)

type CloudEventsFormat struct {
	Source 		string		`json:"source"`
	Type		string		`json:"type"`
	SpecVersion	string		`json:"specversion"`
	ID			string		`json:"id"`
	Data		interface{}	`json:"data"`
}

func (f *CloudEventsFormat) GeneratePayload(message string, mType string, source string) ([]byte, error) {
	f.SpecVersion = "1.0"
	f.Type = mType
	f.ID = bridgemodel.GenerateUUID()
	f.Source = source
	f.Data = message

	reqBytes := new(bytes.Buffer)
	err := json.NewEncoder(reqBytes).Encode(f)
	if err != nil {
		return nil, err
	}

	return reqBytes.Bytes(), nil
}

func (f *CloudEventsFormat) ValidateMsgFormat(msg []byte, ceEnabled bool) (bool, error){
	if !ceEnabled {
		log.Info("Cloud Events disabled, skipping message validation")
		return true, nil
	}
	var err error
	err = json.Unmarshal(msg, &f)
	if err != nil {
		log.Errorf("Failed to unmarshal json: %s", err.Error())
		return false, err
	}
	if f.SpecVersion != "1.0" {
		errMsg := fmt.Sprintf("Invalid ID for cloud event, expected 1.0, got %s", f.ID)
		err = errors.New(errMsg)
		return false, err
	}
	if f.Source == "" {
		errMsg := fmt.Sprintf("Source not set for cloud event")
		err = errors.New(errMsg)
		return false, err
	}
	if f.Type == "" {
		errMsg := fmt.Sprintf("Type not set for cloud event")
		err = errors.New(errMsg)
		return false, err
	}
	if f.ID == "" {
		errMsg := fmt.Sprintf("ID not set for cloud event")
		err = errors.New(errMsg)
		return false, err
	}
	log.Info("Successfully validated the cloud events message format")
	return true, nil
}
