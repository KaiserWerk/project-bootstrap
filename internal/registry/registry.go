package registry

import (
	"context"
	"path"

	"github.com/KaiserWerk/project-bootstrap/internal/gittools"
)

func DownloadRegistry(source, workdir string) ([]byte, error) {
	sourceName := path.Base(source)
	if err := gittools.Clone(context.Background(), source, workdir, sourceName); err != nil {
		return nil, err
	}

	return nil, nil
}
