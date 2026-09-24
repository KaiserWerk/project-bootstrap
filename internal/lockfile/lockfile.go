package lockfile

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type ModuleLock struct {
	Version string            `yaml:"version"`
	Commit  string            `yaml:"commit,omitempty"`
	Files   map[string]string `yaml:"files,omitempty"`
}

type Lockfile struct {
	Modules map[string]ModuleLock `yaml:"modules,omitempty"`
}

func New() *Lockfile {
	return &Lockfile{Modules: map[string]ModuleLock{}}
}

func Load(path string) (*Lockfile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var l Lockfile
	if err := yaml.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	if l.Modules == nil {
		l.Modules = map[string]ModuleLock{}
	}
	return &l, nil
}

func (l *Lockfile) Save(path string) error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}
