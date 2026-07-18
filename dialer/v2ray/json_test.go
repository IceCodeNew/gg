package v2ray

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseVmessURLFuzzyStrings(t *testing.T) {
	raw := `{"v":2,"ps":null,"add":"proxy.example","port":443,"id":"00000000-0000-0000-0000-000000000001","aid":0,"net":"ws","type":"none","host":"front.example","path":"/tunnel","tls":"tls","allowInsecure":1}`
	link := "vmess://" + strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(raw)), "=")

	got, err := ParseVmessURL(link)
	if err != nil {
		t.Fatal(err)
	}
	if got.V != "2" || got.Ps != "" || got.Port != "443" || got.Aid != "0" {
		t.Fatalf("fuzzy fields = V:%q Ps:%q Port:%q Aid:%q", got.V, got.Ps, got.Port, got.Aid)
	}
	if !got.AllowInsecure || got.Protocol != "vmess" {
		t.Fatalf("parsed VMess = %#v", got)
	}
}

func TestFuzzyStringRejectsOtherJSONTypes(t *testing.T) {
	for _, value := range []string{"true", `{}`, `[]`} {
		var got fuzzyString
		if err := json.Unmarshal([]byte(value), &got); err == nil {
			t.Errorf("json.Unmarshal(%s) unexpectedly succeeded", value)
		}
	}
}

func TestFuzzyBool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "true", want: true},
		{input: "false"},
		{input: "1", want: true},
		{input: "0"},
		{input: `"1"`, want: true},
		{input: `"0"`},
		{input: `"true"`, want: true},
		{input: `"false"`},
		{input: "null"},
	}
	for _, test := range tests {
		var got fuzzyBool
		if err := json.Unmarshal([]byte(test.input), &got); err != nil {
			t.Errorf("json.Unmarshal(%s): %v", test.input, err)
		} else if bool(got) != test.want {
			t.Errorf("json.Unmarshal(%s) = %t, want %t", test.input, got, test.want)
		}
	}

	for _, input := range []string{"2", `{}`, `[]`} {
		var got fuzzyBool
		if err := json.Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("json.Unmarshal(%s) unexpectedly succeeded", input)
		}
	}
}

func TestExportVmessURLRoundTrip(t *testing.T) {
	want := &V2Ray{
		Ps:       "node",
		Add:      "proxy.example",
		Port:     "443",
		ID:       "00000000-0000-0000-0000-000000000001",
		Aid:      "0",
		Net:      "ws",
		Type:     "none",
		Host:     "front.example",
		Path:     "/tunnel",
		TLS:      "tls",
		Protocol: "vmess",
	}

	got, err := ParseVmessURL(want.ExportToURL())
	if err != nil {
		t.Fatal(err)
	}
	if *got != *want {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}
}
