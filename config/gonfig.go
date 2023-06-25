package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	DBName     string `json:"dbname"`
	SchemaName string `json:"schemaName"`
	TableName  string `json:"tableName"`
}

func GetConfig() (config Config, err error) {
	configData, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
