package module

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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

	// copy files
	if err := copyDir(src, dstRoot); err != nil {
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
