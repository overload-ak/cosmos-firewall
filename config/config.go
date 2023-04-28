package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Config struct {
	ChainID   string   `json:"chain_id"`
	JSONRPC   []string `json:"json_rpc"`
	Rest      []string `json:"rest"`
	GRPC      []string `json:"grpc"`
	MinFee    string   `json:"min_fee"`
	Whitelist []string `json:"whitelist"`
}

func ReadConfig(fileName string) (*Config, error) {
	data, err := os.ReadFile(path.Join(fileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config.json is not exist, fileName %s, error: %s", fileName, err.Error())
		}
		return nil, err
	}
	var configs *Config
	if err = json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}
