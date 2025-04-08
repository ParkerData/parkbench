package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	GRPCServerAddress string `json:"grpcAddress"`
	HTTPServerAddress string `json:"httpAddress"`
	CSVFilePath       string `json:"csv"`
	Concurrency       int    `json:"concurrency"`
	RepeatTimes       int    `json:"repeat"`
	JWTString         string `json:"jwt"`
	AccountName       string `json:"account"`
	TableName         string `json:"table"`
}

func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
