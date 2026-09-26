package types

// Module represents the structure of a single module, including its metadata and dependencies.
type Module struct {
	// Name is the URL-safe name of the module.
	Name string `yaml:"name"`
	// Description provides a brief summary of the module.
	Description string `yaml:"description"`
	// Version specifies the current version of the module.
	Version string `yaml:"version"`
	// Languages lists the programming languages supported by the module.
	Languages []string `yaml:"languages"`
	// Dependencies maps each language to a list of required dependencies.
	Dependencies map[string][]string `yaml:"dependencies"`
}

// Registry represents the structure of the registry index file containing module and template information.
type Registry struct {
	// Schema specifies the version of the registry schema.
	Schema uint `yaml:"schema"`
	// Modules maps URL-safe module names to their corresponding metadata within the registry.
	Modules map[string]ObjectMetadata `yaml:"modules"`
	// Templates maps URL-safe template names to their corresponding metadata within the registry.
	Templates map[string]ObjectMetadata `yaml:"templates"`
}

// ObjectMetadata represents the metadata for a single project or module template within the registry index.
type ObjectMetadata struct {
	// Description provides a brief summary of the template.
	Description string `yaml:"description"`
	// Versions lists the available versions of the template.
	Versions []string `yaml:"versions"`
}

// ProjectTemplate represents the structure of a single project template, including its metadata.
type ProjectTemplate struct {
	// Name is the URL-safe name of the template.
	Name string `yaml:"name"`
	// Version specifies the current version of the template.
	Version string `yaml:"version"`
	// Language specifies the programming language of the template.
	Language string `yaml:"language"`
	// Type specifies the type/category of the template, e.g. "web", "cli", or "library".
	Type string `yaml:"type"`
	// Modules lists the modules used by the template.
	Modules []string `yaml:"modules"`
}

type PBConfig struct {
	Sources []string `yaml:"sources"`
}

type CompiledIndex struct {
	Modules   map[string]ObjectMetadata `yaml:"modules"`
	Templates map[string]ObjectMetadata `yaml:"templates"`
}