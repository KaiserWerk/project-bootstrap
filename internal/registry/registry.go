package registry

import (
	"context"
	"os"
	"path/filepath"

	"github.com/KaiserWerk/project-bootstrap/internal/gittools"
)

func DownloadRegistry(source, sourceDir, sourceName string) error {
	// just make sure the source directory exists
	_ = os.MkdirAll(sourceDir, 0o755)

	return gittools.Clone(context.Background(), source, sourceDir, sourceName)
}

func UpdateRegistry(source, sourceDir, sourceName string) error {
	return gittools.Pull(context.Background(), filepath.Join(sourceDir, sourceName))
}
