package project

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type ModuleEntry struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version,omitempty"`
}

type ProjectInfo struct {
	Name        string   `yaml:"name"`
	Language    string   `yaml:"language"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

type Manifest struct {
	Project ProjectInfo   `yaml:"project"`
	Modules []ModuleEntry `yaml:"modules,omitempty"`
}

// Load reads a manifest from the given path.
func Load(path string) (*Manifest, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Save writes the manifest to the given path (overwrites).
func (m *Manifest) Save(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}
