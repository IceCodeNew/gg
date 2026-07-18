package yaml_test

import (
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
