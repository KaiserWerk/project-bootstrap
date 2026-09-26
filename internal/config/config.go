package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KaiserWerk/project-bootstrap/internal/types"
	"gopkg.in/yaml.v3"
)

// LoadConfig tries to load ~/.pb/pb-config.yaml and returns default config when absent.
func LoadConfig() (*types.PBConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".pb")
	path := filepath.Join(dir, "pb-config.yaml")

	_ = os.MkdirAll(dir, 0o755) // just to make sure the directory exists

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg types.PBConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to parse config %s: %v\n", path, err)
		return nil, err
	}

	return &cfg, nil
}
