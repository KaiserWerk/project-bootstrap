package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KaiserWerk/project-bootstrap/internal/config"
	golang "github.com/KaiserWerk/project-bootstrap/internal/language/golang"
	modulepkg "github.com/KaiserWerk/project-bootstrap/internal/module"
	"github.com/KaiserWerk/project-bootstrap/internal/registry"
	templatepkg "github.com/KaiserWerk/project-bootstrap/internal/template"
	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "init":
		fmt.Println("pb init: initialize local configuration (stub)")
	case "search":
		args := os.Args[2:]
		search(args)
	case "info":
		if len(os.Args) < 3 {
			fmt.Println("usage: pb info <module>")
			os.Exit(2)
		}
		info(os.Args[2])
	case "create-project":
		if len(os.Args) < 3 {
			fmt.Println("usage: pb create-project <template> [name] [--vars=vars.yaml]")
			os.Exit(2)
		}
		createProjectArgs(os.Args[2:])
	case "add-module":
		if len(os.Args) < 3 {
			fmt.Println("usage: pb add-module <module>")
			os.Exit(2)
		}
		addModule(os.Args[2])
	case "list":
		fmt.Println("pb list: list installed modules (stub)")
	case "update":
		fmt.Println("pb update: update modules (stub)")
	case "doctor":
		fmt.Println("pb doctor: health checks (stub)")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Println("pb — Project Bootstrap CLI")
	fmt.Println()
	fmt.Println("Usage: pb <command> [args]")
	fmt.Println()
	fmt.Println("Commands: init, search, info, create-project, add-module, list, update, doctor")
}

func search(args []string) {
	jsonOut := false
	query := ""
	for _, a := range args {
		if a == "--json" || a == "-j" {
			jsonOut = true
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

func info(module string) {
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
