package utils

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
)

// ── Comment stripping ────────────────────────────────────────

// StripJSONCComments removes // line comments and /* block */ comments from
// a JSONC document. Content inside string literals is preserved verbatim,
// including comment-like sequences (e.g. "http://example.com" or "/* not */").
//
// The function is allocation-conscious: it walks the input byte-by-byte
// without intermediate string conversions when possible.
func StripJSONCComments(data []byte) []byte {
	var out bytes.Buffer
	out.Grow(len(data))

	inString := false
	escape := false
	i := 0
	n := len(data)
	for i < n {
		c := data[i]

		if inString {
			out.WriteByte(c)
			switch {
			case escape:
				escape = false
			case c == '\\':
				escape = true
			case c == '"':
				inString = false
			}
			i++
			continue
		}

		// Not in a string — check for comment markers.
		if c == '"' {
			inString = true
			out.WriteByte(c)
			i++
			continue
		}

		if c == '/' && i+1 < n {
			next := data[i+1]
			if next == '/' {
				// Line comment: drop until (but keep) the newline.
				i += 2
				for i < n && data[i] != '\n' {
					i++
				}
				continue
			}
			if next == '*' {
				// Block comment: drop until closing */.
				i += 2
				for i+1 < n && (data[i] != '*' || data[i+1] != '/') {
					i++
				}
				i += 2
				continue
			}
		}

		out.WriteByte(c)
		i++
	}
	return out.Bytes()
}

// IsJSONC returns true if the data contains JSONC comment markers outside
// of string literals. Cheap pre-check that avoids the full strip pass when
// the input is already pure JSON.
func IsJSONC(data []byte) bool {
	inString := false
	escape := false
	for i := range data {
		c := data[i]
		if inString {
			switch {
			case escape:
				escape = false
			case c == '\\':
				escape = true
			case c == '"':
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == '/' && i+1 < len(data) {
			next := data[i+1]
			if next == '/' || next == '*' {
				return true
			}
		}
	}
	return false
}

// ── JSONC-aware unmarshalers ────────────────────────────────

// UnmarshalJSONCStd decodes JSONC data into v using encoding/json. Comments
// are stripped before unmarshalling.
func UnmarshalJSONCStd(data []byte, v any) error {
	clean := StripJSONCComments(data)
	return json.Unmarshal(clean, v)
}

// UnmarshalJSONCSonic decodes JSONC data into v using bytedance/sonic.
// Comments are stripped before unmarshalling.
func UnmarshalJSONCSonic(data []byte, v any) error {
	clean := StripJSONCComments(data)
	return sonic.Unmarshal(clean, v)
}

// ── JSONC-aware marshaller ──────────────────────────────────

// JSONCHeader describes an optional file header written before the JSON
// payload. Lines are joined with newlines and emitted as // comments.
type JSONCHeader struct {
	Lines []string
}

// MarshalJSONCSonic encodes v to JSON and prepends a // comment header if
// one is supplied. The body itself is valid JSON; the header turns it into
// a JSONC document.
func MarshalJSONCSonic(v any, header JSONCHeader) ([]byte, error) {
	body, err := sonic.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal jsonc: %w", err)
	}
	return withHeader(body, header), nil
}

// MarshalJSONCStd is the encoding/json variant of MarshalJSONCSonic.
func MarshalJSONCStd(v any, header JSONCHeader) ([]byte, error) {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal jsonc: %w", err)
	}
	return withHeader(body, header), nil
}

func withHeader(body []byte, header JSONCHeader) []byte {
	if len(header.Lines) == 0 {
		return body
	}
	var out bytes.Buffer
	for i, line := range header.Lines {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString("// ")
		out.WriteString(line)
	}
	if len(body) > 0 {
		out.WriteByte('\n')
		out.Write(body)
	}
	return out.Bytes()
}

// ── Convenience helpers ─────────────────────────────────────

// LoadJSONC reads a JSONC byte slice (already loaded from disk or an embed
// FS) and decodes it into v with the standard library.
func LoadJSONC(data []byte, v any) error { return UnmarshalJSONCStd(data, v) }

// SaveJSONC encodes v as indented JSON with a comment header.
func SaveJSONC(v any, header JSONCHeader) ([]byte, error) {
	return MarshalJSONCStd(v, header)
}
