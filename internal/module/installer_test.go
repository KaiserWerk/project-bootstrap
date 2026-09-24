package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KaiserWerk/project-bootstrap/internal/lockfile"
	"github.com/KaiserWerk/project-bootstrap/internal/project"
	
	"gopkg.in/yaml.v3"
)

func TestInstallFromRepo_CopiesFilesAndUpdatesManifests(t *testing.T) {
	repoDir := t.TempDir()
	// create a fake module under modules/foo
	modName := "foo"
	modDir := filepath.Join(repoDir, "modules", modName)
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// add a file
	samplePath := filepath.Join(modDir, "hello.txt")
	if err := os.WriteFile(samplePath, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// create a template manifest in repo root (should be ignored)
	m := &project.Manifest{Project: project.ProjectInfo{Name: "tmpl"}}
	mm, _ := yaml.Marshal(m)
	if err := os.WriteFile(filepath.Join(repoDir, "pb.yaml"), mm, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// destination project
	dst := t.TempDir()

	if err := InstallFromRepo(repoDir, modName, dst); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// file should be copied
	got, err := os.ReadFile(filepath.Join(dst, "hello.txt"))
	if err != nil {
		t.Fatalf("expected copied file: %v", err)
	}
	if string(got) != "hello world" {
		t.Fatalf("unexpected file contents: %s", string(got))
	}

	// pb.yaml should exist and include module entry
	b, err := os.ReadFile(filepath.Join(dst, "pb.yaml"))
	if err != nil {
		t.Fatalf("missing pb.yaml: %v", err)
	}
	var gotM project.Manifest
	if err := yaml.Unmarshal(b, &gotM); err != nil {
		t.Fatalf("unmarshal pb.yaml: %v", err)
	}
	found := false
	for _, me := range gotM.Modules {
		if me.Name == modName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("module entry not added to pb.yaml: %#v", gotM.Modules)
	}

	// pb.lock should exist and include module
	lb, err := os.ReadFile(filepath.Join(dst, "pb.lock"))
	if err != nil {
		t.Fatalf("missing pb.lock: %v", err)
	}
	var lock lockfile.Lockfile
	if err := yaml.Unmarshal(lb, &lock); err != nil {
		t.Fatalf("unmarshal pb.lock: %v", err)
	}
	if _, ok := lock.Modules[modName]; !ok {
		t.Fatalf("module not recorded in lockfile: %#v", lock.Modules)
	}
}
