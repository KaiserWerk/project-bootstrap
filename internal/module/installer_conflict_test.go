package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KaiserWerk/project-bootstrap/internal/lockfile"
)

func TestInstallFromRepo_BackupsConflicts(t *testing.T) {
	repo := t.TempDir()
	modDir := filepath.Join(repo, "modules", "foo")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// module file
	if err := os.WriteFile(filepath.Join(modDir, "a.txt"), []byte("new"), 0o644); err != nil {
		t.Fatalf("write module file: %v", err)
	}

	dst := t.TempDir()
	// create conflicting file in dst
	if err := os.WriteFile(filepath.Join(dst, "a.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write dst file: %v", err)
	}

	if err := InstallFromRepo(repo, "foo", dst); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// original should be moved to backup
	backupsDir := filepath.Join(dst, ".pb", "backups")
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("missing backups dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected at least one backup entry, got 0")
	}

	// new file should be present and contain "new"
	b, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatalf("missing new file: %v", err)
	}
	if string(b) != "new" {
		t.Fatalf("unexpected contents: %s", string(b))
	}

	// pb.lock should exist and include file hashes for module
	lf, err := lockfile.Load(filepath.Join(dst, "pb.lock"))
	if err != nil {
		t.Fatalf("failed to load pb.lock: %v", err)
	}
	ml, ok := lf.Modules["foo"]
	if !ok {
		t.Fatalf("pb.lock missing module entry")
	}
	if len(ml.Files) == 0 {
		t.Fatalf("expected file hashes in pb.lock, got none")
	}
}
