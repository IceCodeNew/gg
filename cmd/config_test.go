package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestWriteConfigRoundTrip(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nested", "config.toml")
	settings := map[string]interface{}{
		"no_udp":       true,
		"dial_timeout": 12,
		"subscription": map[string]interface{}{
			"link": "https://example.com/subscription?token=value#fragment",
		},
	}

	if err := WriteConfig(settings, configPath); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0600); got != want {
		t.Fatalf("config permissions = %o, want %o", got, want)
	}

	decoded := viper.New()
	decoded.SetConfigFile(configPath)
	if err := decoded.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if !decoded.GetBool("no_udp") {
		t.Fatal("no_udp = false, want true")
	}
	if got := decoded.GetInt("dial_timeout"); got != 12 {
		t.Fatalf("dial_timeout = %d, want 12", got)
	}
	if got, want := decoded.GetString("subscription.link"), settings["subscription"].(map[string]interface{})["link"]; got != want {
		t.Fatalf("subscription.link = %q, want %q", got, want)
	}
}
