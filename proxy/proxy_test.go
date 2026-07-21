package proxy

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestCloseStopsListenAndServe(t *testing.T) {
	p := New(logrus.New(), nil)
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- p.ListenAndServe(0)
	}()

	select {
	case <-p.tcpListened:
	case <-time.After(time.Second):
		t.Fatal("proxy did not start listening")
	}
	if err := p.Close(); err != nil {
		t.Fatalf("close proxy: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("close proxy again: %v", err)
	}
	select {
	case err := <-serveErr:
		if err != nil {
			t.Fatalf("ListenAndServe returned an error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ListenAndServe did not stop")
	}
}
