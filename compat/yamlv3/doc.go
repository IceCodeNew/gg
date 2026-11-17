// Package yaml preserves the legacy YAML v3 import contract while delegating
// all parsing and encoding to the maintained go.yaml.in/yaml/v4 module.
//
// The package exports the v3 API surface without its deprecated TypeError.
// Conversion failures arrive as *go.yaml.in/yaml/v4.LoadErrors, which callers
// inspect with errors.As.
package yaml
