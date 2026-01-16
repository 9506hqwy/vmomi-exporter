package config

import (
	"context"
	"os"

	"github.com/9506hqwy/vmomi-exporter/pkg/flag"
	"go.yaml.in/yaml/v4"
)

type Config struct {
	CounterConfig `yaml:",omitempty,inline"`
	ObjectConfig  `yaml:",omitempty,inline"`
	RootConfig    `yaml:",omitempty,inline"`
}

func DecodeConfig(config []byte) (*Config, error) {
	var c Config
	err := yaml.Unmarshal(config, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func EncodeConfig(c *Config) (string, error) {
	buf, err := yaml.Marshal(&c)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func DefaultConfig() *Config {
	return &Config{
		CounterConfig: *DefaultCounterConfig(),
		ObjectConfig:  *DefaultObjectConfig(),
		RootConfig:    *DefaultRootConfig(),
	}
}

func GetConfig(ctx context.Context) (*Config, error) {
	filePath, ok := ctx.Value(flag.ExporterConfigKey{}).(string)
	if !ok || filePath == "" {
		return DefaultConfig(), nil
	}

	config, err := LoadFileConfig(filePath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func LoadFileConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	config, err := DecodeConfig(data)
	if err != nil {
		return nil, err
	}

	return config, nil
}
