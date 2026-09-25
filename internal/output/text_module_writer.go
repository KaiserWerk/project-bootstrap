package output

import (
	"fmt"
	"io"
	"os"

	"github.com/KaiserWerk/project-bootstrap/internal/types"
)

type TextModuleWriter struct {
	w io.Writer
}

func (w *TextModuleWriter) WriteModule(modules []types.Module) error {
	for _, module := range modules {
		_, err := fmt.Fprintf(w.w, "\tName: %s\nVersion: %s\nDescription: %s\n\n", module.Name, module.Version, module.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ ModuleWriter = (*TextModuleWriter)(nil)
var DefaultTextModuleWriter = &TextModuleWriter{w: os.Stdout}
