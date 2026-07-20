// Package anytls adapts outbound's AnyTLS share-link dialer to gg.
package anytls

import (
	"net"
	"net/url"
	"strconv"

	outbounddialer "github.com/daeuniverse/outbound/dialer"
	outboundanytls "github.com/daeuniverse/outbound/dialer/anytls"
	_ "github.com/daeuniverse/outbound/protocol/anytls"
	"github.com/mzz2017/gg/dialer"
	"gopkg.in/yaml.v3"
)

func init() {
	dialer.FromLinkRegister("anytls", New)
	dialer.FromClashRegister("anytls", NewFromClash)
}

// NewFromClash creates an AnyTLS dialer from a Clash proxy entry.
func NewFromClash(node *yaml.Node, option *dialer.GlobalOption) (*dialer.Dialer, error) {
	var proxy struct {
		Name           string `yaml:"name"`
		Server         string `yaml:"server"`
		Port           int    `yaml:"port"`
		Password       string `yaml:"password"`
		SNI            string `yaml:"sni"`
		SkipCertVerify bool   `yaml:"skip-cert-verify"`
		UDP            *bool  `yaml:"udp,omitempty"`
	}
	if err := node.Decode(&proxy); err != nil {
		return nil, err
	}
	query := url.Values{}
	if proxy.SNI != "" {
		query.Set("sni", proxy.SNI)
	}
	if proxy.SkipCertVerify {
		query.Set("insecure", "1")
	}
	var user *url.Userinfo
	if proxy.Password != "" {
		user = url.User(proxy.Password)
	}
	link := url.URL{
		Scheme:   "anytls",
		Host:     net.JoinHostPort(proxy.Server, strconv.Itoa(proxy.Port)),
		User:     user,
		RawQuery: query.Encode(),
		Fragment: proxy.Name,
	}
	result, err := New(link.String(), option)
	if err != nil {
		return nil, err
	}
	supportUDP := true
	if proxy.UDP != nil {
		supportUDP = *proxy.UDP
	}
	if supportUDP == result.SupportUDP() {
		return result, nil
	}
	return dialer.NewDialer(result.Dialer, supportUDP, result.Name(), result.Protocol(), result.Link()), nil
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
