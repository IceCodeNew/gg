package yaml_test

import (
	"io"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type apiValue struct{}

func (*apiValue) MarshalYAML() (any, error)      { return nil, nil }
func (*apiValue) UnmarshalYAML(*yaml.Node) error { return nil }
func (*apiValue) IsZero() bool                   { return true }

var (
	_ yaml.Marshaler   = (*apiValue)(nil)
	_ yaml.Unmarshaler = (*apiValue)(nil)
	_ yaml.IsZeroer    = (*apiValue)(nil)
	_ error            = (*yaml.TypeError)(nil)

	_ func(any) ([]byte, error)     = yaml.Marshal
	_ func([]byte, any) error       = yaml.Unmarshal
	_ func(io.Reader) *yaml.Decoder = yaml.NewDecoder
	_ func(io.Writer) *yaml.Encoder = yaml.NewEncoder

	_ yaml.Kind = yaml.DocumentNode
	_ yaml.Kind = yaml.SequenceNode
	_ yaml.Kind = yaml.MappingNode
	_ yaml.Kind = yaml.ScalarNode
	_ yaml.Kind = yaml.AliasNode

	_ yaml.Style = yaml.TaggedStyle
	_ yaml.Style = yaml.DoubleQuotedStyle
	_ yaml.Style = yaml.SingleQuotedStyle
	_ yaml.Style = yaml.LiteralStyle
	_ yaml.Style = yaml.FoldedStyle
	_ yaml.Style = yaml.FlowStyle
)

func assertAPISurface(decoder *yaml.Decoder, encoder *yaml.Encoder, node *yaml.Node, typeError *yaml.TypeError) {
	_ = decoder.Decode
	_ = decoder.KnownFields
	_ = encoder.Close
	_ = encoder.Encode
	_ = encoder.SetIndent
	_ = node.Decode
	_ = node.Encode
	_ = node.IsZero
	_ = node.LongTag
	_ = node.SetString
	_ = node.ShortTag
	_ = typeError.Error

	_ = yaml.Node{
		Kind:        yaml.ScalarNode,
		Style:       yaml.TaggedStyle,
		Tag:         "tag",
		Value:       "value",
		Anchor:      "anchor",
		Alias:       node,
		Content:     []*yaml.Node{node},
		HeadComment: "head",
		LineComment: "line",
		FootComment: "foot",
		Line:        1,
		Column:      1,
	}
	_ = yaml.TypeError{Errors: []string{"error"}}
}

func TestLegacyV3APISurface(*testing.T) {
	assertAPISurface(
		yaml.NewDecoder(strings.NewReader("")),
		yaml.NewEncoder(io.Discard),
		&yaml.Node{},
		&yaml.TypeError{},
	)
}
