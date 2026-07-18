// Package dialutil contains shared adapters for proxy dialers.
package dialutil

import (
	"context"
	"net"

	"golang.org/x/net/proxy"
)

// DialContext uses a dialer's native context support when available and
// otherwise closes any connection that arrives after cancellation.
func DialContext(ctx context.Context, d proxy.Dialer, network, address string) (net.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if d, ok := d.(interface {
		DialContext(context.Context, string, string) (net.Conn, error)
	}); ok {
		return d.DialContext(ctx, network, address)
	}

	type result struct {
		conn net.Conn
		err  error
	}
	resultCh := make(chan result)
	go func() {
		conn, err := d.Dial(network, address)
		select {
		case resultCh <- result{conn: conn, err: err}:
		case <-ctx.Done():
			if conn != nil {
				_ = conn.Close()
			}
		}
	}()

	select {
	case result := <-resultCh:
		if err := ctx.Err(); err != nil {
			if result.conn != nil {
				_ = result.conn.Close()
			}
			return nil, err
		}
		return result.conn, result.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
