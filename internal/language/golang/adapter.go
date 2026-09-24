package golang

import (
	"fmt"
	"os/exec"
	"path/filepath"
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
