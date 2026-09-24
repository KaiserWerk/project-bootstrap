package registry

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KaiserWerk/project-bootstrap/internal/project"
)

type Registry struct {
	CacheDir string
	Sources  []string
}

type Entry struct {
	Name     string
	RepoDir  string
	Manifest *project.Manifest
}

func New(sources []string) *Registry {
	home, _ := os.UserHomeDir()
	cache := filepath.Join(home, ".pb", "registry")
	return &Registry{CacheDir: cache, Sources: sources}
}

func (r *Registry) ensureRepoDir(source string) string {
	s := strings.ReplaceAll(source, "/", "_")
	s = strings.ReplaceAll(s, ":", "_")
	return filepath.Join(r.CacheDir, s)
}

func (r *Registry) EnsureRepo(source string) (string, error) {
	dir := r.ensureRepoDir(source)
	if _, err := os.Stat(dir); err == nil {
		cmd := exec.Command("git", "-C", dir, "pull", "--ff-only")
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "pb: git pull failed for %s: %s\n", source, string(out))
		}
		return dir, nil
	}

	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return "", err
	}

	srcURL := source
	if !strings.HasPrefix(source, "http://") && !strings.HasPrefix(source, "https://") && !strings.HasPrefix(source, "git@") {
		srcURL = "https://" + source
	}
	cmd := exec.Command("git", "clone", "--depth", "1", srcURL, dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git clone failed: %v: %s", err, string(out))
	}
	return dir, nil
}

func (r *Registry) UpdateAll() error {
	for _, src := range r.Sources {
		if _, err := r.EnsureRepo(src); err != nil {
			return err
		}
	}
	return nil
}

// Search scans cached repositories for manifests that match the query and returns entries.
func (r *Registry) Search(query string) ([]Entry, error) {
	if err := r.UpdateAll(); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed updating repos: %v\n", err)
	}

	var results []Entry
	// try using index.json for faster searches
	if idxEntries, err := r.LoadIndex(); err == nil {
		ql := strings.ToLower(query)
		for _, ie := range idxEntries {
			match := false
			if query == "" || strings.Contains(strings.ToLower(ie.Name), ql) {
				match = true
			}
			if !match && ie.Description != "" {
				if strings.Contains(strings.ToLower(ie.Description), ql) {
					match = true
				}
			}
			if !match && len(ie.Tags) > 0 {
				for _, t := range ie.Tags {
					if strings.Contains(strings.ToLower(t), ql) {
						match = true
						break
					}
				}
			}
			if !match && len(ie.Modules) > 0 {
				for _, m := range ie.Modules {
					if strings.Contains(strings.ToLower(m.Name), ql) {
						match = true
						break
					}
				}
			}
			if match {
				var manifest *project.Manifest
				if ie.Manifest != "" {
					if mm, err := project.Load(ie.Manifest); err == nil {
						manifest = mm
					}
				}
				results = append(results, Entry{Name: ie.Name, RepoDir: ie.RepoDir, Manifest: manifest})
			}
		}
		return results, nil
	}

	if _, err := os.Stat(r.CacheDir); os.IsNotExist(err) {
		return results, nil
	}

	filepath.WalkDir(r.CacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && d.Name() == "pb.yaml" {
			m, err := project.Load(path)
			if err == nil {
				match := false
				if query == "" || strings.Contains(m.Project.Name, query) {
					match = true
				}
				if !match {
					for _, me := range m.Modules {
						if strings.Contains(me.Name, query) {
							match = true
							break
						}
					}
				}
				if !match && m.Project.Description != "" {
					if strings.Contains(strings.ToLower(m.Project.Description), strings.ToLower(query)) {
						match = true
					}
				}
				if !match && len(m.Project.Tags) > 0 {
					ql := strings.ToLower(query)
					for _, t := range m.Project.Tags {
						if strings.Contains(strings.ToLower(t), ql) {
							match = true
							break
						}
					}
				}
				if match {
					// repo dir is two levels up from pb.yaml in our layout
					repoDir := filepath.Dir(path)
					results = append(results, Entry{Name: m.Project.Name, RepoDir: repoDir, Manifest: m})
				}
			}
		}
		return nil
	})

	// fallback: look for directories that partially match
	if len(results) == 0 && query != "" {
		entries, _ := os.ReadDir(r.CacheDir)
		for _, e := range entries {
			if e.IsDir() && strings.Contains(e.Name(), query) {
				results = append(results, Entry{Name: e.Name(), RepoDir: filepath.Join(r.CacheDir, e.Name()), Manifest: nil})
			}
		}
	}
	return results, nil
}

// FindModule searches for a module by name and returns the first matching entry.
func (r *Registry) FindModule(name string) (*Entry, error) {
	if err := r.UpdateAll(); err != nil {
		fmt.Fprintf(os.Stderr, "pb: failed updating repos: %v\n", err)
	}
	if _, err := os.Stat(r.CacheDir); os.IsNotExist(err) {
		return nil, errors.New("registry cache not found")
	}

	var found *Entry
	filepath.WalkDir(r.CacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && d.Name() == "pb.yaml" {
			m, err := project.Load(path)
			if err == nil {
				if m.Project.Name == name {
					repoDir := filepath.Dir(path)
					e := Entry{Name: m.Project.Name, RepoDir: repoDir, Manifest: m}
					found = &e
					return fs.SkipDir
				}
				for _, me := range m.Modules {
					if me.Name == name {
						repoDir := filepath.Dir(path)
						e := Entry{Name: me.Name, RepoDir: repoDir, Manifest: m}
						found = &e
						return fs.SkipDir
					}
				}
			}
		}
		return nil
	})

	if found == nil {
		return nil, fmt.Errorf("module %s not found", name)
	}
	return found, nil
}
