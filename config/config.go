package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

var envRe = regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

type Config struct {
	Database `yaml:"database"`
	HTTP     `yaml:"http"`
}

type Database struct {
	URL string `yaml:"url"`
}

type HTTP struct {
	Addr string `yaml:"addr"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	expanded := envRe.ReplaceAllStringFunc(string(data), func(m string) string {
		key := envRe.FindStringSubmatch(m)[1]
		if v := os.Getenv(key); v != "" {
			return v
		}
		return m
	})

	var c Config
	if err := yaml.Unmarshal([]byte(expanded), &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &c, nil
}
