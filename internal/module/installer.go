package module

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	golang "github.com/KaiserWerk/project-bootstrap/internal/language/golang"
	"github.com/KaiserWerk/project-bootstrap/internal/lockfile"
	"github.com/KaiserWerk/project-bootstrap/internal/project"
)

// InstallFromRepo copies module files from repoDir (or modules/<name>) into dstRoot,
// updates pb.yaml and pb.lock in dstRoot.
func InstallFromRepo(repoDir, moduleName, dstRoot string) error {
	// prefer modules/<moduleName> if present
	candidate := filepath.Join(repoDir, "modules", moduleName)
	src := repoDir
	if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
		src = candidate
	}

	// detect conflicts
	conflicts, err := detectConflicts(src, dstRoot)
	if err != nil {
		return fmt.Errorf("conflict detection failed: %w", err)
	}

	var backupDir string
	if len(conflicts) > 0 {
		backupDir, err = createBackup(dstRoot, conflicts)
		if err != nil {
			return fmt.Errorf("failed to backup conflicting files: %w", err)
		}
		fmt.Printf("pb: backed up %d conflicting files to %s\n", len(conflicts), backupDir)
	}

	// copy files
	if err := copyDir(src, dstRoot); err != nil {
		// attempt rollback
		if backupDir != "" {
			if rerr := restoreBackup(backupDir, dstRoot); rerr != nil {
				return fmt.Errorf("copy failed: %v; rollback failed: %v", err, rerr)
			}
		}
		return fmt.Errorf("copy failed: %w", err)
	}

	// update pb.yaml
	manifestPath := filepath.Join(dstRoot, "pb.yaml")
	var m *project.Manifest
	if _, err := os.Stat(manifestPath); err == nil {
		mm, err := project.Load(manifestPath)
		if err != nil {
			return fmt.Errorf("failed to load existing pb.yaml: %w", err)
		}
		m = mm
	} else {
		// create minimal manifest
		m = &project.Manifest{Project: project.ProjectInfo{Name: filepath.Base(dstRoot), Language: "", Type: ""}}
	}
	// append module entry if missing
	found := false
	for _, me := range m.Modules {
		if me.Name == moduleName {
			found = true
			break
		}
	}
	if !found {
		m.Modules = append(m.Modules, project.ModuleEntry{Name: moduleName})
		if err := m.Save(manifestPath); err != nil {
			return fmt.Errorf("failed to save pb.yaml: %w", err)
		}
	}

	// update pb.lock
	lockPath := filepath.Join(dstRoot, "pb.lock")
	var lf *lockfile.Lockfile
	if _, err := os.Stat(lockPath); err == nil {
		ll, err := lockfile.Load(lockPath)
		if err != nil {
			return fmt.Errorf("failed to load pb.lock: %w", err)
		}
		lf = ll
	} else {
		lf = lockfile.New()
	}
	lf.Modules[moduleName] = lockfile.ModuleLock{Version: "", Commit: ""}
	if err := lf.Save(lockPath); err != nil {
		return fmt.Errorf("failed to save pb.lock: %w", err)
	}

	// If the module source contains a go.mod, try to add a require to the project's go.mod.
	goModPath := filepath.Join(src, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// read module path from go.mod
		f, err := os.Open(goModPath)
		if err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "module ") {
					modPath := strings.TrimSpace(strings.TrimPrefix(line, "module"))
					modPath = strings.TrimSpace(modPath)
					version := ""
					// if manifest had a version for this module, use it
					for _, me := range m.Modules {
						if me.Name == moduleName && me.Version != "" {
							version = me.Version
							break
						}
					}
					if err := golang.AddRequire(modPath, version, dstRoot); err != nil {
						// non-fatal: warn and continue
						fmt.Fprintf(os.Stderr, "pb: warning: failed to add go require %s: %v\n", modPath, err)
					}
					break
				}
			}
		}
	}

	return nil
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if e.Name() == ".git" {
				continue
			}
			if err := copyDir(s, d); err != nil {
				return err
			}
			continue
		}
		// skip manifest files
		if e.Name() == "pb.yaml" {
			continue
		}
		if err := copyFile(s, d); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// detectConflicts returns a list of relative paths (from src) that would collide in dstRoot.
func detectConflicts(src, dstRoot string) ([]string, error) {
	var conflicts []string
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
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
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return nil
		}
		target := filepath.Join(dstRoot, rel)
		if _, err := os.Stat(target); err == nil {
			conflicts = append(conflicts, rel)
		}
		return nil
	})
	return conflicts, err
}

// createBackup moves conflicting files into a timestamped backup directory under dstRoot/.pb/backups.
func createBackup(dstRoot string, conflicts []string) (string, error) {
	ts := time.Now().UTC().Format("20060102T150405Z")
	backupDir := filepath.Join(dstRoot, ".pb", "backups", ts)
	for _, rel := range conflicts {
		srcPath := filepath.Join(dstRoot, rel)
		destPath := filepath.Join(backupDir, rel)
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return "", err
		}
		if err := os.Rename(srcPath, destPath); err != nil {
			return "", err
		}
	}
	return backupDir, nil
}

// restoreBackup moves files from backupDir back into dstRoot (best-effort).
func restoreBackup(backupDir, dstRoot string) error {
	return filepath.WalkDir(backupDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(backupDir, path)
		if err != nil {
			return nil
		}
		dest := filepath.Join(dstRoot, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := os.Rename(path, dest); err != nil {
			return err
		}
		return nil
	})
}
