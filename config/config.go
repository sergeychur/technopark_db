package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port   string `json:"port"`
	DBHost string `json:"dbhost"`
	DBPort string `json:"dbport"`
	DBUser string `json:"dbuser"`
	DBPass string `json:"dbpassword"`
	DBName string `json:"dbname"`
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
