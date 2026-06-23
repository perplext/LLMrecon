package loader

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

func newLoader() *TemplateLoader {
	// repoManager is only needed for remote sources; nil is fine for local/parse.
	return NewTemplateLoader(time.Hour, nil)
}

// parseTemplateFromContent is the real parser (format.LoadFromFile only reads
// raw bytes). These cover the issue's parsing + error paths.
func TestParseTemplateFromContent(t *testing.T) {
	jsonTmpl, err := parseTemplateFromContent("t.json", []byte(`{"id":"json-id","name":"n"}`))
	if err != nil {
		t.Fatalf("parse JSON: %v", err)
	}
	if jsonTmpl.ID != "json-id" {
		t.Fatalf("JSON ID = %q, want json-id", jsonTmpl.ID)
	}

	yamlTmpl, err := parseTemplateFromContent("t.yaml", []byte("id: yaml-id\nname: n\n"))
	if err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	if yamlTmpl.ID != "yaml-id" {
		t.Fatalf("YAML ID = %q, want yaml-id", yamlTmpl.ID)
	}
}

func TestParseTemplateFromContent_Errors(t *testing.T) {
	// Malformed YAML.
	if _, err := parseTemplateFromContent("t.yaml", []byte("id: [unclosed\n  : :")); err == nil {
		t.Error("malformed YAML must error")
	}
	// Malformed JSON.
	if _, err := parseTemplateFromContent("t.json", []byte("{not json")); err == nil {
		t.Error("malformed JSON must error")
	}
	// Unsupported extension.
	if _, err := parseTemplateFromContent("t.txt", []byte("whatever")); err == nil {
		t.Error("unsupported extension must error")
	}
}

func TestIsTemplateFile(t *testing.T) {
	cases := map[string]bool{
		"a.yaml": true, "a.yml": true, "a.json": true,
		"a.txt": false, "a.go": false, "a": false,
	}
	for name, want := range cases {
		if got := isTemplateFile(name); got != want {
			t.Errorf("isTemplateFile(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestLoad_MissingFile(t *testing.T) {
	l := newLoader()
	if _, err := l.Load(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("Load on missing file must error")
	}
}

func TestLoad_ReadsExistingFile(t *testing.T) {
	l := newLoader()
	p := filepath.Join(t.TempDir(), "t.yaml")
	if err := os.WriteFile(p, []byte("id: x\nname: y\n"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	tmpl, err := l.Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// format.LoadFromFile reads raw bytes into Content (it does not deep-parse).
	if len(tmpl.Content) == 0 {
		t.Fatal("loaded template should carry file content")
	}
}

func TestLoadTemplates_SourceRouting(t *testing.T) {
	l := newLoader()
	ctx := context.Background()

	if _, err := l.LoadTemplates(ctx, filepath.Join(t.TempDir(), "missing"), string(interfaces.FileSource)); err == nil {
		t.Error("missing FileSource path must error")
	}
	if _, err := l.LoadTemplates(ctx, "x", string(interfaces.GitLabSource)); err == nil {
		t.Error("GitLab source should report not implemented")
	}
	if _, err := l.LoadTemplates(ctx, "x", string(interfaces.DatabaseSource)); err == nil {
		t.Error("Database source should report not implemented")
	}
	if _, err := l.LoadTemplates(ctx, "x", "bogus-source"); err == nil {
		t.Error("unsupported source type must error")
	}
}

func TestLoadFromBytes_StubTemplate(t *testing.T) {
	l := newLoader()
	got, err := l.LoadFromBytes([]byte("anything"))
	if err != nil {
		t.Fatalf("LoadFromBytes: %v", err)
	}
	if got == nil {
		t.Fatal("LoadFromBytes should return a template")
	}

	tmpl, err := l.LoadFromBytesWithFormat([]byte("data"), "yaml")
	if err != nil || tmpl == nil {
		t.Fatalf("LoadFromBytesWithFormat: tmpl=%v err=%v", tmpl, err)
	}
}

func TestLoadFromURL_NotImplemented(t *testing.T) {
	if _, err := newLoader().LoadFromURL("https://example.com/t.yaml"); err == nil {
		t.Fatal("LoadFromURL should report not implemented")
	}
}

func TestCacheMethods(t *testing.T) {
	l := newLoader()

	// Seed the internal cache directly (same package): one fresh, one expired.
	l.cache["fresh"] = cacheEntry{template: &format.Template{ID: "fresh"}, expiration: time.Now().Add(time.Hour)}
	l.cache["stale"] = cacheEntry{template: &format.Template{ID: "stale"}, expiration: time.Now().Add(-time.Hour)}

	if l.GetCacheSize() != 2 {
		t.Fatalf("GetCacheSize = %d, want 2", l.GetCacheSize())
	}
	if pruned := l.PruneCache(); pruned != 1 {
		t.Fatalf("PruneCache removed %d, want 1 (the expired entry)", pruned)
	}
	if l.GetCacheSize() != 1 {
		t.Fatalf("after prune size = %d, want 1", l.GetCacheSize())
	}
	l.ClearCache()
	if l.GetCacheSize() != 0 {
		t.Fatalf("after clear size = %d, want 0", l.GetCacheSize())
	}
}
