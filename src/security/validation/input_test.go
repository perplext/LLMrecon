package validation

import (
	"strings"
	"testing"
)

func TestValidatePrompt(t *testing.T) {
	v := NewInputValidator()
	cases := []struct {
		name    string
		prompt  string
		wantErr string // substring; "" = no error
	}{
		{"plain text", "Summarize this document.", ""},
		{"allowed whitespace controls", "line1\nline2\tcol\rret", ""},
		{"empty", "", ""},
		{"too long", strings.Repeat("a", 10001), "maximum length"},
		{"at max length", strings.Repeat("a", 10000), ""},
		{"null byte", "before\x00after", "null bytes"},
		{"control character", "bad\x01char", "control characters"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertErr(t, v.ValidatePrompt(c.prompt), c.wantErr)
		})
	}
}

func TestValidateURL(t *testing.T) {
	v := NewInputValidator()
	cases := []struct {
		name, url, wantErr string
	}{
		{"valid https", "https://example.com/path", ""},
		{"valid http", "http://example.com", ""},
		{"ftp scheme", "ftp://example.com", "invalid URL scheme"},
		{"file scheme", "file:///etc/passwd", "invalid URL scheme"},
		{"localhost SSRF", "http://localhost/admin", "private networks"},
		{"loopback SSRF", "http://127.0.0.1/", "private networks"},
		{"private 192.168", "http://192.168.1.1/", "private networks"},
		{"private 10.x", "http://10.0.0.5/", "private networks"},
		{"private 172.x", "http://172.16.0.1/", "private networks"},
		{"unparseable", "http://\x00bad", "invalid URL"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertErr(t, v.ValidateURL(c.url), c.wantErr)
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	v := NewInputValidator()
	cases := []struct {
		name, path, wantErr string
	}{
		{"normal relative path", "data/templates/foo.yaml", ""},
		{"path traversal", "../../etc/hosts", "path traversal"},
		{"null byte", "file\x00.txt", "null bytes"},
		{"etc passwd", "/etc/passwd", "sensitive file"},
		{"etc shadow", "/etc/shadow", "sensitive file"},
		{"ssh key", "/home/u/id_rsa", "sensitive file"},
		{"dotenv", "config/.env", "sensitive file"},
		{"git config", "repo/.git/config", "sensitive file"},
		{"case-insensitive sensitive", "/ETC/PASSWD", "sensitive file"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertErr(t, v.ValidateFilePath(c.path), c.wantErr)
		})
	}
}

func TestSanitizeString(t *testing.T) {
	v := NewInputValidator()

	if got := v.SanitizeString("clean text"); got != "clean text" {
		t.Errorf("clean input changed: %q", got)
	}
	if got := v.SanitizeString("a\x00b"); got != "ab" {
		t.Errorf("null byte not removed: %q", got)
	}
	if got := v.SanitizeString("a\x01\x02b"); got != "ab" {
		t.Errorf("control chars not removed: %q", got)
	}
	if got := v.SanitizeString("a\nb\tc\rd"); got != "a\nb\tc\rd" {
		t.Errorf("allowed whitespace controls were stripped: %q", got)
	}
	long := strings.Repeat("x", 10050)
	if got := v.SanitizeString(long); len(got) != 10000 {
		t.Errorf("over-length not truncated to 10000: got len %d", len(got))
	}
}

func TestValidateAPIKey(t *testing.T) {
	v := NewInputValidator()
	cases := []struct {
		name, key, wantErr string
	}{
		{"valid", strings.Repeat("k", 40), ""},
		{"min length boundary", strings.Repeat("k", 20), ""},
		{"max length boundary", strings.Repeat("k", 200), ""},
		{"empty", "", "cannot be empty"},
		{"too short", strings.Repeat("k", 19), "invalid length"},
		{"too long", strings.Repeat("k", 201), "invalid length"},
		{"contains space", "key with spaces and more padding here", "invalid characters"},
		{"contains control", strings.Repeat("k", 19) + "\x01", "invalid characters"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertErr(t, v.ValidateAPIKey(c.key), c.wantErr)
		})
	}
}

func TestValidateModelName(t *testing.T) {
	v := NewInputValidator()
	cases := []struct {
		name, model, wantErr string
	}{
		{"simple", "gpt-4o", ""},
		{"vendor path", "anthropic/claude-3.5-sonnet", ""},
		{"with colon", "ollama:llama3", ""},
		{"empty", "", "cannot be empty"},
		{"too long", strings.Repeat("m", 101), "too long"},
		{"space", "gpt 4", "invalid characters"},
		{"bang", "model!", "invalid characters"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertErr(t, v.ValidateModelName(c.model), c.wantErr)
		})
	}
}

// assertErr checks that err matches the wantErr substring expectation.
func assertErr(t *testing.T, err error, wantErr string) {
	t.Helper()
	if wantErr == "" {
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		return
	}
	if err == nil {
		t.Errorf("expected error containing %q, got nil", wantErr)
		return
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Errorf("error %q does not contain %q", err.Error(), wantErr)
	}
}
