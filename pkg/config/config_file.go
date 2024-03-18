package config

import (
	"fmt"
	"io"
	"os"
	"time"

	yaml "gopkg.in/yaml.v3"
)

type ConfigFile struct {
	Binance struct {
		APIKey       string `yaml:"api_key"`
		APIKeySecret string `yaml:"api_key_secret"`
		Tld          string `yaml:"tld"`
	} `yaml:"binance"`
	Bridge string   `yaml:"bridge"`
	Coins  []string `yaml:"coins"`

	StartCoin *string `yaml:"start_coin"`

	Jump struct {
		WhenGain   float64       `yaml:"when_gain"`
		DecreaseBy float64       `yaml:"decrease_by"`
		After      time.Duration `yaml:"after"`
	} `yaml:"jump"`
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

	// To debug if the config is correctly parsed
	// yamled, _ := yaml.Marshal(data)
	// fmt.Print(string(yamled))

	return data, nil
}
