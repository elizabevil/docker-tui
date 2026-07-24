package config

import (
	"testing"

	"github.com/bytedance/sonic"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	if theme == nil {
		t.Fatal("DefaultTheme returned nil")
	}
	if theme.Name == "" {
		t.Error("expected theme name to be set")
	}
	if theme.Colors.Green == "" {
		t.Error("expected green color to be set")
	}
}

func TestLoadEmbeddedTheme(t *testing.T) {
	themes := []string{"default", "dark", "light", "nord", "dracula", "solarized"}
	for _, name := range themes {
		theme, err := LoadTheme(name)
		if err != nil {
			t.Errorf("LoadTheme(%q) failed: %v", name, err)
			continue
		}
		if theme.Name == "" {
			t.Errorf("theme %q has no name", name)
		}
		if theme.Colors.Cyan == "" {
			t.Errorf("theme %q has no cyan color", name)
		}
	}
}

func TestListThemes(t *testing.T) {
	names := ListThemes()
	if len(names) == 0 {
		t.Fatal("ListThemes returned empty")
	}
	found := false
	for _, n := range names {
		if n == "default" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'default' in theme list")
	}
}

func TestJSONCComments(t *testing.T) {
	// JSON with both // and /* */ comments
	input := []byte(`{
		// line comment
		"name": "test",
		/* block comment */
		"colors": {
			"green": "#2ecc71" // inline
		}
	}`)

	cleaned, _ := sonic.Marshal(input)

	// sonic should be able to parse the cleaned JSON
	parsed := make(map[string]interface{})
	if err := sonic.Unmarshal(cleaned, &parsed); err != nil {
		t.Fatalf("failed to parse JSONC: %v\ncleaned:\n%s", err, string(cleaned))
	}

	if parsed["name"] != "test" {
		t.Errorf("expected name='test', got %v", parsed["name"])
	}

	colors := parsed["colors"].(map[string]interface{})
	if colors["green"] != "#2ecc71" {
		t.Errorf("expected green='#2ecc71', got %v", colors["green"])
	}
}

func TestJSONCBlockCommentMidLine(t *testing.T) {
	input := []byte(`{
		"name": /* comment */ "test",
		"colors": {/* nested */ "green": "#2ecc71"}
	}`)

	cleaned, _ := sonic.Marshal(input)
	parsed := make(map[string]interface{})
	if err := sonic.Unmarshal(cleaned, &parsed); err != nil {
		t.Fatalf("failed to parse: %v\ncleaned:\n%s", err, string(cleaned))
	}

	if parsed["name"] != "test" {
		t.Errorf("expected name='test', got %v", parsed["name"])
	}
}

func TestStripJSONCommentsNoComments(t *testing.T) {
	input := []byte(`{"a": 1, "b": [2, 3]}`)
	cleaned, _ := sonic.Marshal(input)
	if string(cleaned) != string(input) {
		t.Errorf("expected no change:\n  got:  %s\n  want: %s", string(cleaned), string(input))
	}
}

func TestStripJSONCommentsStringWithSlashes(t *testing.T) {
	// Slashes inside strings should not be treated as comments
	input := []byte(`{"url": "http://example.com/foo"}`)
	cleaned, _ := sonic.Marshal(input)
	parsed := make(map[string]interface{})
	if err := sonic.Unmarshal(cleaned, &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if parsed["url"] != "http://example.com/foo" {
		t.Errorf("expected url to be preserved, got %v", parsed["url"])
	}
}

func TestStripJSONCommentsEscapedQuotes(t *testing.T) {
	// Escaped quotes inside strings
	input := []byte(`{"msg": "he said \"hello\" // not a comment"}`)
	cleaned, _ := sonic.Marshal(input)
	parsed := make(map[string]interface{})
	if err := sonic.Unmarshal(cleaned, &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if parsed["msg"] != "he said \"hello\" // not a comment" {
		t.Errorf("msg not preserved correctly, got %v", parsed["msg"])
	}
}
