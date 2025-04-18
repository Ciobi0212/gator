package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configName string = ".gatorconfig.json"

type Config struct {
	Db_url           string `json:"db_url"`
	Current_username string `json:"current_username"`
}

func getConfigPath() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("err getting home dir: %w", err)
	}

	path := homePath + "/" + configName

	return path, nil
}

func (c *Config) SetUser(user string) error {
	c.Current_username = user

	fileBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	err = writeConfig(fileBytes)
	if err != nil {
		return fmt.Errorf("err writing config: %w", err)
	}

	return nil
}

func writeConfig(bytes []byte) error {
	path, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("err getting config path: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("err opening file: %w", err)
	}

	_, err = f.Write(bytes)
	if err != nil {
		return fmt.Errorf("err writing file: %w", err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("err closing file: %w", err)
	}

	return nil
}

func ReadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("err getting config path: %w", err)
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("err reading file: %w", err)
	}

	var config Config
	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("err unmarhaling file: %w", err)
	}

	return &config, nil
}
