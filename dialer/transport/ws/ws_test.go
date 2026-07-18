package ws

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/coder/websocket"
)

type recordingDialer struct {
	mu      sync.Mutex
	network string
	address string
}

type blockingDialer struct {
	started chan struct{}
	release chan struct{}
}

func (d *blockingDialer) Dial(_, _ string) (net.Conn, error) {
	close(d.started)
	<-d.release
	client, server := net.Pipe()
	_ = server.Close()
	return client, nil
}

func (d *recordingDialer) Dial(network, address string) (net.Conn, error) {
	d.mu.Lock()
	d.network = network
	d.address = address
	d.mu.Unlock()
	return net.Dial(network, address)
}

func (d *recordingDialer) call() (network, address string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.network, d.address
}

func TestWsDial(t *testing.T) {
	serverErr := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "front.example" {
			serverErr <- fmt.Errorf("Host = %q, want front.example", r.Host)
			return
		}
		if r.URL.Path != "/tunnel" || r.URL.RawQuery != "" {
			serverErr <- fmt.Errorf("request URL = %q, want /tunnel", r.URL.RequestURI())
			return
		}

		websocketConn, err := websocket.Accept(w, r, nil)
		if err != nil {
			serverErr <- err
			return
		}
		conn := websocket.NetConn(context.Background(), websocketConn, websocket.MessageBinary)
		defer conn.Close()

		request := make([]byte, len("ping"))
		if _, err := io.ReadFull(conn, request); err != nil {
			serverErr <- err
			return
		}
		if string(request) != "ping" {
			serverErr <- fmt.Errorf("request = %q, want ping", request)
			return
		}
		_, err = conn.Write([]byte("pong"))
		serverErr <- err
	}))
	defer server.Close()

	dialer := &recordingDialer{}
	transport, err := NewWs("ws"+strings.TrimPrefix(server.URL, "http")+"?host=front.example&path=tunnel", dialer)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := transport.Dial("tcp", "unused.example:443")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

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

	network, address := dialer.call()
	if network != "tcp" || address != server.Listener.Addr().String() {
		t.Fatalf("proxy dial = %s %s, want tcp %s", network, address, server.Listener.Addr())
	}
}

func TestWsDialContextCancellation(t *testing.T) {
	dialer := &blockingDialer{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	transport, err := NewWs("ws://proxy.example/tunnel", dialer)
	if err != nil {
		t.Fatal(err)
	}
	httpTransport := transport.dialOptions.HTTPClient.Transport.(*http.Transport)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := httpTransport.DialContext(ctx, "tcp", "proxy.example:80")
		errCh <- err
	}()
	<-dialer.started
	cancel()

	if err := <-errCh; err != context.Canceled {
		t.Fatalf("DialContext() error = %v, want context.Canceled", err)
	}
	close(dialer.release)
}

func TestNewWsTLSOptions(t *testing.T) {
	transport, err := NewWs("wss://proxy.example/tunnel?sni=origin.example&allowInsecure=true", &recordingDialer{})
	if err != nil {
		t.Fatal(err)
	}

	httpTransport, ok := transport.dialOptions.HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("HTTP transport type = %T, want *http.Transport", transport.dialOptions.HTTPClient.Transport)
	}
	want := &tls.Config{ServerName: "origin.example", InsecureSkipVerify: true}
	if httpTransport.TLSClientConfig.ServerName != want.ServerName ||
		httpTransport.TLSClientConfig.InsecureSkipVerify != want.InsecureSkipVerify {
		t.Fatalf("TLS config = %#v, want %#v", httpTransport.TLSClientConfig, want)
	}
}
