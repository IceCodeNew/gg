// Package yaml preserves the legacy YAML v3 import contract while delegating
// all parsing and encoding to the maintained go.yaml.in/yaml/v4 module.
//
// Conversion failures arrive as *go.yaml.in/yaml/v4.LoadErrors. Callers that
// inspect them through errors.As with *yaml.TypeError keep working; a direct
// type assertion to *yaml.TypeError does not.
package yaml
