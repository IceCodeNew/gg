package yaml_test

import (
	"errors"
	"testing"

	yamlv4 "go.yaml.in/yaml/v4"
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

func TestUnmarshalReportsLoadErrors(t *testing.T) {
	var value struct {
		Count int `yaml:"count"`
	}

	// Given YAML whose scalar cannot be converted to the target field type,
	// When the caller unmarshals it,
	err := yaml.Unmarshal([]byte("count: not-a-number\n"), &value)

	// Then the failure is inspectable through the maintained error type.
	if err == nil {
		t.Fatal("expected a conversion error")
	}
	var loadErrors *yamlv4.LoadErrors
	if !errors.As(err, &loadErrors) {
		t.Fatalf("errors.As(%v, *yamlv4.LoadErrors) = false", err)
	}
	if len(loadErrors.Errors) == 0 {
		t.Fatal("LoadErrors.Errors is empty")
	}
}
