package registry

// func TestBuildIndex_IncludesModuleMetadata(t *testing.T) {
// 	tmp := t.TempDir()
// 	repo := filepath.Join(tmp, "repo1")
// 	if err := os.MkdirAll(filepath.Join(repo, "modules", "foo"), 0o755); err != nil {
// 		t.Fatalf("mkdir: %v", err)
// 	}
// 	manifest := `project:
//   name: repo1
//   description: Example repo
//   tags:
//     - ex
// modules:
//   - name: foo
//     version: "0.1.0"
// `
// 	if err := os.WriteFile(filepath.Join(repo, "pb.yaml"), []byte(manifest), 0o644); err != nil {
// 		t.Fatalf("write manifest: %v", err)
// 	}
// 	// add a file in module
// 	if err := os.WriteFile(filepath.Join(repo, "modules", "foo", "a.txt"), []byte("x"), 0o644); err != nil {
// 		t.Fatalf("write module file: %v", err)
// 	}

// 	reg := &Registry{CacheDir: tmp, Sources: nil}
// 	idxPath, err := reg.BuildIndex()
// 	if err != nil {
// 		t.Fatalf("BuildIndex failed: %v", err)
// 	}
// 	if idxPath == "" {
// 		t.Fatalf("expected index path")
// 	}
// }
