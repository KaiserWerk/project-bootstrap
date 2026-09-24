package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearch_MatchesNameTagsDescription(t *testing.T) {
	tmp := t.TempDir()
	// create a fake repo with pb.yaml
	repo := filepath.Join(tmp, "repo1")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	manifest := `project:
  name: auth-lib
  description: "A small auth library"
  tags:
    - auth
    - security
modules:
  - name: frontend-auth
    version: "1.0.0"
`
	if err := os.WriteFile(filepath.Join(repo, "pb.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	reg := &Registry{CacheDir: tmp, Sources: nil}

	// search by name fragment
	res, err := reg.Search("auth")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("expected results for 'auth', got none")
	}

	// search by tag
	res2, err := reg.Search("security")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(res2) == 0 {
		t.Fatalf("expected results for 'security', got none")
	}

	// search by description phrase
	res3, err := reg.Search("small auth")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(res3) == 0 {
		t.Fatalf("expected results for 'small auth', got none")
	}

}
