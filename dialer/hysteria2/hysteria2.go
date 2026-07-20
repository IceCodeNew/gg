// Package hysteria2 adapts outbound's Hysteria2 dialer to gg.
package hysteria2

import (
	"net"
	"net/url"
	"strconv"

	outbounddialer "github.com/daeuniverse/outbound/dialer"
	outboundhysteria2 "github.com/daeuniverse/outbound/dialer/hysteria2"
	_ "github.com/daeuniverse/outbound/protocol/hysteria2"
	"github.com/mzz2017/gg/dialer"
	"gopkg.in/yaml.v3"
)

func init() {
	dialer.FromLinkRegister("hysteria2", New)
	dialer.FromLinkRegister("hy2", New)
	dialer.FromClashRegister("hysteria2", NewFromClash)
}

// New creates a Hysteria2 dialer from a share link.
func New(link string, option *dialer.GlobalOption) (*dialer.Dialer, error) {
	extra := &outbounddialer.ExtraOption{}
	if option != nil {
		extra.AllowInsecure = option.AllowInsecure
	}
	proxyDialer, property, err := outboundhysteria2.NewHysteria2(
		extra,
		dialer.ToNetproxyDialer(dialer.SymmetricDirect),
		link,
	)
	if err != nil {
		return nil, err
	}
	exportedLink := property.Link
	if parsed, parseErr := url.Parse(exportedLink); parseErr == nil && parsed.User != nil && parsed.User.Username() == "" {
		parsed.User = nil
		exportedLink = parsed.String()
	}
	return dialer.NewDialer(
		dialer.FromNetproxyDialer(proxyDialer),
		true,
		property.Name,
		property.Protocol,
		exportedLink,
	), nil
}

// NewFromClash creates a Hysteria2 dialer from a Clash proxy entry.
func NewFromClash(node *yaml.Node, option *dialer.GlobalOption) (*dialer.Dialer, error) {
	var proxy struct {
		Name           string `yaml:"name"`
		Server         string `yaml:"server"`
		Port           int    `yaml:"port"`
		Password       string `yaml:"password"`
		SNI            string `yaml:"sni"`
		SkipCertVerify bool   `yaml:"skip-cert-verify"`
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
		Scheme:   "hysteria2",
		Host:     net.JoinHostPort(proxy.Server, strconv.Itoa(proxy.Port)),
		User:     user,
		RawQuery: query.Encode(),
		Fragment: proxy.Name,
	}
	return New(link.String(), option)
}
