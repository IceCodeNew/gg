package socks

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"slices"
	"testing"

	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/protocol/socks/socks4"
	"github.com/sagernet/sing/protocol/socks/socks5"
)

type pipeDialer struct {
	conn net.Conn
}

func (d *pipeDialer) DialContext(context.Context, string, M.Socksaddr) (net.Conn, error) {
	return d.conn, nil
}

func (d *pipeDialer) ListenPacket(context.Context, M.Socksaddr) (net.PacketConn, error) {
	return nil, fmt.Errorf("unexpected UDP dial")
}

func TestSingDialerSOCKS5(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- serveSOCKS5(serverConn)
	}()

	d, err := newSingDialer("socks5://user:password@proxy.example:1080", &pipeDialer{conn: clientConn})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := d.Dial("tcp", "target.example:443")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	assertTunnel(t, conn, serverErr)
}

func serveSOCKS5(conn net.Conn) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	auth, err := socks5.ReadAuthRequest(reader)
	if err != nil {
		return err
	}
	if !slices.Contains(auth.Methods, socks5.AuthTypeUsernamePassword) {
		return fmt.Errorf("auth methods = %v", auth.Methods)
	}
	if err := socks5.WriteAuthResponse(conn, socks5.AuthResponse{Method: socks5.AuthTypeUsernamePassword}); err != nil {
		return err
	}
	credentials, err := socks5.ReadUsernamePasswordAuthRequest(reader)
	if err != nil {
		return err
	}
	if credentials.Username != "user" || credentials.Password != "password" {
		return fmt.Errorf("credentials = %#v", credentials)
	}
	if err := socks5.WriteUsernamePasswordAuthResponse(conn, socks5.UsernamePasswordAuthResponse{}); err != nil {
		return err
	}
	request, err := socks5.ReadRequest(reader)
	if err != nil {
		return err
	}
	if request.Command != socks5.CommandConnect || request.Destination.String() != "target.example:443" {
		return fmt.Errorf("request = %#v", request)
	}
	if err := socks5.WriteResponse(conn, socks5.Response{
		ReplyCode: socks5.ReplyCodeSuccess,
		Bind:      M.ParseSocksaddr("127.0.0.1:0"),
	}); err != nil {
		return err
	}
	return echoTunnel(conn)
}

func TestSingDialerSOCKS4A(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- serveSOCKS4A(serverConn)
	}()

	d, err := newSingDialer("socks4a://user@proxy.example:1080", &pipeDialer{conn: clientConn})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := d.Dial("tcp", "target.example:443")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	assertTunnel(t, conn, serverErr)
}

func TestSingDialerSOCKS4(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- serveSOCKS4(serverConn)
	}()

	d, err := newSingDialer("socks4://user@proxy.example:1080", &pipeDialer{conn: clientConn})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := d.Dial("tcp", "192.0.2.1:443")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	assertTunnel(t, conn, serverErr)
}

func serveSOCKS4(conn net.Conn) error {
	defer conn.Close()
	request, err := socks4.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return err
	}
	if request.Command != socks4.CommandConnect || request.Destination.String() != "192.0.2.1:443" || request.Username != "user" {
		return fmt.Errorf("request = %#v", request)
	}
	if err := socks4.WriteResponse(conn, socks4.Response{
		ReplyCode:   socks4.ReplyCodeGranted,
		Destination: M.ParseSocksaddr("127.0.0.1:0"),
	}); err != nil {
		return err
	}
	return echoTunnel(conn)
}

func serveSOCKS4A(conn net.Conn) error {
	defer conn.Close()
	request, err := socks4.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return err
	}
	if request.Command != socks4.CommandConnect || request.Destination.String() != "target.example:443" || request.Username != "user" {
		return fmt.Errorf("request = %#v", request)
	}
	if err := socks4.WriteResponse(conn, socks4.Response{
		ReplyCode:   socks4.ReplyCodeGranted,
		Destination: M.ParseSocksaddr("127.0.0.1:0"),
	}); err != nil {
		return err
	}
	return echoTunnel(conn)
}

func echoTunnel(conn net.Conn) error {
	request := make([]byte, len("ping"))
	if _, err := io.ReadFull(conn, request); err != nil {
		return err
	}
	if string(request) != "ping" {
		return fmt.Errorf("tunnel request = %q", request)
	}
	_, err := conn.Write([]byte("pong"))
	return err
}

func assertTunnel(t *testing.T, conn net.Conn, serverErr <-chan error) {
	t.Helper()
	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, len("pong"))
	if _, err := io.ReadFull(conn, response); err != nil {
		t.Fatal(err)
	}
	if string(response) != "pong" {
		t.Fatalf("response = %q, want pong", response)
	}
	if err := <-serverErr; err != nil {
		t.Fatal(err)
	}
}

func TestSocksDialerUDPFlag(t *testing.T) {
	d, err := (&Socks{
		Server:   "proxy.example",
		Port:     1080,
		Protocol: "socks5",
		UDP:      true,
	}).Dialer()
	if err != nil {
		t.Fatal(err)
	}
	if !d.SupportUDP() {
		t.Fatal("SOCKS5 UDP support is disabled")
	}
	if _, ok := d.Dialer.(*singDialer); !ok {
		t.Fatalf("dialer type = %T, want *singDialer", d.Dialer)
	}
	defaultDialer, err := (&Socks{Server: "proxy.example", Port: 1080, UDP: true}).Dialer()
	if err != nil {
		t.Fatal(err)
	}
	if !defaultDialer.SupportUDP() {
		t.Fatal("default SOCKS5 UDP support is disabled")
	}

	for _, protocol := range []string{"socks4", "socks4a"} {
		d, err := (&Socks{Server: "proxy.example", Port: 1080, Protocol: protocol, UDP: true}).Dialer()
		if err != nil {
			t.Fatal(err)
		}
		if d.SupportUDP() {
			t.Errorf("%s unexpectedly supports UDP", protocol)
		}
	}
}

func TestSocksDialerDefaultProtocolRoundTrip(t *testing.T) {
	for _, udp := range []bool{true, false} {
		t.Run(fmt.Sprintf("udp=%t", udp), func(t *testing.T) {
			want := &Socks{
				Name:     "default proxy",
				Server:   "proxy.example",
				Port:     1080,
				Username: "user",
				Password: "password",
				UDP:      udp,
			}
			d, err := want.Dialer()
			if err != nil {
				t.Fatal(err)
			}
			if d.Protocol() != "socks5" {
				t.Fatalf("protocol = %q, want socks5", d.Protocol())
			}

			got, err := ParseSocksURL(d.Link())
			if err != nil {
				t.Fatalf("parse link %q: %v", d.Link(), err)
			}
			if got.Protocol != "socks5" || got.Name != want.Name || got.Server != want.Server || got.Port != want.Port ||
				got.Username != want.Username || got.Password != want.Password || got.UDP != want.UDP {
				t.Fatalf("round trip = %#v", got)
			}
		})
	}
}
