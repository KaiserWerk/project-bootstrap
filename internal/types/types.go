package types

// Module
type (
	// Module represents the structure of a single module, including its metadata and dependencies.
	Module struct {
		Name         string              `yaml:"name"`
		Description  string              `yaml:"description"`
		Version      string              `yaml:"version"`
		Languages    []string            `yaml:"languages"`
		Dependencies map[string][]string `yaml:"dependencies"`
		Tags         []string            `yaml:"tags"`
	}
)

// Registry
type (
	// Registry represents the structure of the registry index file containing module and template information.
	Registry struct {
		Schema    uint                            `yaml:"schema"`
		Modules   map[string]RegistryModuleInfo   `yaml:"modules"`
		Templates map[string]RegistryTemplateInfo `yaml:"templates"`
	}
	// RegistryModuleInfo represents the metadata for a single module within the registry index.
	RegistryModuleInfo struct {
		Description string   `yaml:"description"`
		Tags        []string `yaml:"tags"`
		Versions    []string `yaml:"versions"`
	}

	// RegistryTemplateInfo represents the metadata for a single project template within the registry index.
	RegistryTemplateInfo struct {
		Description string   `yaml:"description"`
		Versions    []string `yaml:"versions"`
	}
)
