package yaml_test

import (
	"errors"
	"testing"

	"gopkg.in/yaml.v3"
)

type nodeValue string

func (v *nodeValue) UnmarshalYAML(node *yaml.Node) error {
	*v = nodeValue(node.Value)
	return nil
}

func TestUnmarshalUsesMaintainedNodeType(t *testing.T) {
	var value nodeValue
	if err := yaml.Unmarshal([]byte("value\n"), &value); err != nil {
		t.Fatal(err)
	}
	if value != "value" {
		t.Fatalf("value = %q, want value", value)
	}
}

func TestUnmarshalReportsLegacyTypeError(t *testing.T) {
	var value struct {
		Count int `yaml:"count"`
	}

	// Given YAML whose scalar cannot be converted to the target field type,
	// When the caller unmarshals it,
	err := yaml.Unmarshal([]byte("count: not-a-number\n"), &value)

	// Then the conversion failure must still be inspectable as a v3 TypeError.
	if err == nil {
		t.Fatal("expected a conversion error")
	}
	var typeErr *yaml.TypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("errors.As(%v, *yaml.TypeError) = false", err)
	}
	if len(typeErr.Errors) == 0 {
		t.Fatal("TypeError.Errors is empty")
	}
}
