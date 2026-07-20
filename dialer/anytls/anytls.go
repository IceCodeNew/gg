// Package anytls adapts outbound's AnyTLS share-link dialer to gg.
package anytls

import (
	outbounddialer "github.com/daeuniverse/outbound/dialer"
	outboundanytls "github.com/daeuniverse/outbound/dialer/anytls"
	_ "github.com/daeuniverse/outbound/protocol/anytls"
	"github.com/mzz2017/gg/dialer"
)

func init() {
	dialer.FromLinkRegister("anytls", New)
}

// New creates an AnyTLS dialer from a share link.
func New(link string, option *dialer.GlobalOption) (*dialer.Dialer, error) {
	extraOption := &outbounddialer.ExtraOption{}
	if option != nil {
		extraOption.AllowInsecure = option.AllowInsecure
	}

	proxyDialer, property, err := outboundanytls.NewAnytls(
		extraOption,
		dialer.ToNetproxyDialer(dialer.SymmetricDirect),
		link,
	)
	if err != nil {
		return nil, err
	}
	return dialer.NewDialer(
		dialer.FromNetproxyDialer(proxyDialer),
		true,
		property.Name,
		property.Protocol,
		property.Link,
	), nil
}
