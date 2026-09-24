package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Sources []string `yaml:"sources"`
}

// LoadConfig tries to load ~/.pb/pb-config.yaml and returns default config when absent.
func LoadConfig() Config {
	var cfg Config
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}
	path := filepath.Join(home, ".pb", "pb-config.yaml")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		// no config, return empty
		return cfg
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to parse config %s: %v\n", path, err)
		return Config{}
	}
	return cfg
}
