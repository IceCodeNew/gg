package downstream

import (
	"github.com/mzz2017/gg/dialer"
	"gopkg.in/yaml.v3"
)

func legacyCreator(*yaml.Node, *dialer.GlobalOption) (*dialer.Dialer, error) {
	return nil, nil
}

var _ dialer.FromClashCreator = legacyCreator
