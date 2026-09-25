package output

import (
	"encoding/json"
	"io"
	"os"

	"github.com/KaiserWerk/project-bootstrap/internal/types"
)

type JSONModuleWriter struct {
	w io.Writer
}

func (w *JSONModuleWriter) WriteModule(modules []types.Module) error {
	data, err := json.MarshalIndent(modules, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.w.Write(data)
	return err
}

var _ ModuleWriter = (*JSONModuleWriter)(nil)
var DefaultJSONModuleWriter = &JSONModuleWriter{w: os.Stdout}
