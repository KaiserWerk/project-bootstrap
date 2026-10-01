package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/KaiserWerk/project-bootstrap/internal/config"
	"github.com/KaiserWerk/project-bootstrap/internal/global"
	"github.com/KaiserWerk/project-bootstrap/internal/output"
	"github.com/KaiserWerk/project-bootstrap/internal/registry"
	"github.com/KaiserWerk/project-bootstrap/internal/types"

	"gopkg.in/yaml.v3"
)

func main() {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	homeDir = filepath.Clean(homeDir)
	homeDir = filepath.Join(homeDir, global.DirectorynameConfigFolder)
	_ = os.MkdirAll(homeDir, 0o755)

	rootCmd := &cobra.Command{
		Use:   "pb",
		Short: "Project Bootstrap CLI",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	// search
	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search modules",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jsonOut, _ := cmd.Flags().GetBool("json")
			refresh, _ := cmd.Flags().GetBool("refresh")
			searchModules(args[0], homeDir, jsonOut, refresh)
		},
		Example: "pb search 'my query'",
	}
	searchCmd.Flags().BoolP("json", "j", false, "Output JSON")
	searchCmd.Flags().BoolP("refresh", "r", false, "Refresh index")

	// info
	infoCmd := &cobra.Command{
		Use:   "info <module>",
		Short: "Show information about a module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moduleName := args[0]
			getModuleInfo(moduleName, homeDir)
		},
		Example: "pb info my-module",
	}

	// create-project
	createCmd := &cobra.Command{
		Use:   "create-project <template> [name]",
		Short: "Create a project from a template",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			templateName := args[0]
			projectName := args[1]
			createProject(templateName, projectName, workDir, homeDir)
		},
		Example: "pb create-project my-template my-project-name",
	}

	// create-registry
	createRegistryCmd := &cobra.Command{
		Use:   "create-registry",
		Short: "Create a new registry file",
		Long: `Creates a new registry file in the current working directory, prefilled with some example values.` +
			`The file is used as an index file for a module/template repository. Edit the file manually as needed.`,
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			createRegistry(workDir)
		},
		Example: "pb create-registry",
	}

	// update-registry
	updateRegistryCmd := &cobra.Command{
		Use:   "update-registry",
		Short: "Update the registry file",
		Long:  `Updates the registry file in the current working directory. The file is used as an index file for a module/template repository. Edit the file manually as needed.`,
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			updateRegistry(workDir)
		},
		Example: "pb update-registry",
	}

	// create-module
	createModuleCmd := &cobra.Command{
		Use:   "create-module <name>",
		Short: "Create a new module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			createModule(workDir, name)
		},
		Example: "pb create-module my-cool-module",
	}

	// create-template
	createTemplateCmd := &cobra.Command{
		Use:   "create-template <name>",
		Short: "Create a new template",
		Long: `Creates a new project template in the current directory. The template can later be used to bootstrap new projects.` +
			`Edit the template files manually as needed and upload them to your repository.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			fmt.Printf("Creating template: %s\n", name)
			createProjectTemplate(workDir, name)
		},
		Example: "pb create-template my-cool-template",
	}

	// create-config
	createConfigCmd := &cobra.Command{
		Use:   "create-config",
		Short: "Create a new global configuration file",
		Long: `Creates a new global configuration file at ~/.pb/pb-config.yaml with default values. Edit the file manually as needed.` +
			`Subsequent runs of pb will use this configuration. Re-running this command while a configuration file exists will have no effect.`,
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			createConfig(homeDir)
		},
		Example: "pb create-config",
	}

	// add-module
	addModuleCmd := &cobra.Command{
		Use:   "add-module <module>",
		Short: "Add a module",
		Long: `Adds the code files of the named module in the current directory, typically your project.` +
			`The module info and its files will be fetched from the first configured source where it's available and added to the project.` +
			`A version string without prefix (e.g., "1.0.0" instead of "v1.0.0") can be appended to the name to add that specific version.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moduleName := args[0]
			moduleVersion := ""
			if strings.Contains(moduleName, "@") {
				parts := strings.Split(moduleName, "@")
				moduleName = parts[0]
				moduleVersion = parts[1]
			}
			addModule(moduleName, moduleVersion, workDir, homeDir)
		},
	}

	buildCacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Build registry cache",
		Long: `Builds the local cache of all configured registry sources. This involves cloning or pulling the latest changes from each ` +
			`source repository into the local cache directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			buildCache(homeDir)
		},
		Example: "pb cache",
	}

	rootCmd.AddCommand(searchCmd, infoCmd, createCmd, addModuleCmd, buildCacheCmd, createModuleCmd, createRegistryCmd, createTemplateCmd, createConfigCmd, updateRegistryCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func updateRegistry(workDir string) {
	// read registry file from the current working directory
	registryPath := filepath.Join(workDir, global.FilenameRegistryYAML)
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "pb: registry file does not exist at %s\n", registryPath)
		return
	}
	// load the registry file
	data, err := os.ReadFile(registryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to read registry file: %v\n", err)
		return
	}
	var reg types.Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to unmarshal registry file: %v\n", err)
		return
	}
	fmt.Println("registry.yaml loaded successfully.")

	reg.Modules = make(map[string]types.ObjectMetadata)   // clear the modules map to store the current module versions
	reg.Templates = make(map[string]types.ObjectMetadata) // clear the templates map to store the current template versions

	// read all direct subdirectories of the modules directory. They represent individual modules. The subdirectory name is the module name.
	// Underneath, the version directories must be read by name to produce a string slice of versions for each module.
	// From the newest/highest version, the <module-name>/<version>/module.yaml is read and parsed to get the most reced descrption.
	filepath.Walk(filepath.Join(workDir, "modules"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "pb: failed to walk modules directory: %v\n", err)
			return err
		}

		if info.IsDir() {
			moduleName := info.Name()
			fmt.Println("Found module:", moduleName)

			var versions []string
			var description string
			filepath.Walk(filepath.Join(workDir, "modules", moduleName), func(path string, info os.FileInfo, err error) error {
				if err != nil {
					fmt.Fprintf(os.Stderr, "pb: failed to walk module directory: %v\n", err)
					return err
				}

				if info.IsDir() && info.Name() != moduleName {
					version := info.Name()
					fmt.Println("Found version:", version)
					versions = append(versions, version)
				}

				return nil
			})

			slices.Sort(versions)

			if len(versions) > 0 {
				latestVersion := versions[len(versions)-1]
				modulePath := filepath.Join(workDir, "modules", moduleName, latestVersion, global.FilenameModuleYAML)
				data, err := os.ReadFile(modulePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "pb: failed to read module file: %v\n", err)
					return nil
				}

				var module types.Module
				if err := yaml.Unmarshal(data, &module); err != nil {
					fmt.Fprintf(os.Stderr, "pb: failed to unmarshal module file: %v\n", err)
					return nil
				}

				description = module.Description
			}

			reg.Modules[moduleName] = types.ObjectMetadata{
				Description: description,
				Versions:    versions,
			}
		}

		return nil
	})

	// read module metadata
	filepath.Walk(filepath.Join(workDir, "templates"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "pb: failed to walk templates directory: %v\n", err)
			return err
		}

		if info.IsDir() {
			templateName := info.Name()
			fmt.Println("Found template:", templateName)

			var versions []string
			var description string
			filepath.Walk(filepath.Join(workDir, "templates", templateName), func(path string, info os.FileInfo, err error) error {
				if err != nil {
					fmt.Fprintf(os.Stderr, "pb: failed to walk templates directory: %v\n", err)
					return err
				}

				if info.IsDir() && info.Name() != templateName {
					version := info.Name()
					fmt.Println("Found version:", version)
					versions = append(versions, version)
				}

				return nil
			})

			slices.Sort(versions)

			if len(versions) > 0 {
				latestVersion := versions[len(versions)-1]
				modulePath := filepath.Join(workDir, "templates", templateName, latestVersion, global.FilenameTemplateYAML)
				data, err := os.ReadFile(modulePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "pb: failed to read template file: %v\n", err)
					return nil
				}

				var module types.Module
				if err := yaml.Unmarshal(data, &module); err != nil {
					fmt.Fprintf(os.Stderr, "pb: failed to unmarshal template file: %v\n", err)
					return nil
				}

				description = module.Description
			}

			reg.Modules[templateName] = types.ObjectMetadata{
				Description: description,
				Versions:    versions,
			}
		}

		return nil
	})

	y, err := yaml.Marshal(reg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to marshal registry to YAML: %v\n", err)
		return
	}
	if err := os.WriteFile(registryPath, y, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to write back registry file: %v\n", err)
		return
	}
}

func createConfig(workDir string) {
	fmt.Println("Creating config in", workDir)
	configPath := filepath.Join(workDir, global.FilenameConfigYAML)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := types.PBConfig{
			Sources: []string{"github.com/KaiserWerk/pb-registry"},
		}

		y, err := yaml.Marshal(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pb: failed to marshal config to YAML: %v\n", err)
			return
		}

		if err := os.WriteFile(configPath, y, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "pb: failed to create config file: %v\n", err)
			return
		}
	} else {
		fmt.Println("Config already exists at", configPath)
	}
}

func buildCache(workDir string) {
	// 1. read sources from configuration file
	pbConfig, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to load config: %v\n", err)
		return
	}
	// 2. download (git clone) or update (git pull) each source repository in(to) ~/.pb/sources.
	for _, source := range pbConfig.Sources {
		sourceDir := path.Join(workDir, "cache", "sources")
		sourceName := path.Base(source)
		// check whether the source exists locally
		sourceExists := sourceExists(filepath.Join(sourceDir, sourceName))
		// source exists locally, consider updating it
		if sourceExists {
			if err := registry.UpdateRegistry(source, sourceDir, sourceName); err != nil {
				fmt.Fprintf(os.Stderr, "pb: failed to update registry from source %s: %v\n", source, err)
				return
			}
		} else {
			// download the registry from the source
			if err := registry.DownloadRegistry(source, sourceDir, sourceName); err != nil {
				fmt.Fprintf(os.Stderr, "pb: failed to download registry from source %s: %v\n", source, err)
				return
			}
		}
	}

	fmt.Println("cache built successfully.")
}

func sourceExists(sourceDir string) bool {
	_, err := os.Stat(sourceDir)
	return err == nil
}

func createProjectTemplate(cwd, name string) {
	t := types.ProjectTemplate{
		Name:     name,
		Version:  "0.0.0",
		Language: "golang",
		Type:     "cli",
		Modules:  []string{},
	}

	p := filepath.Join(cwd, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to create template directory: %v\n", err)
		return
	}

	y, err := yaml.Marshal(t)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to marshal template to YAML: %v\n", err)
		return
	}
	if err := os.WriteFile(filepath.Join(p, global.FilenameTemplateYAML), y, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to write %s: %v\n", global.FilenameTemplateYAML, err)
		return
	}

	fmt.Printf("pb: done writing %s. You're ready to add files and folder.\n", global.FilenameTemplateYAML)
}

func createRegistry(cwd string) {
	i := types.Registry{
		Schema: 1,
		Modules: map[string]types.ObjectMetadata{
			"cool-module": {
				Description: "cool-module",
				Versions:    []string{"0.0.0-alpha", "0.0.1"},
			},
		},
		Templates: map[string]types.ObjectMetadata{
			"cool-project-template": {
				Description: "cool-project-template",
				Versions:    []string{"1.3.22", "1.3.23"},
			},
		},
	}

	y, err := yaml.Marshal(i)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to marshal registry to YAML: %v\n", err)
		return
	}

	if err := os.WriteFile(filepath.Join(cwd, global.FilenameRegistryYAML), y, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to write %s: %v\n", global.FilenameRegistryYAML, err)
		return
	}

	fmt.Println("pb: " + global.FilenameRegistryYAML + " created.")
}

func createModule(cwd, name string) {
	p := filepath.Join(cwd, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to create module directory: %v\n", err)
		return
	}

	mod := types.Module{
		Name:        name,
		Description: "Example module",
		Version:     "0.0.0",
		Languages:   []string{"golang"},
		Dependencies: map[string][]string{
			"golang": {"github.com/example/jwt@v1.2.3"},
		},
	}

	y, err := yaml.Marshal(mod)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to marshal module to YAML: %v\n", err)
		return
	}
	if err := os.WriteFile(filepath.Join(p, global.FilenameModuleYAML), y, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to write %s: %v\n", global.FilenameModuleYAML, err)
		return
	}

	fmt.Printf("pb: done writing %s. You're ready to add files and folder.\n", global.FilenameModuleYAML)
}

func searchModules(query, workDir string, jsonOut, refresh bool) {
	var outputWriter output.ModuleWriter = output.DefaultTextModuleWriter
	if jsonOut {
		outputWriter = output.DefaultJSONModuleWriter
	}

	if refresh {
		buildCache(workDir)
	}

	// first the cache lookup
	modules := getModulesFromCache(query, filepath.Join(workDir, "cache", "sources"))
	outputWriter.WriteRegistryModule(modules)

}

func getModulesFromCache(query, workDir string) []types.ObjectMetadata {
	// loop through all local, canced source directories and collect matching modules
	// directory: ~/.pb/cache/sources/<source_name>/tool-registry/registry.yaml, which is of type types.Registry.
	// workDir points to ~/.pb/cache/sources.
	var modules []types.ObjectMetadata

	// iterate over all source directories in the cache
	files, err := os.ReadDir(workDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to read cache directory: %v\n", err)
		return modules
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		registryPath := filepath.Join(workDir, f.Name(), "tool-registry", "registry.yaml")
		data, err := os.ReadFile(registryPath)
		if err != nil {
			continue
		}
		var reg types.Registry
		if err := yaml.Unmarshal(data, &reg); err != nil {
			continue
		}
		for name, meta := range reg.Modules {
			if strings.Contains(name, query) {
				meta.Name = name
				modules = append(modules, meta)
			}
		}
	}

	return modules
}

func getModuleInfo(moduleName, homeDir string) {
	// iterate over sources in cache
	moduleDir, err := getModuleTemplate(moduleName, "", homeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to get module template: %v\n", err)
		return
	}

	moduleInfo, err := getModuleMetadata(moduleDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to get module metadata: %v\n", err)
		return
	}
	fmt.Println("pb info:")
	fmt.Printf("  name: %s\n", moduleInfo.Name)
	fmt.Printf("  description: %s\n", moduleInfo.Description)
	fmt.Printf("  version: %s\n", moduleInfo.Version)
	fmt.Println("  languages:")
	for _, lang := range moduleInfo.Languages {
		fmt.Printf("    - %s\n", lang)
	}
	fmt.Println("  dependencies:")
	for lang, dep := range moduleInfo.Dependencies {
		fmt.Printf("    %s:\n", lang)
		for _, d := range dep {
			fmt.Printf("      - %s\n", d)
		}
	}
}

func getModuleMetadata(moduleDir string) (types.Module, error) {
	cont, err := os.ReadFile(filepath.Join(moduleDir, global.FilenameModuleYAML))
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to read module YAML: %v\n", err)
		return types.Module{}, err
	}

	var module types.Module
	if err := yaml.Unmarshal(cont, &module); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to unmarshal module YAML: %v\n", err)
		return types.Module{}, err
	}

	return module, nil
}

// copies the files belonging to the specified module from the cache to the current working directory.
func addModule(moduleName, moduleVersion, workDir, homeDir string) {
	// 1. determine the module template path in the local cache.
	moduleDir, err := getModuleTemplate(moduleName, moduleVersion, homeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to get module template: %v\n", err)
		return
	}

	// 3. recursively copy template files from templatePath into the projectDir.
	if err := copyRecursively(moduleDir, workDir); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to copy module files: %v\n", err)
		return
	}

	versionStr := ""
	if moduleVersion != "" {
		versionStr = "@" + moduleVersion
	}

	fmt.Printf("pb add-module: successfully added module %s%s to %s\n", moduleName, versionStr, workDir)
}

func createProject(templateName, projectName, workDir, configDir string) {
	projectDir := filepath.Join(workDir, projectName)

	// 1. create project directory
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to create project directory: %v\n", err)
		return
	}

	// 2. determine the template path by iterating through local sources. if not found, sync sources once and retry
	templatePath, err := getProjectTemplate(templateName, configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to get project template: %v\n", err)
		return
	}

	// 3. recursively copy template files from templatePath into the projectDir.
	if err := copyRecursively(templatePath, projectDir); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to copy template files: %v\n", err)
		return
	}
	fmt.Printf("pb create-project: successfully created project at %s\n", projectDir)

}

// copyRecursively copies all files and directories from sourceDir to targetDir, preserving the directory structure.
// Directories are created as needed. YAML files are not copied.
func copyRecursively(sourceDir, targetDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(targetDir, relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}
		if filepath.Ext(info.Name()) == ".yaml" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, 0o644)
	})
}

// getModuleTemplate retrieves the module template by name from the local cached registry.
// It returns the local path to the module template directory, and an error if any.
// If a template with the given name is not found, the sources will be synced and the search retried.
func getModuleTemplate(moduleName, moduleVersion, configDir string) (string, error) {
	sourcesDir := filepath.Join(configDir, "cache", "sources")

	// read all registry.yaml files from the subdirectories sourcesDir/<source_name>/tool-registry/registry.yaml
	entries, err := os.ReadDir(sourcesDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		registry, err := readSourceRegistry(entry, sourcesDir, entry.Name())
		if err != nil {
			fmt.Printf("failed to read registry.yaml from source %q: %s. Refreshing the cache might fix this.\n", entry.Name(), err.Error())
			continue
		}

		for name, meta := range registry.Modules {
			if name == moduleName {
				// module was found by name, but the requested version may not exist.
				if moduleVersion == "" {
					// if no specific version is requested, use the latest available version
					moduleVersion = latestVersion(meta.Versions)
				} else {
					// if a specific version is requested, check if it exists among the available versions
					if !slices.Contains(meta.Versions, moduleVersion) {
						// if not, skip the remaining modules to check the next source
						fmt.Printf("module %q with version %s was not found in source %q\n", moduleName, moduleVersion, entry.Name())
						break
					}
				}
				return filepath.Join(sourcesDir, entry.Name(), "tool-registry", "modules", name, moduleVersion), nil
			}
		}
	}
	return "", fmt.Errorf("template %q not found. A cache refresh may help.", moduleName)
}

func readSourceRegistry(entry os.DirEntry, sourcesDir, s string) (types.Registry, error) {
	if !entry.IsDir() {
		return types.Registry{}, fmt.Errorf("entry is not a directory")
	}
	registryPath := filepath.Join(sourcesDir, entry.Name(), "tool-registry", "registry.yaml")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return types.Registry{}, fmt.Errorf("registry.yaml not found")
	}

	// read and parse the registry.yaml file
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return types.Registry{}, err
	}
	var registry types.Registry
	err = yaml.Unmarshal(data, &registry)

	return registry, err
}

// getProjectTemplate retrieves the project template by name from the local cached registry.
// It returns its metadata, the local path to the template directory, and an error if any.
// If a template with the given name is not found, the sources will be synced and the search retried.
func getProjectTemplate(templateName string, configDir string) (string, error) {
	sourcesDir := filepath.Join(configDir, "cache", "sources")
	// read all registry.yaml files from the subdirectories sourcesDir/<source_name>/tool-registry/registry.yaml
	entries, err := os.ReadDir(sourcesDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		registryPath := filepath.Join(sourcesDir, entry.Name(), "tool-registry", "registry.yaml")
		if _, err := os.Stat(registryPath); os.IsNotExist(err) {
			continue
		}

		// read and parse the registry.yaml file
		data, err := os.ReadFile(registryPath)
		if err != nil {
			return "", err
		}
		var registry types.Registry
		if err := yaml.Unmarshal(data, &registry); err != nil {
			return "", err
		}
		for name, meta := range registry.Templates {
			if name == templateName {
				return filepath.Join(sourcesDir, entry.Name(), "tool-registry", "templates", name, latestVersion(meta.Versions)), nil
			}
		}
	}
	return "", fmt.Errorf("template %q not found. A cache refresh may help.", templateName)
}

// latestVersion compares semantic version strings with the 'v' prefix and returns the latest/newest version.
func latestVersion(s []string) string {
	if len(s) == 0 {
		return ""
	}
	latest := s[0]
	for _, v := range s[1:] {
		if compareVersions(v, latest) > 0 {
			latest = v
		}
	}
	return latest
}

// compareVersions compares two semantic version strings with the 'v' prefix.
// It returns 1 if a > b, -1 if a < b, and 0 if equal.
func compareVersions(a, b string) int {
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		if aParts[i] > bParts[i] {
			return 1
		} else if aParts[i] < bParts[i] {
			return -1
		}
	}
	if len(aParts) > len(bParts) {
		return 1
	} else if len(aParts) < len(bParts) {
		return -1
	}
	return 0
}

// func createProjectArgs(args []string) {
// 	tmpl := args[0]
// 	name := ""
// 	varsPath := ""
// 	for _, a := range args[1:] {
// 		if strings.HasPrefix(a, "--vars=") {
// 			varsPath = strings.TrimPrefix(a, "--vars=")
// 			continue
// 		}
// 		if name == "" {
// 			name = a
// 		}
// 	}
// 	if name == "" {
// 		name = tmpl + "-project"
// 	}

// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "pb: failed to load config: %v\n", err)
// 		return
// 	}
// 	reg := registry.New(cfg.Sources)
// 	e, err := reg.FindModule(tmpl)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "pb: template not found: %v\n", err)
// 		os.Exit(2)
// 	}

// 	dst := filepath.Join(".", name)
// 	fmt.Printf("create-project: using template %s from %s -> %s\n", tmpl, e.RepoDir, dst)

// 	vars := map[string]string{}
// 	if varsPath != "" {
// 		data, err := os.ReadFile(varsPath)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "pb: failed reading vars file: %v\n", err)
// 			os.Exit(2)
// 		}
// 		if err := yaml.Unmarshal(data, &vars); err != nil {
// 			fmt.Fprintf(os.Stderr, "pb: failed parsing vars file: %v\n", err)
// 			os.Exit(2)
// 		}
// 	}

// 	// collect variables from template and prompt for any missing
// 	needed, err := templatepkg.CollectVariables(e.RepoDir)
// 	if err == nil && len(needed) > 0 {
// 		isTTY := term.IsTerminal(int(os.Stdin.Fd()))
// 		reader := bufio.NewReader(os.Stdin)
// 		for _, v := range needed {
// 			if _, ok := vars[v]; ok {
// 				continue
// 			}
// 			// check manifest defaults/validation
// 			varDef := ""
// 			varValidate := ""
// 			varRequired := false
// 			if e.Manifest != nil {
// 				if tv, ok := e.Manifest.TemplateVars[v]; ok {
// 					varDef = tv.Default
// 					varValidate = tv.Validate
// 					varRequired = tv.Required
// 				}
// 			}

// 			if !isTTY {
// 				if varDef != "" {
// 					vars[v] = varDef
// 					continue
// 				}
// 				if varRequired {
// 					fmt.Fprintf(os.Stderr, "pb: missing required variable %s and no TTY available\n", v)
// 					os.Exit(2)
// 				}
// 				// non-required and no default: leave empty
// 				vars[v] = ""
// 				continue
// 			}

// 			for {
// 				prompt := v
// 				if varDef != "" {
// 					prompt = fmt.Sprintf("%s (default: %s)", v, varDef)
// 				}
// 				fmt.Printf("Enter value for %s: ", prompt)
// 				line, _ := reader.ReadString('\n')
// 				val := strings.TrimSpace(line)
// 				if val == "" && varDef != "" {
// 					val = varDef
// 				}
// 				if val == "" && varRequired {
// 					fmt.Println("value required")
// 					continue
// 				}
// 				if varValidate != "" && val != "" {
// 					re, rerr := regexp.Compile(varValidate)
// 					if rerr != nil {
// 						fmt.Fprintf(os.Stderr, "pb: invalid validate regex for %s: %v\n", v, rerr)
// 						break
// 					}
// 					if !re.MatchString(val) {
// 						fmt.Println("value does not match required pattern")
// 						continue
// 					}
// 				}
// 				vars[v] = val
// 				break
// 			}
// 		}
// 	}

// 	if err := templatepkg.ApplyTemplate(e.RepoDir, dst, vars); err != nil {
// 		fmt.Fprintf(os.Stderr, "pb: failed creating project: %v\n", err)
// 		os.Exit(1)
// 	}
// 	fmt.Println("Project created at", dst)
// }

// func addModule(mod string) {
// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "pb: failed to load config: %v\n", err)
// 		return
// 	}
// 	reg := registry.New(cfg.Sources)
// 	e, err := reg.FindModule(mod)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "pb: %v\n", err)
// 		os.Exit(2)
// 	}
// 	fmt.Printf("Installing module %s from %s (manifest project: %s)\n", mod, e.RepoDir, e.Name)
// 	if err := modulepkg.InstallFromRepo(e.RepoDir, mod, "."); err != nil {
// 		fmt.Fprintf(os.Stderr, "pb: install failed: %v\n", err)
// 		os.Exit(1)
// 	}
// 	fmt.Println("Module installed into current project")
// 	if e.Manifest != nil && e.Manifest.Project.Language == "go" {
// 		if err := golang.RunTidy("."); err != nil {
// 			fmt.Fprintf(os.Stderr, "pb: go mod tidy failed: %v\n", err)
// 		} else {
// 			fmt.Println("Ran go mod tidy")
// 		}
// 	}
// }

// func indexCmd() {
// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "pb: failed to load config: %v\n", err)
// 		return
// 	}
// 	reg := registry.New(cfg.Sources)
// 	idxPath, err := reg.BuildIndex()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "pb: failed building index: %v\n", err)
// 		os.Exit(1)
// 	}
// 	fmt.Println("Index written to:", idxPath)
// }
