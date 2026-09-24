package lockfile

import (
	"path/filepath"
	"testing"
)

func TestLockfileReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pb.lock")

	l := New()
	l.Modules["frontend-auth"] = ModuleLock{Version: "1.0.0", Commit: "abc123"}

	if err := l.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	m, ok := got.Modules["frontend-auth"]
	if !ok {
		t.Fatalf("module not found in lockfile")
	}
	if m.Version != "1.0.0" || m.Commit != "abc123" {
		t.Fatalf("unexpected module values: %#v", m)
	}
}
