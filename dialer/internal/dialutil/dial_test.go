package dialutil

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

type dialerFunc func(string, string) (net.Conn, error)

func (f dialerFunc) Dial(network, address string) (net.Conn, error) {
	return f(network, address)
}

type contextDialerFunc func(context.Context, string, string) (net.Conn, error)

func (f contextDialerFunc) Dial(string, string) (net.Conn, error) {
	panic("unexpected Dial call")
}

func (f contextDialerFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

type closeTrackingConn struct {
	net.Conn
	closed chan struct{}
	once   sync.Once
}

func (c *closeTrackingConn) Close() error {
	c.once.Do(func() { close(c.closed) })
	return c.Conn.Close()
}

func TestDialContextUsesNativeImplementation(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "value")
	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	got, err := DialContext(ctx, contextDialerFunc(func(gotCtx context.Context, network, address string) (net.Conn, error) {
		if gotCtx.Value(contextKey{}) != "value" || network != "tcp" || address != "example.com:443" {
			t.Fatalf("DialContext(%v, %q, %q)", gotCtx, network, address)
		}
		return client, nil
	}), "tcp", "example.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if got != client {
		t.Fatalf("connection = %T, want original pipe connection", got)
	}
}

func TestDialContextClosesConnectionAfterCancellation(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	client, server := net.Pipe()
	t.Cleanup(func() { _ = server.Close() })
	tracked := &closeTrackingConn{Conn: client, closed: make(chan struct{})}
	d := dialerFunc(func(string, string) (net.Conn, error) {
		close(started)
		<-release
		return tracked, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, err := DialContext(ctx, d, "tcp", "example.com:443")
		result <- err
	}()
	<-started
	cancel()
	close(release)
	if err := <-result; err != context.Canceled {
		t.Fatalf("error = %v, want context.Canceled", err)
	}

	select {
	case <-tracked.closed:
	case <-time.After(time.Second):
		t.Fatal("connection was not closed after cancellation")
	}
}

func TestDialContextAlreadyCanceledDoesNotDial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	_, err := DialContext(ctx, dialerFunc(func(string, string) (net.Conn, error) {
		called = true
		return nil, nil
	}), "tcp", "example.com:443")
	if err != context.Canceled {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if called {
		t.Fatal("Dial called with an already canceled context")
	}
}
