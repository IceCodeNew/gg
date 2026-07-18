package ws

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/coder/websocket"
	"github.com/mzz2017/gg/common"
	"github.com/mzz2017/gg/dialer/internal/dialutil"
	"golang.org/x/net/proxy"
)

// Ws is a base Ws struct
type Ws struct {
	wsAddr      string
	dialOptions websocket.DialOptions
}

// NewWs returns a Ws infra.
func NewWs(s string, d proxy.Dialer) (*Ws, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("NewWs: %w", err)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialutil.DialContext(ctx, d, network, address)
		},
	}

	query := u.Query()
	host := query.Get("host")
	if host == "" {
		host = u.Hostname()
	}
	path := u.Path
	if path == "" {
		path = query.Get("path")
	}
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	wsUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path,
	}
	t := &Ws{
		wsAddr: wsUrl.String(),
		dialOptions: websocket.DialOptions{
			HTTPClient: &http.Client{Transport: transport},
			Host:       host,
		},
	}
	if u.Scheme == "wss" {
		skipVerify := common.StringToBool(query.Get("allowInsecure")) ||
			common.StringToBool(query.Get("skipVerify"))
		transport.TLSClientConfig = &tls.Config{
			ServerName:         u.Query().Get("sni"),
			InsecureSkipVerify: skipVerify,
		}
	}
	return t, nil
}

// Dial connects to the address addr on the network net via the infra.
func (s *Ws) Dial(network, addr string) (net.Conn, error) {
	rc, _, err := websocket.Dial(context.Background(), s.wsAddr, &s.dialOptions)
	if err != nil {
		return nil, fmt.Errorf("[Ws]: dial to %s: %w", s.wsAddr, err)
	}
	return websocket.NetConn(context.Background(), rc, websocket.MessageBinary), nil
}
