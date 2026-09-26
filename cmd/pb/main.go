package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
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
	cwd, err := os.Getwd()
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
	homeDir = filepath.Join(homeDir, ".pb")
	_ = os.MkdirAll(homeDir, 0o755)

	rootCmd := &cobra.Command{
		Use:   "pb",
		Short: "Project Bootstrap CLI",
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
			getModuleInfo(args[0])
		},
		Example: "pb info my-module",
	}

	// create-project
	createCmd := &cobra.Command{
		Use:   "create-project <template> [name]",
		Short: "Create a project from a template",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			template := args[0]
			name := args[1]
			createProject(template, name, cwd)
		},
		Example: "pb create-project my-template my-project-name",
	}

	// create-registry
	createIndexCmd := &cobra.Command{
		Use:   "create-registry",
		Short: "Create a new registry file",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			createRegistry(cwd)
		},
		Example: "pb create-registry",
	}

	// create-module
	createModuleCmd := &cobra.Command{
		Use:   "create-module <name>",
		Short: "Create a new module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			createModule(cwd, name)
		},
		Example: "pb create-module my-cool-module",
	}

	// create-template
	createTemplateCmd := &cobra.Command{
		Use:   "create-template <name>",
		Short: "Create a new template",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			fmt.Printf("Creating template: %s\n", name)
			createProjectTemplate(cwd, name)
		},
		Example: "pb create-template my-cool-template",
	}

	// create-config
	createConfigCmd := &cobra.Command{
		Use:   "create-config",
		Short: "Create a new configuration file",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			createConfig(homeDir)
		},
		Example: "pb create-config",
	}

	// add-module
	addCmd := &cobra.Command{
		Use:   "add-module <module>",
		Short: "Add/install a module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			//addModule(args[0])
		},
	}

	// list, update, index, doctor
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed modules",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pb list: list installed modules (stub)")
		},
		Example: "pb list",
	}
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update modules",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pb update: update modules (stub)")
		},
		Example: "pb update",
	}
	buildIndexCmd := &cobra.Command{
		Use:   "cache",
		Short: "Build registry cache",
		Run: func(cmd *cobra.Command, args []string) {
			buildCache(homeDir)
		},
		Example: "pb cache",
	}
	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Health checks",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pb doctor: health checks (stub)")
		},
		Example: "pb doctor",
	}

	rootCmd.AddCommand(searchCmd, infoCmd, createCmd, addCmd, listCmd, updateCmd, buildIndexCmd, doctorCmd, createModuleCmd, createIndexCmd, createTemplateCmd, createConfigCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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

func getModuleInfo(module string) {
	// cfg, err := config.LoadConfig()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "pb: failed to load config: %v\n", err)
	// 	return
	// }
	// reg := registry.New(cfg.Sources)
	// e, err := reg.FindModule(module)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "pb: %v\n", err)
	// 	os.Exit(2)
	// }
	// if e.Manifest != nil {
	// 	fmt.Printf("Module: %s\n", module)
	// 	fmt.Printf("Project Name: %s\n", e.Manifest.Project.Name)
	// 	fmt.Printf("Language: %s\n", e.Manifest.Project.Language)
	// 	if len(e.Manifest.Modules) > 0 {
	// 		fmt.Println("Modules included:")
	// 		for _, m := range e.Manifest.Modules {
	// 			fmt.Printf(" - %s @ %s\n", m.Name, m.Version)
	// 		}
	// 	}
	// 	fmt.Printf("Source repo path: %s\n", e.RepoDir)
	// } else {
	// 	fmt.Printf("Module: %s (no manifest available)\n", module)
	// }
}

func createProject(template, name, workDir string) {
	dst := filepath.Join(workDir, name)
	fmt.Printf("pb create-project: template=%s name=%s dst=%s\n", template, name, dst)
	// TODO: implement template copy and variable prompts
}

// getProjectTemplate retrieves the project template by name from the registry.
// It returns its metadata, the local path to the template directory, and an error if any.
// If a template with the given name is not found, the sources will be synced and the search retried.
func getProjectTemplate(template string) (types.ProjectTemplate, string, error) {
	panic("getProjectTemplate not implemented")
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
