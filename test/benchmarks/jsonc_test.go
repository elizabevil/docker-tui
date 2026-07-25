package benchmarks

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// benchDataJSONC holds a JSONC dataset with comments to verify the parser
// tolerates // and /* */ comments before the actual JSON payload.
type benchDataJSONC struct {
	Meta struct {
		Name        string `json:"name"`
		Version     int    `json:"version"`
		Description string `json:"description"`
	} `json:"meta"`

	Panel map[string]string `json:"panel"`
	Key   map[string]string `json:"key"`
	Toast map[string]string `json:"toast"`

	Dialog map[string]string `json:"dialog"`
	Status map[string]string `json:"status"`
	Table  map[string]string `json:"table"`
	Help   map[string]string `json:"help"`
	Msg    map[string]string `json:"msg"`
}

var jsoncBytes []byte

func init() {
	var err error
	jsoncBytes, err = os.ReadFile("benchdata.jsonc")
	if err != nil {
		panic("benchdata.jsonc: " + err.Error())
	}
}

// ── JSONC parser benchmarks ────────────────────────────────

func BenchmarkStripJSONCComments(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = utils.StripJSONCComments(jsoncBytes)
	}
}

func BenchmarkUnmarshalJSONCStd(b *testing.B) {
	clean := utils.StripJSONCComments(jsoncBytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var d benchDataJSONC
		if err := json.Unmarshal(clean, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalJSONCSonic(b *testing.B) {
	clean := utils.StripJSONCComments(jsoncBytes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var d benchDataJSONC
		if err := sonic.Unmarshal(clean, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalJSONCRawStd(b *testing.B) {
	// Without stripping comments — expected to fail.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var d benchDataJSONC
		_ = json.Unmarshal(jsoncBytes, &d)
	}
}

// ── JSONC correctness tests ────────────────────────────────

func TestStripJSONCComments(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no comments",
			in:   `{"a":1,"b":2}`,
			want: `{"a":1,"b":2}`,
		},
		{
			name: "line comments",
			in:   `// leading` + "\n" + `{"a":1} // trailing` + "\n",
			want: "\n" + `{"a":1} ` + "\n",
		},
		{
			name: "block comments",
			in:   `/* head */{"a":1}/* tail */`,
			want: `{"a":1}`,
		},
		{
			name: "url inside string",
			in:   `{"url":"http://example.com/foo"}`,
			want: `{"url":"http://example.com/foo"}`,
		},
		{
			name: "comment-like content inside string",
			in:   `{"msg":"he said \"hi\" // not a comment"}`,
			want: `{"msg":"he said \"hi\" // not a comment"}`,
		},
		{
			name: "block-comment-like content inside string",
			in:   `{"msg":"a /* not */ comment"}`,
			want: `{"msg":"a /* not */ comment"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(utils.StripJSONCComments([]byte(tc.in)))
			if got != tc.want {
				t.Fatalf("got %q\nwant %q", got, tc.want)
			}
		})
	}
}

func TestUnmarshalJSONCStd(t *testing.T) {
	clean := utils.StripJSONCComments(jsoncBytes)
	var d benchDataJSONC
	if err := json.Unmarshal(clean, &d); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if d.Meta.Name == "" {
		t.Error("expected meta.name to be set")
	}
	if d.Panel["containers"] != "Containers" {
		t.Errorf("expected panel.containers=Containers, got %q", d.Panel["containers"])
	}
}

func TestUnmarshalJSONCSonic(t *testing.T) {
	clean := utils.StripJSONCComments(jsoncBytes)
	var d benchDataJSONC
	if err := sonic.Unmarshal(clean, &d); err != nil {
		t.Fatalf("sonic.Unmarshal: %v", err)
	}
	if d.Meta.Name == "" {
		t.Error("expected meta.name to be set")
	}
	if d.Panel["containers"] != "Containers" {
		t.Errorf("expected panel.containers=Containers, got %q", d.Panel["containers"])
	}
}

func TestRawJSONCUnmarshalFails(t *testing.T) {
	// Sanity check: without comment stripping the parser must reject the input.
	var d benchDataJSONC
	if err := json.Unmarshal(jsoncBytes, &d); err == nil {
		t.Fatal("expected raw JSONC to fail, got nil")
	} else {
		t.Log(err)
	}
	if err := sonic.Unmarshal(jsoncBytes, &d); err == nil {
		t.Fatal("expected raw JSONC to fail in sonic, got nil")
	} else {
		t.Log(err)
	}
}

func TestJSONCRoundtripStd(t *testing.T) {
	clean := utils.StripJSONCComments(jsoncBytes)
	var d benchDataJSONC
	if err := json.Unmarshal(clean, &d); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var again benchDataJSONC
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatalf("unmarshal again: %v", err)
	}
	if again.Meta.Name != d.Meta.Name {
		t.Errorf("meta.name roundtrip: got %q want %q", again.Meta.Name, d.Meta.Name)
	}
}

func TestJSONCRoundtripSonic(t *testing.T) {
	clean := utils.StripJSONCComments(jsoncBytes)
	var d benchDataJSONC
	if err := sonic.Unmarshal(clean, &d); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := sonic.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var again benchDataJSONC
	if err := sonic.Unmarshal(out, &again); err != nil {
		t.Fatalf("unmarshal again: %v", err)
	}
	if again.Meta.Name != d.Meta.Name {
		t.Errorf("meta.name roundtrip: got %q want %q", again.Meta.Name, d.Meta.Name)
	}
}
