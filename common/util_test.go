package common_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/mzz2017/gg/common"
	"github.com/mzz2017/gg/config"
)

func TestObjectToKV(t *testing.T) {
	type nested struct {
		Enabled bool `mapstructure:"enabled"`
	}
	type object struct {
		Name       string `mapstructure:"name"`
		Nested     nested `mapstructure:"nested"`
		Count      int
		Empty      string `mapstructure:"empty,unused,omitempty"`
		Ignored    string `mapstructure:"-"`
		unexported string
	}

	got := common.ObjectToKV(&object{
		Name:    "proxy",
		Nested:  nested{Enabled: true},
		Count:   2,
		Ignored: "ignored",
	}, "mapstructure")
	sort.Strings(got)
	want := []string{"Count=2", "name=proxy", "nested.enabled=true"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ObjectToKV() = %#v, want %#v", got, want)
	}
}

func TestObjectToKVConfigParams(t *testing.T) {
	got := common.ObjectToKV(config.Params{}, "mapstructure")
	sort.Strings(got)
	want := []string{
		"allow_insecure=false",
		"cache.subscription.last_node=",
		"no_udp=false",
		"node=",
		"proxy_private=false",
		"subscription.cache_last_node=false",
		"subscription.link=",
		"subscription.select=",
		"test_node_before_use=false",
		"test_url=",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ObjectToKV() = %#v, want %#v", got, want)
	}
}

func TestObjectToKVUnsupportedValues(t *testing.T) {
	var typedNil *struct{}
	for _, value := range []any{nil, typedNil, 1, map[string]any{}} {
		if got := common.ObjectToKV(value, "mapstructure"); got != nil {
			t.Errorf("ObjectToKV(%#v) = %#v, want nil", value, got)
		}
	}
}
