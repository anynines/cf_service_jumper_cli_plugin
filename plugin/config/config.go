package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/configuration/confighelpers"
)

var (
	ErrForwardConfigMissing = errors.New("[ERR] forward.json not found")
	ErrTargetBlank          = errors.New("[ERR] target not present")
)

type ForwardConfig struct {
	Target string `json:"target"`
}

func newForwardConfig() ForwardConfig {
	return ForwardConfig{}
}

func DefaultFilePath() (string, error) {
	defaultFilePath, err := confighelpers.DefaultFilePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(defaultFilePath), "forward.json"), nil
}

func GetConfig() (ForwardConfig, error) {
	var forwardConfig ForwardConfig

	defaultFilePath, err := DefaultFilePath()
	if err != nil {
		return forwardConfig, err
	}

	jsonBytes, err := ioutil.ReadFile(defaultFilePath)
	if err != nil {
		return forwardConfig, ErrForwardConfigMissing
	}

	err = json.Unmarshal(jsonBytes, &forwardConfig)
	return forwardConfig, err
}

func SetTarget(target string) error {
	defaultFilePath, err := DefaultFilePath()
	if err != nil {
		return err
	}

	config, err := GetConfig()
	if err != nil && err != ErrForwardConfigMissing {
		return err
	}
	if err == ErrForwardConfigMissing {
		config = newForwardConfig()
	}

	config.Target = target

	data, err := json.Marshal(&config)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(defaultFilePath, data, 0700)
}
