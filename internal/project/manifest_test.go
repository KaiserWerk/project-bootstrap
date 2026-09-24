package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManifestReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pb.yaml")

	m := &Manifest{
		Project: ProjectInfo{Name: "demo", Language: "go", Type: "web"},
		Modules: []ModuleEntry{{Name: "frontend-auth", Version: "1.0.0"}},
	}

	if err := m.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got.Project.Name != "demo" {
		t.Fatalf("expected project name demo, got %s", got.Project.Name)
	}
	if len(got.Modules) != 1 || got.Modules[0].Name != "frontend-auth" {
		t.Fatalf("modules mismatch: %#v", got.Modules)
	}

	// ensure file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("manifest file missing: %v", err)
	}
}
