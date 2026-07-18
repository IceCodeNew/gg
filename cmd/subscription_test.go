package cmd

import (
	"encoding/base64"
	"testing"

	"github.com/mzz2017/gg/dialer"
	_ "github.com/mzz2017/gg/dialer/socks"
)

func TestResolveSubscriptionAsClash(t *testing.T) {
	config := []byte(`proxies:
  - name: local-socks
    type: socks5
    server: 192.0.2.1
    port: 1080
    username: alice
    password: secret
    udp: true
`)
	tests := map[string][]byte{
		"plain":  config,
		"base64": []byte(base64.StdEncoding.EncodeToString(config)),
	}

	for name, subscription := range tests {
		t.Run(name, func(t *testing.T) {
			dialers, err := resolveSubscriptionAsClash(NewLogger(0), &dialer.GlobalOption{}, subscription)
			if err != nil {
				t.Fatal(err)
			}
			if len(dialers) != 1 {
				t.Fatalf("dialers = %d, want 1", len(dialers))
			}
			if got := dialers[0].Name(); got != "local-socks" {
				t.Fatalf("name = %q, want local-socks", got)
			}
			if got := dialers[0].Protocol(); got != "socks5" {
				t.Fatalf("protocol = %q, want socks5", got)
			}
			if !dialers[0].SupportUDP() {
				t.Fatal("expected UDP support")
			}
		})
	}
}
