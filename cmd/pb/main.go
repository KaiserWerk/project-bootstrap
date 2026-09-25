package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bufio"
	"regexp"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/KaiserWerk/project-bootstrap/internal/config"
	"github.com/KaiserWerk/project-bootstrap/internal/global"
	golang "github.com/KaiserWerk/project-bootstrap/internal/language/golang"
	modulepkg "github.com/KaiserWerk/project-bootstrap/internal/module"
	"github.com/KaiserWerk/project-bootstrap/internal/registry"
	templatepkg "github.com/KaiserWerk/project-bootstrap/internal/template"
	"github.com/KaiserWerk/project-bootstrap/internal/types"
	"gopkg.in/yaml.v3"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

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
			noIndex, _ := cmd.Flags().GetBool("no-index")
			search(args[0], jsonOut, noIndex)
		},
		Example: "pb search 'my query'",
	}
	searchCmd.Flags().BoolP("json", "j", false, "Output JSON")
	searchCmd.Flags().BoolP("no-index", "n", false, "Do not refresh index")

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
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			varsPath, _ := cmd.Flags().GetString("vars")
			a := []string{args[0]}
			if len(args) > 1 {
				a = append(a, args[1])
			}
			if varsPath != "" {
				a = append(a, "--vars="+varsPath)
			}
			createProjectArgs(a)
		},
		Example: "pb create-project my-template my-project --vars=vars.yaml",
	}
	createCmd.Flags().String("vars", "", "Path to vars YAML file")

	// create-registry
	createIndexCmd := &cobra.Command{
		Use:   "create-registry",
		Short: "Create a new registry file",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pb: Creating registry file")
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
			fmt.Printf("Creating module: %s\n", name)
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
			createTemplate(cwd, name)
		},
		Example: "pb create-template my-cool-template",
	}

	// add-module
	addCmd := &cobra.Command{
		Use:   "add-module <module>",
		Short: "Add/install a module",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			addModule(args[0])
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
	indexCmdWrap := &cobra.Command{
		Use:   "index",
		Short: "Build registry index",
		Run: func(cmd *cobra.Command, args []string) {
			indexCmd()
		},
		Example: "pb index",
	}
	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Health checks",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pb doctor: health checks (stub)")
		},
		Example: "pb doctor",
	}

	rootCmd.AddCommand(searchCmd, infoCmd, createCmd, addCmd, listCmd, updateCmd, indexCmdWrap, doctorCmd, createModuleCmd, createIndexCmd, createTemplateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func createTemplate(cwd, name string) {
	t := types.Template{
		Name:     name,
		Version:  "0.0.0",
		Language: "golang",
		Type:     "cli",
		Modules:  []string{},
	}

	p := filepath.Join(cwd, name)
	if err := os.MkdirAll(filepath.Join(p, "content"), 0o755); err != nil {
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
		Modules: map[string]types.RegistryModuleInfo{
			"cool-module": {
				Description: "cool-module",
				Tags:        []string{"TagA", "TagB"},
				Versions:    []string{"0.0.0-alpha", "0.0.1"},
			},
		},
		Templates: map[string]types.RegistryTemplateInfo{
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
	if err := os.MkdirAll(filepath.Join(p, "content"), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed to create module directory: %v\n", err)
		return
	}

	mod := types.Module{
		Name:        name,
		Description: "Example module",
		Version:     "0.0.0",
		Languages:   []string{"golang"},
		Tags:        []string{"TagA", "TagB"},
		Dependencies: map[string][]string{
			"golang": []string{"github.com/example/jwt@v5.3.1"},
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

func search(query string, jsonOut, noIndex bool) {
	panic("unimplemented")
}

func searchOld(args []string) {
	jsonOut := false
	query := ""
	noIndex := false
	for _, a := range args {
		if a == "--json" || a == "-j" {
			jsonOut = true
			continue
		}
		if a == "--no-index" || a == "-n" {
			noIndex = true
			continue
		}
		if query == "" {
			query = a
		}
	}

	cfg := config.LoadConfig()
	fmt.Printf("pb search: query=%q\n", query)
	fmt.Printf("Configured sources: %v\n", cfg.Sources)
	reg := registry.New(cfg.Sources)

	// auto-update index so `pb search` reflects latest registry metadata (unless disabled)
	if !noIndex {
		if idxPath, err := reg.BuildIndex(); err != nil {
			fmt.Fprintf(os.Stderr, "pb: warning: failed building index: %v\n", err)
		} else {
			fmt.Printf("pb: refreshed index at %s\n", idxPath)
		}
	} else {
		fmt.Println("pb: skipping index refresh (--no-index)")
	}
	results, err := reg.Search(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: search error: %v\n", err)
		os.Exit(1)
	}
	if len(results) == 0 {
		fmt.Println("No results found")
		return
	}
	if jsonOut {
		out := struct {
			Results []registry.Entry `json:"results"`
		}{Results: results}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return
	}
	fmt.Println("Results:")
	for _, r := range results {
		name := r.Name
		if name == "" && r.Manifest != nil {
			name = r.Manifest.Project.Name
		}
		fmt.Printf(" - %s (repo: %s)\n", name, r.RepoDir)
	}
}

func getModuleInfo(module string) {
	cfg := config.LoadConfig()
	reg := registry.New(cfg.Sources)
	e, err := reg.FindModule(module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: %v\n", err)
		os.Exit(2)
	}
	if e.Manifest != nil {
		fmt.Printf("Module: %s\n", module)
		fmt.Printf("Project Name: %s\n", e.Manifest.Project.Name)
		fmt.Printf("Language: %s\n", e.Manifest.Project.Language)
		if len(e.Manifest.Modules) > 0 {
			fmt.Println("Modules included:")
			for _, m := range e.Manifest.Modules {
				fmt.Printf(" - %s @ %s\n", m.Name, m.Version)
			}
		}
		fmt.Printf("Source repo path: %s\n", e.RepoDir)
	} else {
		fmt.Printf("Module: %s (no manifest available)\n", module)
	}
}

func createProject(template, name string) {
	if name == "" {
		name = template + "-project"
	}
	dst := filepath.Join(".", name)
	fmt.Printf("pb create-project: template=%s name=%s dst=%s\n", template, name, dst)
	// TODO: implement template copy and variable prompts
}

func createProjectArgs(args []string) {
	tmpl := args[0]
	name := ""
	varsPath := ""
	for _, a := range args[1:] {
		if strings.HasPrefix(a, "--vars=") {
			varsPath = strings.TrimPrefix(a, "--vars=")
			continue
		}
		if name == "" {
			name = a
		}
	}
	if name == "" {
		name = tmpl + "-project"
	}

	cfg := config.LoadConfig()
	reg := registry.New(cfg.Sources)
	e, err := reg.FindModule(tmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: template not found: %v\n", err)
		os.Exit(2)
	}

	dst := filepath.Join(".", name)
	fmt.Printf("create-project: using template %s from %s -> %s\n", tmpl, e.RepoDir, dst)

	vars := map[string]string{}
	if varsPath != "" {
		data, err := os.ReadFile(varsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pb: failed reading vars file: %v\n", err)
			os.Exit(2)
		}
		if err := yaml.Unmarshal(data, &vars); err != nil {
			fmt.Fprintf(os.Stderr, "pb: failed parsing vars file: %v\n", err)
			os.Exit(2)
		}
	}

	// collect variables from template and prompt for any missing
	needed, err := templatepkg.CollectVariables(e.RepoDir)
	if err == nil && len(needed) > 0 {
		isTTY := term.IsTerminal(int(os.Stdin.Fd()))
		reader := bufio.NewReader(os.Stdin)
		for _, v := range needed {
			if _, ok := vars[v]; ok {
				continue
			}
			// check manifest defaults/validation
			varDef := ""
			varValidate := ""
			varRequired := false
			if e.Manifest != nil {
				if tv, ok := e.Manifest.TemplateVars[v]; ok {
					varDef = tv.Default
					varValidate = tv.Validate
					varRequired = tv.Required
				}
			}

			if !isTTY {
				if varDef != "" {
					vars[v] = varDef
					continue
				}
				if varRequired {
					fmt.Fprintf(os.Stderr, "pb: missing required variable %s and no TTY available\n", v)
					os.Exit(2)
				}
				// non-required and no default: leave empty
				vars[v] = ""
				continue
			}

			for {
				prompt := v
				if varDef != "" {
					prompt = fmt.Sprintf("%s (default: %s)", v, varDef)
				}
				fmt.Printf("Enter value for %s: ", prompt)
				line, _ := reader.ReadString('\n')
				val := strings.TrimSpace(line)
				if val == "" && varDef != "" {
					val = varDef
				}
				if val == "" && varRequired {
					fmt.Println("value required")
					continue
				}
				if varValidate != "" && val != "" {
					re, rerr := regexp.Compile(varValidate)
					if rerr != nil {
						fmt.Fprintf(os.Stderr, "pb: invalid validate regex for %s: %v\n", v, rerr)
						break
					}
					if !re.MatchString(val) {
						fmt.Println("value does not match required pattern")
						continue
					}
				}
				vars[v] = val
				break
			}
		}
	}

	if err := templatepkg.ApplyTemplate(e.RepoDir, dst, vars); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed creating project: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Project created at", dst)
}

func addModule(mod string) {
	cfg := config.LoadConfig()
	reg := registry.New(cfg.Sources)
	e, err := reg.FindModule(mod)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("Installing module %s from %s (manifest project: %s)\n", mod, e.RepoDir, e.Name)
	if err := modulepkg.InstallFromRepo(e.RepoDir, mod, "."); err != nil {
		fmt.Fprintf(os.Stderr, "pb: install failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Module installed into current project")
	if e.Manifest != nil && e.Manifest.Project.Language == "go" {
		if err := golang.RunTidy("."); err != nil {
			fmt.Fprintf(os.Stderr, "pb: go mod tidy failed: %v\n", err)
		} else {
			fmt.Println("Ran go mod tidy")
		}
	}
}

func indexCmd() {
	cfg := config.LoadConfig()
	reg := registry.New(cfg.Sources)
	idxPath, err := reg.BuildIndex()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed building index: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Index written to:", idxPath)
}
