package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"charm.land/huh/v2"
	"github.com/mzz2017/gg/dialer"
)

func TestSelectNodeSortsByLatencyAndReturnsChoice(t *testing.T) {
	nodes := []*DialerWithLatency{
		newDialerWithLatency("offline", -1),
		newDialerWithLatency("slow", 80),
		newDialerWithLatency("fast", 20),
	}

	selected, err := selectNode(nodes, func(prompt *huh.Select[*DialerWithLatency]) error {
		var output bytes.Buffer
		return prompt.RunAccessible(&output, strings.NewReader("2\n"))
	})
	if err != nil {
		t.Fatal(err)
	}
	if selected.Dialer.Name() != "slow" {
		t.Fatalf("selected node = %q, want slow", selected.Dialer.Name())
	}
	wantOrder := []string{"fast", "slow", "offline"}
	for i, want := range wantOrder {
		if got := nodes[i].Dialer.Name(); got != want {
			t.Fatalf("nodes[%d] = %q, want %q", i, got, want)
		}
	}
}

func TestSelectNodeReturnsRunnerError(t *testing.T) {
	wantErr := errors.New("selection failed")
	_, err := selectNode(
		[]*DialerWithLatency{newDialerWithLatency("node", 20)},
		func(*huh.Select[*DialerWithLatency]) error { return wantErr },
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestPromptForNodeLink(t *testing.T) {
	link, err := promptForNodeLink(func(input *huh.Input) error {
		return input.RunAccessible(
			&bytes.Buffer{},
			strings.NewReader("   \n  socks5://127.0.0.1:1080  \n"),
		)
	})
	if err != nil {
		t.Fatal(err)
	}
	if link != "socks5://127.0.0.1:1080" {
		t.Fatalf("link = %q, want trimmed SOCKS link", link)
	}
}

func TestPromptForNodeLinkReturnsRunnerError(t *testing.T) {
	wantErr := errors.New("input failed")
	_, err := promptForNodeLink(func(*huh.Input) error { return wantErr })
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestValidateShareLink(t *testing.T) {
	if err := validateShareLink(" \t "); !errors.Is(err, errShareLinkRequired) {
		t.Fatalf("whitespace error = %v, want %v", err, errShareLinkRequired)
	}
	if err := validateShareLink("socks5://127.0.0.1:1080"); err != nil {
		t.Fatalf("valid link error = %v", err)
	}
}

func TestNodeDetails(t *testing.T) {
	node := &DialerWithLatency{
		Dialer:  dialer.NewDialer(nil, true, "test-node", "socks5", ""),
		Latency: 42,
	}
	got := nodeDetails(node)
	for _, want := range []string{"test-node", "socks5", "true", "42 ms"} {
		if !strings.Contains(got, want) {
			t.Errorf("details %q do not contain %q", got, want)
		}
	}
	if got := nodeDetails(nil); got != "" {
		t.Fatalf("nil node details = %q, want empty", got)
	}
}

func newDialerWithLatency(name string, latency int) *DialerWithLatency {
	return &DialerWithLatency{
		Dialer:  dialer.NewDialer(nil, false, name, "test", ""),
		Latency: latency,
	}
}
