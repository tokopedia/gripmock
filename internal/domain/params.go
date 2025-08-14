package domain

// Arguments represents the configuration for protobuf file processing.
type Arguments struct {
	protoPath []string
	imports   []string
}

// New creates a new Arguments instance with the specified protobuf paths and imports.
func New(protoPath []string, imports []string) *Arguments {
	return &Arguments{
		protoPath: protoPath,
		imports:   imports,
	}
}

// ProtoPath returns the list of protobuf file paths.
func (p *Arguments) ProtoPath() []string {
	return p.protoPath
}

// Imports returns the list of import paths for protobuf files.
func (p *Arguments) Imports() []string {
	return p.imports
}
