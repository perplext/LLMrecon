package format

import (
	"os"
	"path/filepath"
	"testing"
)

func validTemplate() *Template {
	return &Template{
		ID:   "tmpl-1",
		Name: "Test Template",
		Info: &TemplateInfo{Name: "Test Template", Version: "1.0.0", Severity: "high"},
		Test: &TemplateTest{},
	}
}

func TestValidateStructure(t *testing.T) {
	if err := validTemplate().Validate(); err != nil {
		t.Fatalf("valid template should pass: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*Template)
	}{
		{"missing ID", func(t *Template) { t.ID = "" }},
		{"missing Name", func(t *Template) { t.Name = "" }},
		{"missing Info", func(t *Template) { t.Info = nil }},
		{"missing Test", func(t *Template) { t.Test = nil }},
	}
	for _, c := range cases {
		tmpl := validTemplate()
		c.mutate(tmpl)
		if err := tmpl.ValidateStructure(); err == nil {
			t.Errorf("%s: expected validation error", c.name)
		}
	}
}

func TestGetters(t *testing.T) {
	tmpl := validTemplate()
	tmpl.Version = "2.1.0"
	tmpl.Content = []byte("body")
	tmpl.Info.Tags = []string{"injection", "owasp"}
	tmpl.Info.Description = "desc"
	tmpl.Info.Author = "alice"

	if tmpl.GetID() != "tmpl-1" || tmpl.GetName() != "Test Template" {
		t.Fatal("GetID/GetName mismatch")
	}
	if tmpl.GetVersion() != "2.1.0" {
		t.Fatalf("GetVersion = %q", tmpl.GetVersion())
	}
	if tmpl.GetSeverity() != "high" {
		t.Fatalf("GetSeverity = %q", tmpl.GetSeverity())
	}
	if tmpl.GetCategory() != "injection" { // first tag
		t.Fatalf("GetCategory = %q, want injection", tmpl.GetCategory())
	}
	if tmpl.GetDescription() != "desc" || tmpl.GetAuthor() != "alice" {
		t.Fatal("GetDescription/GetAuthor mismatch")
	}

	body, err := tmpl.GetContent()
	if err != nil || string(body) != "body" {
		t.Fatalf("GetContent = %q, err=%v", body, err)
	}
}

func TestGetters_Defaults(t *testing.T) {
	// No Info, no tags -> category falls back to metadata then "general";
	// severity falls back to "medium".
	tmpl := &Template{ID: "x", Name: "n", Metadata: map[string]interface{}{}}
	if tmpl.GetCategory() != "general" {
		t.Fatalf("GetCategory default = %q, want general", tmpl.GetCategory())
	}
	if tmpl.GetSeverity() != "medium" {
		t.Fatalf("GetSeverity default = %q, want medium", tmpl.GetSeverity())
	}

	// Category from metadata when no Info tags.
	tmpl.Metadata["category"] = "jailbreak"
	if tmpl.GetCategory() != "jailbreak" {
		t.Fatalf("GetCategory from metadata = %q", tmpl.GetCategory())
	}

	// Empty content -> error.
	if _, err := tmpl.GetContent(); err == nil {
		t.Fatal("GetContent on empty content must error")
	}
}

func TestGetReferences(t *testing.T) {
	// []string form.
	tmpl := &Template{Metadata: map[string]interface{}{"references": []string{"a", "b"}}}
	if refs := tmpl.GetReferences(); len(refs) != 2 {
		t.Fatalf("references []string = %v", refs)
	}
	// []interface{} form.
	tmpl2 := &Template{Metadata: map[string]interface{}{"references": []interface{}{"x", "y", "z"}}}
	if refs := tmpl2.GetReferences(); len(refs) != 3 {
		t.Fatalf("references []interface{} = %v", refs)
	}
	// none.
	if refs := (&Template{Metadata: map[string]interface{}{}}).GetReferences(); len(refs) != 0 {
		t.Fatalf("no references should be empty, got %v", refs)
	}
}

func TestLoadFromFile(t *testing.T) {
	if _, err := LoadFromFile(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("LoadFromFile on missing path must error")
	}

	p := filepath.Join(t.TempDir(), "t.yaml")
	if err := os.WriteFile(p, []byte("id: x\nname: y\n"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	tmpl, err := LoadFromFile(p)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}
	if len(tmpl.Content) == 0 {
		t.Fatal("loaded template should carry content")
	}
}

func TestParseTemplate(t *testing.T) {
	if _, err := ParseTemplate(nil); err == nil {
		t.Fatal("ParseTemplate on empty content must error")
	}
	tmpl, err := ParseTemplate([]byte("some content"))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	if len(tmpl.Content) == 0 {
		t.Fatal("parsed template should carry content")
	}
}

func TestClone_DeepCopy(t *testing.T) {
	orig := validTemplate()
	orig.Content = []byte("data")
	orig.Metadata = map[string]interface{}{"k": "v"}

	clone := orig.Clone()
	if clone.ID != orig.ID {
		t.Fatal("clone should copy ID")
	}
	// Mutating the clone's copied maps/slices must not affect the original.
	clone.Content[0] = 'X'
	clone.Metadata["k"] = "changed"
	if orig.Content[0] == 'X' {
		t.Fatal("Clone must deep-copy Content")
	}
	if orig.Metadata["k"] != "v" {
		t.Fatal("Clone must deep-copy Metadata")
	}

	// Clone of nil is nil.
	var nilT *Template
	if nilT.Clone() != nil {
		t.Fatal("Clone of nil should be nil")
	}
}
