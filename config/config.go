package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port         string   `json:"port"`
}

func NewConfig(pathToConfig string) (*Config, error) {
	conf := new(Config)
	configFile, err := os.Open(pathToConfig)
	if err != nil {
		return nil, err
	}

	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}