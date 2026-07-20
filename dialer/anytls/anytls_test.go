package anytls

import (
	"net/url"
	"testing"

	"github.com/mzz2017/gg/dialer"
)

func TestNew(t *testing.T) {
	got, err := New(
		"anytls://test-auth@192.0.2.1:443?sni=proxy.example&insecure=1#anytls-node",
		&dialer.GlobalOption{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name() != "anytls-node" {
		t.Fatalf("name = %q, want anytls-node", got.Name())
	}
	if got.Protocol() != "anytls" {
		t.Fatalf("protocol = %q, want anytls", got.Protocol())
	}
	if !got.SupportUDP() {
		t.Fatal("expected UDP support")
	}
	link, err := url.Parse(got.Link())
	if err != nil {
		t.Fatal(err)
	}
	if link.Scheme != "anytls" {
		t.Fatalf("link scheme = %q, want anytls", link.Scheme)
	}
}
