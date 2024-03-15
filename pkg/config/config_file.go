package config

import (
	"fmt"
	"io"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type ConfigFile struct {
	Binance struct {
		APIKey       string
		APIKeySecret string
		Tld          string
	}
	Bridge string
	Coins  []string
}

func ParseConfigFile() (ConfigFile, error) {
	var res ConfigFile

	filepath := "config/config.yaml" // TODO Get it more dynamically ?
	file, err := os.Open(filepath)
	if err != nil {
		return res, fmt.Errorf("failed opening file: %w", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return res, fmt.Errorf("failed reading file: %w", err)
	}

	var data ConfigFile
	if err := yaml.Unmarshal(content, &data); err != nil {
		return res, fmt.Errorf("failed unmarshaling file: %w", err)
	}

	return data, nil
}
