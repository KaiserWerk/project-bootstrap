package golang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// RunTidy runs `go mod tidy` in the given project directory.
func RunTidy(projectDir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = filepath.Clean(projectDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod tidy failed: %v: %s", err, string(out))
	}
	return nil
}

// AddRequire runs `go get module@version` (or `go get module` if version empty) in the given project dir.
func AddRequire(module, version, projectDir string) error {
	// Prefer editing go.mod directly when a version is provided.
	if version != "" {
		if err := AddRequireEdit(module, version, projectDir); err == nil {
			return nil
		}
		// fallthrough to running `go get` if edit fails
	}

	arg := module
	if version != "" {
		arg = module + "@" + version
	}
	cmd := exec.Command("go", "get", arg)
	cmd.Dir = filepath.Clean(projectDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go get failed: %v: %s", err, string(out))
	}
	return nil
}

// AddRequireEdit edits the project's go.mod to add or update a require directive.
// It requires a non-empty version string (e.g. "v1.2.3").
func AddRequireEdit(modulePath, version, projectDir string) error {
	if version == "" {
		return fmt.Errorf("version required for AddRequireEdit")
	}
	goModPath := filepath.Join(projectDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	mf, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return err
	}
	// Remove existing require if present then add
	_ = mf.DropRequire(modulePath)
	if err := mf.AddRequire(modulePath, version); err != nil {
		return err
	}
	newData, err := mf.Format()
	if err != nil {
		return err
	}
	if err := os.WriteFile(goModPath, newData, 0644); err != nil {
		return err
	}
	return nil
}
