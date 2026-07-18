package socks

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/mzz2017/gg/dialer"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	singSocks "github.com/sagernet/sing/protocol/socks"
	"golang.org/x/net/proxy"
	"gopkg.in/yaml.v3"
)

func init() {
	dialer.FromLinkRegister("socks", NewSocks) // socks -> socks5
	dialer.FromLinkRegister("socks4", NewSocks)
	dialer.FromLinkRegister("socks4a", NewSocks)
	dialer.FromLinkRegister("socks5", NewSocks)
	dialer.FromClashRegister("socks5", NewSocks5FromClashObj)
}

type Socks struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Protocol string `json:"protocol"`
	UDP      bool   `json:"udp"`
}

type singDialer struct {
	client *singSocks.Client
}

func (d *singDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *singDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.client.DialContext(ctx, network, M.ParseSocksaddr(address))
}

func newSingDialer(rawURL string, next N.Dialer) (proxy.Dialer, error) {
	client, err := singSocks.NewClientFromURL(next, rawURL)
	if err != nil {
		return nil, err
	}
	return &singDialer{client: client}, nil
}

func NewSocks(link string, opt *dialer.GlobalOption) (*dialer.Dialer, error) {
	s, err := ParseSocksURL(link)
	if err != nil {
		return nil, dialer.InvalidParameterErr
	}
	return s.Dialer()
}

func NewSocks5FromClashObj(o *yaml.Node, opt *dialer.GlobalOption) (*dialer.Dialer, error) {
	s, err := ParseClashSocks5(o)
	if err != nil {
		return nil, err
	}
	return s.Dialer()
}

func (s *Socks) Dialer() (*dialer.Dialer, error) {
	protocol := s.Protocol
	link := s.ExportToURL()
	if protocol == "" {
		protocol = "socks5"
		canonical := *s
		canonical.Protocol = protocol
		link = canonical.ExportToURL()
	}
	supportUDP := s.UDP
	switch protocol {
	case "socks", "socks4", "socks4a", "socks5":
		if protocol == "socks4" || protocol == "socks4a" {
			supportUDP = false
		}
	default:
		return nil, fmt.Errorf("unexpected protocol: %v", protocol)
	}
	d, err := newSingDialer(link, N.SystemDialer)
	if err != nil {
		return nil, err
	}
	return dialer.NewDialer(d, supportUDP, s.Name, protocol, link), nil
}

func ParseClashSocks5(o *yaml.Node) (data *Socks, err error) {
	type Socks5Option struct {
		Name           string `yaml:"name"`
		Server         string `yaml:"server"`
		Port           int    `yaml:"port"`
		UserName       string `yaml:"username,omitempty"`
		Password       string `yaml:"password,omitempty"`
		TLS            bool   `yaml:"tls,omitempty"`
		UDP            bool   `yaml:"udp,omitempty"`
		SkipCertVerify bool   `yaml:"skip-cert-verify,omitempty"`
	}
	var option Socks5Option
	if err = o.Decode(&option); err != nil {
		return nil, err
	}
	if option.TLS {
		return nil, fmt.Errorf("%w: tls=true", dialer.UnexpectedFieldErr)
	}
	if option.SkipCertVerify {
		return nil, fmt.Errorf("%w: skip-cert-verify=true", dialer.UnexpectedFieldErr)
	}
	return &Socks{
		Name:     option.Name,
		Server:   option.Server,
		Port:     option.Port,
		Username: option.UserName,
		Password: option.Password,
		Protocol: "socks5",
		UDP:      option.UDP,
	}, nil
}

func ParseSocksURL(link string) (data *Socks, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, dialer.InvalidParameterErr
	}
	pwd, _ := u.User.Password()
	strPort := u.Port()
	port, err := strconv.Atoi(strPort)
	if err != nil {
		return nil, err
	}
	// socks -> socks5
	if u.Scheme == "socks" {
		u.Scheme = "socks5"
	}
	udp := u.Scheme == "socks5"
	if udp && u.Query().Get("udp") == "false" {
		udp = false
	}
	return &Socks{
		Name:     u.Fragment,
		Server:   u.Hostname(),
		Port:     port,
		Username: u.User.Username(),
		Password: pwd,
		Protocol: u.Scheme,
		UDP:      udp,
	}, nil
}

func (s *Socks) ExportToURL() string {
	var user *url.Userinfo
	if s.Password != "" {
		user = url.UserPassword(s.Username, s.Password)
	} else {
		user = url.User(s.Username)
	}
	query := url.Values{}
	if (s.Protocol == "socks" || s.Protocol == "socks5") && !s.UDP {
		query.Set("udp", "false")
	}
	u := url.URL{
		Scheme:   s.Protocol,
		User:     user,
		Host:     net.JoinHostPort(s.Server, strconv.Itoa(s.Port)),
		RawQuery: query.Encode(),
		Fragment: s.Name,
	}
	return u.String()
}
