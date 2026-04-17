package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

var envRe = regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

type Config struct {
	Database  `yaml:"database"`
	Migration `yaml:"migration"`
	HTTP      `yaml:"http"`
}

type Database struct {
	URL      string `yaml:"url"`
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type Migration struct {
	Dir string `yaml:"dir"`
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
		return nil, fmt.Errorf("read config %q: %w", configPath, err)
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
		return nil, fmt.Errorf("parse config %q: %w", configPath, err)
	}

	return &c, nil
}

func (d *Database) ConnStr(extraParams string) string {
	if d.URL != "" {
		return d.URL + extraParams
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s", d.User, d.Password, d.Host, d.Name)
	if extraParams != "" {
		connStr += "?" + extraParams
	}
	return connStr
}
