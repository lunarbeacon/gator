package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (*Config, error) {
	configFilepath, err := getConfigFilePath()
	if err != nil {
		return &Config{}, fmt.Errorf("could not get config file path: %w", err)
	}

	content, err := os.ReadFile(configFilepath)
	if err != nil {
		return &Config{}, fmt.Errorf("failed to read config file %v", err)
	}

	// 2. Define struct instance
	var config Config

	// 3. Unmarshal JSON into struct
	err = json.Unmarshal(content, &config)
	if err != nil {
		return &Config{}, fmt.Errorf("failed to unmarshal config file %v", err)
	}

	return &config, nil
}

func SetUser(config *Config, username string) error {
	config.CurrentUserName = username
	configFilepath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("could not get config file path: %w", err)
	}
	err = write(configFilepath, config)
	if err != nil {
		return fmt.Errorf("could not write config file to %s: %w", configFilepath, err)
	}
	return nil
}

func write(configFilePath string, config *Config) error {
	// Marshal with indentation for readability
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config file %v", err)
	}

	// Write to file
	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", configFilePath, err)
	}
	return nil
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory %v", err)
	}
	configFilepath := filepath.Join(home, configFileName)
	return configFilepath, nil

}
