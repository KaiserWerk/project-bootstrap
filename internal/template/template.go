package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"text/template"
)

// ApplyTemplate copies files from srcRoot into dstRoot, processing each file as a Go text/template
// using the provided vars map. It skips .git directories and does not copy a top-level pb.yaml.
func ApplyTemplate(srcRoot, dstRoot string, vars map[string]string) error {
	if _, err := os.Stat(srcRoot); err != nil {
		return fmt.Errorf("src template not found: %w", err)
	}
	if err := os.MkdirAll(dstRoot, 0o755); err != nil {
		return err
	}

	return filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// skip .git
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return nil
		}
		if rel == "pb.yaml" || rel == "" {
			// skip template manifest and root
			if d.IsDir() {
				return nil
			}
			return nil
		}
		dest := filepath.Join(dstRoot, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		// read source
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// try to treat as text template
		tpl, terr := template.New(rel).Option("missingkey=zero").Parse(string(data))
		if terr == nil {
			var buf bytes.Buffer
			if err := tpl.Execute(&buf, vars); err == nil {
				// write templated content
				if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
					return err
				}
				return os.WriteFile(dest, buf.Bytes(), 0o644)
			}
		}

		// fallback: copy raw bytes
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
}

// CollectVariables scans template files under srcRoot and returns a deduplicated list
// of variable identifiers referenced as `{{ .Var }}` in templates.
func CollectVariables(srcRoot string) ([]string, error) {
	var vars []string
	seen := map[string]bool{}
	// regex to match {{ .var }} (simple identifier)
	re := regexp.MustCompile(`{{\s*\.([a-zA-Z0-9_]+)\s*}}`)

	err := filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "pb.yaml" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		matches := re.FindAllSubmatch(data, -1)
		for _, m := range matches {
			if len(m) >= 2 {
				name := string(m[1])
				if !seen[name] {
					seen[name] = true
					vars = append(vars, name)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return vars, nil
}
