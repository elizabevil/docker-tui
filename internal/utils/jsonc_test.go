package utils

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
)

// ── StripJSONCComments ──────────────────────────────────────

func TestStripJSONCComments(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"no comments", `{"a":1,"b":2}`, `{"a":1,"b":2}`},
		{"empty", "", ""},
		{"leading line comment", "// hi\n{}", "\n{}"},
		{"trailing line comment", "{} // bye", "{} "},
		{"block comment", "/* x */{}", "{}"},
		{"nested-looking block comment in string", `{"a":"/* not a comment */"}`, `{"a":"/* not a comment */"}`},
		{"line comment marker in string", `{"a":"// not a comment"}`, `{"a":"// not a comment"}`},
		{"url inside string", `{"url":"http://example.com/foo"}`, `{"url":"http://example.com/foo"}`},
		{"escaped quote inside string", `{"msg":"he said \"hi\" // not a comment"}`, `{"msg":"he said \"hi\" // not a comment"}`},
		{"escaped backslash", `{"a":"back\\slash // not a comment"}`, `{"a":"back\\slash // not a comment"}`},
		{"block comment spans newline", "/* line1\nline2 */\n{}", "\n{}"},
		{"multiple comments", "// a\n/* b */{\"k\":\"v\"} // c\n", "\n{\"k\":\"v\"} \n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(StripJSONCComments([]byte(tc.in)))
			if got != tc.want {
				t.Fatalf("got %q\nwant %q", got, tc.want)
			}
		})
	}
}

func TestStripJSONCComments_NoChangeWhenNoComments(t *testing.T) {
	input := []byte(`{"a":1,"b":"x","c":[1,2,3]}`)
	out := StripJSONCComments(input)
	if string(out) != string(input) {
		t.Fatalf("expected unchanged, got %s", out)
	}
}

// ── IsJSONC ────────────────────────────────────────────────

func TestIsJSONC(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{`{"a":1}`, false},
		{`{"url":"http://x"}`, false},
		{`{"a":1} // tail`, true},
		{`/* head */{}`, true},
		{`{"a":"// not"}`, false},
		{`{"a":"/* not */"}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := IsJSONC([]byte(tc.in)); got != tc.want {
				t.Fatalf("IsJSONC(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// ── Unmarshalers ────────────────────────────────────────────

func TestUnmarshalJSONCStd(t *testing.T) {
	in := []byte(`{
		// a friendly payload
		"meta": { "name": "hi" /* version */ },
		"list": [1, 2 /* three */, 4]
	}`)
	var v map[string]any
	if err := UnmarshalJSONCStd(in, &v); err != nil {
		t.Fatalf("UnmarshalJSONCStd: %v", err)
	}
	meta, ok := v["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta missing or wrong type: %T", v["meta"])
	}
	if meta["name"] != "hi" {
		t.Errorf("meta.name = %v", meta["name"])
	}
}

func TestUnmarshalJSONCSonic(t *testing.T) {
	in := []byte(`{
		// comment
		"name": "test",
		"value": 42
	}`)
	var v struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	if err := UnmarshalJSONCSonic(in, &v); err != nil {
		t.Fatalf("UnmarshalJSONCSonic: %v", err)
	}
	if v.Name != "test" || v.Value != 42 {
		t.Errorf("got %+v", v)
	}
}

func TestLoadJSONC(t *testing.T) {
	in := []byte(`// hi
{"k":"v"}`)
	var v map[string]string
	if err := LoadJSONC(in, &v); err != nil {
		t.Fatalf("LoadJSONC: %v", err)
	}
	if v["k"] != "v" {
		t.Errorf("got %v", v)
	}
}

func TestUnmarshalJSONCPureJSONStillWorks(t *testing.T) {
	in := []byte(`{"a":1,"b":"two"}`)
	var v map[string]any
	if err := UnmarshalJSONCStd(in, &v); err != nil {
		t.Fatalf("unmarshal pure json: %v", err)
	}
	if v["a"].(float64) != 1 {
		t.Errorf("got %v", v)
	}
}

func TestUnmarshalJSONCInvalidPayload(t *testing.T) {
	in := []byte(`{"a": }`)
	var v map[string]any
	if err := UnmarshalJSONCStd(in, &v); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// ── Marshaller ──────────────────────────────────────────────

func TestMarshalJSONCSonicWithHeader(t *testing.T) {
	v := map[string]int{"x": 1, "y": 2}
	out, err := MarshalJSONCSonic(v, JSONCHeader{Lines: []string{"generated", "do not edit"}})
	if err != nil {
		t.Fatalf("MarshalJSONCSonic: %v", err)
	}
	s := string(out)
	if !strings.HasPrefix(s, "// generated\n// do not edit\n") {
		t.Errorf("missing header, got: %s", s)
	}
	// Verify the body roundtrips.
	var got map[string]int
	if err := json.Unmarshal(out[bytes.IndexByte(out, '{'):], &got); err != nil {
		t.Fatalf("body invalid: %v", err)
	}
	if got["x"] != 1 || got["y"] != 2 {
		t.Errorf("roundtrip mismatch: %v", got)
	}
}

func TestMarshalJSONCStdWithHeader(t *testing.T) {
	v := struct {
		Name string `json:"name"`
	}{Name: "abc"}
	out, err := MarshalJSONCStd(v, JSONCHeader{Lines: []string{"a", "b", "c"}})
	if err != nil {
		t.Fatalf("MarshalJSONCStd: %v", err)
	}
	if !strings.HasPrefix(string(out), "// a\n// b\n// c\n") {
		t.Errorf("missing header, got: %s", out)
	}
	// Body should be indented JSON (contains newlines and 2-space indent).
	if !strings.Contains(string(out), "  \"name\": \"abc\"") {
		t.Errorf("expected indented body, got: %s", out)
	}
}

func TestMarshalJSONCNoHeader(t *testing.T) {
	v := map[string]string{"k": "v"}
	out, err := MarshalJSONCSonic(v, JSONCHeader{})
	if err != nil {
		t.Fatalf("MarshalJSONCSonic: %v", err)
	}
	// Must start with { (no leading header).
	if out[0] != '{' {
		t.Errorf("expected body to start with '{', got: %s", out)
	}
}

func TestSaveJSONC(t *testing.T) {
	v := map[string]string{"k": "v"}
	out, err := SaveJSONC(v, JSONCHeader{Lines: []string{"saved"}})
	if err != nil {
		t.Fatalf("SaveJSONC: %v", err)
	}
	if !strings.HasPrefix(string(out), "// saved\n") {
		t.Errorf("missing header, got: %s", out)
	}
}

// ── Round-trip ──────────────────────────────────────────────

func TestJSONCRoundTripSonic(t *testing.T) {
	type payload struct {
		Name    string         `json:"name"`
		Tags    []string       `json:"tags"`
		Counts  map[string]int `json:"counts"`
		Enabled bool           `json:"enabled"`
	}
	src := payload{
		Name:    "demo",
		Tags:    []string{"a", "b"},
		Counts:  map[string]int{"x": 1, "y": 2},
		Enabled: true,
	}
	body, err := MarshalJSONCSonic(src, JSONCHeader{Lines: []string{"auto"}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got payload
	if err := UnmarshalJSONCSonic(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Name != src.Name || len(got.Tags) != 2 || got.Counts["x"] != 1 || got.Enabled != src.Enabled {
		t.Errorf("roundtrip mismatch: %+v vs %+v", got, src)
	}
}

func TestJSONCRoundTripStd(t *testing.T) {
	src := map[string]any{"k": "v", "n": 7, "b": true}
	body, err := MarshalJSONCStd(src, JSONCHeader{Lines: []string{"std"}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := UnmarshalJSONCStd(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["k"] != "v" || got["n"].(float64) != 7 || got["b"] != true {
		t.Errorf("roundtrip mismatch: %v", got)
	}
}

// ── Sonic-specific helpers ──────────────────────────────────

func TestUnmarshalJSONCSonicEquivalentToStd(t *testing.T) {
	src := map[string]any{"k": "v", "n": 42, "list": []any{1, 2, 3}}
	body, err := sonic.Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	jsonc := []byte("// header\n" + string(body) + "\n// footer\n")

	var a, b map[string]any
	if err := UnmarshalJSONCStd(jsonc, &a); err != nil {
		t.Fatalf("std: %v", err)
	}
	if err := UnmarshalJSONCSonic(jsonc, &b); err != nil {
		t.Fatalf("sonic: %v", err)
	}
	if a["k"] != b["k"] || a["n"] != b["n"] {
		t.Errorf("std/sonic differ: %v vs %v", a, b)
	}
}