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
	if len(modules) == 0 {
		_, err := fmt.Fprintln(w.w, "\tNo modules found.")
		if err != nil {
			return err
		}
		return nil
	}
	for _, module := range modules {
		_, err := fmt.Fprintf(w.w, "\tName: %s\nVersion: %s\nDescription: %s\n\n", module.Name, module.Version, module.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *TextModuleWriter) WriteRegistryModule(modules []types.ObjectMetadata) error {
	if len(modules) == 0 {
		_, err := fmt.Fprintln(w.w, "  No modules in registry found.")
		if err != nil {
			return err
		}
		return nil
	}
	for _, module := range modules {
		_, err := fmt.Fprintf(w.w, "\nName: %s\nDescription: %s\nVersions:\n", module.Name, module.Description)
		if err != nil {
			return err
		}

		for _, version := range module.Versions {
			_, err := fmt.Fprintf(w.w, "    - %s\n", version)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(w.w)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ ModuleWriter = (*TextModuleWriter)(nil)
var DefaultTextModuleWriter = &TextModuleWriter{w: os.Stdout}
