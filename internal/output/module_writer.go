package output

import "github.com/KaiserWerk/project-bootstrap/internal/types"

type ModuleWriter interface {
	WriteModule(modules []types.Module) error
	WriteRegistryModule(modules []types.ObjectMetadata) error
}
