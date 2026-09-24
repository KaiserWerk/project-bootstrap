package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KaiserWerk/project-bootstrap/internal/project"
)

type ModuleIndex struct {
	Name    string            `json:"name"`
	Path    string            `json:"path,omitempty"`
	Version string            `json:"version,omitempty"`
	Files   map[string]string `json:"files,omitempty"`
}

type IndexEntry struct {
	Name        string        `json:"name"`
	RepoDir     string        `json:"repo_dir"`
	Manifest    string        `json:"manifest_path,omitempty"`
	Description string        `json:"description,omitempty"`
	Tags        []string      `json:"tags,omitempty"`
	Commit      string        `json:"commit,omitempty"`
	Modules     []ModuleIndex `json:"modules,omitempty"`
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
			m, err := project.Load(path)
			if err != nil {
				return nil
			}

			idx := IndexEntry{Name: m.Project.Name, RepoDir: repoDir, Manifest: path, Description: m.Project.Description, Tags: m.Project.Tags}

			// detect git commit
			if gitPath, err := exec.LookPath("git"); err == nil {
				cmd := exec.Command(gitPath, "-C", repoDir, "rev-parse", "HEAD")
				if out, err := cmd.Output(); err == nil {
					idx.Commit = strings.TrimSpace(string(out))
				}
			}

			// collect modules metadata
			for _, me := range m.Modules {
				mi := ModuleIndex{Name: me.Name, Version: me.Version}
				// if modules/<name> exists, compute file hashes
				modDir := filepath.Join(repoDir, "modules", me.Name)
				if fi, err := os.Stat(modDir); err == nil && fi.IsDir() {
					files, _ := computeHashesForDir(modDir)
					mi.Path = filepath.ToSlash(filepath.Join("modules", me.Name))
					mi.Files = files
				}
				idx.Modules = append(idx.Modules, mi)
			}

			entries = append(entries, idx)
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

// computeHashesForDir computes SHA256 hashes for files under dir, excluding pb.yaml and .git.
func computeHashesForDir(dir string) (map[string]string, error) {
	files := map[string]string{}
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "pb.yaml" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		h, err := hashFileSHA256(path)
		if err == nil {
			files[filepath.ToSlash(rel)] = h
		}
		return nil
	})
	return files, nil
}

// hashFileSHA256 returns hex-encoded SHA256 of the file at path.
func hashFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
