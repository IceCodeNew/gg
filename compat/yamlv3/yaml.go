package yaml

import (
	"io"

	upstream "go.yaml.in/yaml/v3"
)

// This is the complete exported API of gopkg.in/yaml.v3 v3.0.1. Type aliases
// retain every exported method and field; the maintained module only adds two
// Encoder indentation methods.
type (
	Unmarshaler = upstream.Unmarshaler
	Marshaler   = upstream.Marshaler
	Decoder     = upstream.Decoder
	Encoder     = upstream.Encoder
	TypeError   = upstream.TypeError
	Kind        = upstream.Kind
	Style       = upstream.Style
	Node        = upstream.Node
	IsZeroer    = upstream.IsZeroer
)

const (
	DocumentNode = upstream.DocumentNode
	SequenceNode = upstream.SequenceNode
	MappingNode  = upstream.MappingNode
	ScalarNode   = upstream.ScalarNode
	AliasNode    = upstream.AliasNode

	TaggedStyle       = upstream.TaggedStyle
	DoubleQuotedStyle = upstream.DoubleQuotedStyle
	SingleQuotedStyle = upstream.SingleQuotedStyle
	LiteralStyle      = upstream.LiteralStyle
	FoldedStyle       = upstream.FoldedStyle
	FlowStyle         = upstream.FlowStyle
)

func Unmarshal(in []byte, out any) error {
	return upstream.Unmarshal(in, out)
}

func Marshal(in any) ([]byte, error) {
	return upstream.Marshal(in)
}

func NewDecoder(r io.Reader) *Decoder {
	return upstream.NewDecoder(r)
}

func NewEncoder(w io.Writer) *Encoder {
	return upstream.NewEncoder(w)
}
