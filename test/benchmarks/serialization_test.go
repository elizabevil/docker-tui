package benchmarks

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/pelletier/go-toml/v2"
)

// benchData holds the full i18n-like dataset with nesting.
type benchData struct {
	Meta struct {
		Name        string `json:"name" toml:"name"`
		Version     int    `json:"version" toml:"version"`
		Description string `json:"description" toml:"description"`
	} `json:"meta" toml:"meta"`

	Panel map[string]string `json:"panel" toml:"panel"`
	Key   map[string]string `json:"key" toml:"key"`
	Toast map[string]string `json:"toast" toml:"toast"`

	Dialog map[string]string `json:"dialog" toml:"dialog"`
	Status map[string]string `json:"status" toml:"status"`
	Table  map[string]string `json:"table" toml:"table"`
	Help   map[string]string `json:"help" toml:"help"`
	Msg    map[string]string `json:"msg" toml:"msg"`
	Error  map[string]string `json:"error" toml:"error"`
	Format map[string]string `json:"format" toml:"format"`
}

var (
	jsonBytes []byte
	tomlBytes []byte
)

func init() {
	var err error
	jsonBytes, err = os.ReadFile("benchdata.json")
	if err != nil {
		panic("benchdata.json: " + err.Error())
	}
	tomlBytes, err = os.ReadFile("benchdata.toml")
	if err != nil {
		panic("benchdata.toml: " + err.Error())
	}
}

// ── Unmarshal benchmarks ──────────────────────────────────

func BenchmarkUnmarshalJSONStd(b *testing.B) {
	var d benchData
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(jsonBytes, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalJSONSonic(b *testing.B) {
	var d benchData
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := sonic.Unmarshal(jsonBytes, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalTOML(b *testing.B) {
	var d benchData
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := toml.Unmarshal(tomlBytes, &d); err != nil {
			b.Fatal(err)
		}
	}
}

// ── Marshal benchmarks ────────────────────────────────────

func BenchmarkMarshalJSONStd(b *testing.B) {
	var d benchData
	if err := json.Unmarshal(jsonBytes, &d); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalJSONSonic(b *testing.B) {
	var d benchData
	if err := sonic.Unmarshal(jsonBytes, &d); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := sonic.Marshal(d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalTOML(b *testing.B) {
	var d benchData
	if err := toml.Unmarshal(tomlBytes, &d); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := toml.Marshal(d); err != nil {
			b.Fatal(err)
		}
	}
}

// ── Round-trip benchmarks ─────────────────────────────────

func BenchmarkRoundtripJSONStd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var d benchData
		if err := json.Unmarshal(jsonBytes, &d); err != nil {
			b.Fatal(err)
		}
		if _, err := json.Marshal(d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundtripJSONSonic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var d benchData
		if err := sonic.Unmarshal(jsonBytes, &d); err != nil {
			b.Fatal(err)
		}
		if _, err := sonic.Marshal(d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundtripTOML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var d benchData
		if err := toml.Unmarshal(tomlBytes, &d); err != nil {
			b.Fatal(err)
		}
		if _, err := toml.Marshal(d); err != nil {
			b.Fatal(err)
		}
	}
}
