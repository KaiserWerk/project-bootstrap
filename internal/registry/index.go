package registry

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type IndexEntry struct {
	Name     string `json:"name"`
	RepoDir  string `json:"repo_dir"`
	Manifest string `json:"manifest_path,omitempty"`
}

// BuildIndex scans cached repositories for `pb.yaml` manifests and writes an index.json.
func (r *Registry) BuildIndex() (string, error) {
	if _, err := os.Stat(r.CacheDir); os.IsNotExist(err) {
		return "", fmt.Errorf("cache dir not found: %s", r.CacheDir)
	}

	var entries []IndexEntry
	filepath.WalkDir(r.CacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && d.Name() == "pb.yaml" {
			repoDir := filepath.Dir(path)
			e := IndexEntry{Name: filepath.Base(repoDir), RepoDir: repoDir, Manifest: path}
			entries = append(entries, e)
		}
		return nil
	})

	if len(entries) == 0 {
		return "", fmt.Errorf("no entries found to index")
	}

	idxPath := filepath.Join(r.CacheDir, "index.json")
	f, err := os.Create(idxPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		return "", err
	}
	return idxPath, nil
}

// LoadIndex returns index entries if index.json exists.
func (r *Registry) LoadIndex() ([]IndexEntry, error) {
	idxPath := filepath.Join(r.CacheDir, "index.json")
	data, err := os.ReadFile(idxPath)
	if err != nil {
		return nil, err
	}
	var entries []IndexEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
