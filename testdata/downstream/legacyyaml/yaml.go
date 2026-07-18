package yaml

type Node struct{}

func (*Node) Decode(any) error {
	return nil
}
