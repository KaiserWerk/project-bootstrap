package golang

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddRequireEdit_WritesRequire(t *testing.T) {
	dir := t.TempDir()
	// create minimal go.mod
	goMod := "module example.com/test\n\nrequire (\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	if err := AddRequireEdit("rsc.io/quote", "v1.5.0", dir); err != nil {
		t.Fatalf("AddRequireEdit failed: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "rsc.io/quote v1.5.0") {
		t.Fatalf("go.mod did not contain require: %s", s)
	}
}
