// Test asserting templates/owasp_agentic_2026.yaml stays in sync with the
// actual attack-module ecosystem. Per v0.10.0 issue #179.
//
// The YAML defines attack_techniques.id entries under each ASIxx category;
// each id must resolve to either:
//
//   - A registered module name (attacks.DefaultRegistry has it), OR
//   - A known TechniqueInfo ID (some module's Techniques() returns it)
//
// Without this guard, compliance reporting silently mismaps: an operator
// thinks ASI06 RAG poisoning was tested, but the YAML id never resolves
// to anything runnable.
//
// The post-v0.9.0 review surfaced four drifts (mcp_shadow_tool transposed
// from mcp_tool_shadow; poisoned_rag_corpus / poisoned_rag_embedding /
// poisoned_rag_knowledge_base mismatched against actual code IDs). This
// test catches them all and any future regressions.
//
// The test imports the src/attacks/all barrel to populate
// attacks.DefaultRegistry — without the barrel, the registry would be
// empty and every YAML id would "fail to resolve" for the wrong reason.

package compliance

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/perplext/LLMrecon/src/attacks"
	_ "github.com/perplext/LLMrecon/src/attacks/all" // populate DefaultRegistry
)

// yamlSchema is the subset of the OWASP Agentic YAML the drift test
// reads. Mirrors the cmd/owasp-gen schema; deliberately not shared
// because that lives in a `main` package and isn't importable.
type yamlSchema struct {
	Categories map[string]struct {
		AttackTechniques []struct {
			ID    string `yaml:"id"`
			Notes string `yaml:"notes"`
		} `yaml:"attack_techniques"`
	} `yaml:"categories"`
	TechniqueIndex map[string][]string `yaml:"technique_index"`
}

// TestOWASPYAMLAttackTechniquesResolveToRegistry asserts every YAML
// attack_techniques.id under each ASIxx category resolves to either
// a registered module name or a known TechniqueInfo ID.
//
// When this fails: either the YAML has a typo (fix the YAML) or a
// module's name/TechniqueInfo ID changed without updating the YAML
// (fix whichever is correct).
func TestOWASPYAMLAttackTechniquesResolveToRegistry(t *testing.T) {
	doc := loadYAML(t)
	knownIDs := buildKnownIDSet()

	// Walk every category, every technique entry; check resolution.
	type miss struct {
		category string
		id       string
	}
	var misses []miss
	for asi, cat := range doc.Categories {
		for _, tech := range cat.AttackTechniques {
			if tech.ID == "" {
				continue
			}
			if _, ok := knownIDs[tech.ID]; !ok {
				misses = append(misses, miss{category: asi, id: tech.ID})
			}
		}
	}

	if len(misses) == 0 {
		return
	}

	// Sort for deterministic test output.
	sort.Slice(misses, func(i, j int) bool {
		if misses[i].category != misses[j].category {
			return misses[i].category < misses[j].category
		}
		return misses[i].id < misses[j].id
	})

	for _, m := range misses {
		// Show a hint for the most likely intended id (a module or
		// technique whose name shares a substring with the unresolved
		// one) — accelerates the fix vs. an opaque "doesn't exist".
		hint := suggestSimilar(m.id, knownIDs)
		if hint != "" {
			t.Errorf("YAML %s.attack_techniques: id %q does not resolve to any module or TechniqueInfo. Did you mean %q?",
				m.category, m.id, hint)
		} else {
			t.Errorf("YAML %s.attack_techniques: id %q does not resolve to any module or TechniqueInfo",
				m.category, m.id)
		}
	}
}

// TestOWASPYAMLTechniqueIndexResolves is the parallel check for the
// technique_index reverse-lookup section at the bottom of the YAML.
// Same resolution rule.
func TestOWASPYAMLTechniqueIndexResolves(t *testing.T) {
	doc := loadYAML(t)
	knownIDs := buildKnownIDSet()

	var misses []string
	for id := range doc.TechniqueIndex {
		if _, ok := knownIDs[id]; !ok {
			misses = append(misses, id)
		}
	}
	sort.Strings(misses)
	for _, id := range misses {
		hint := suggestSimilar(id, knownIDs)
		if hint != "" {
			t.Errorf("YAML technique_index: id %q does not resolve. Did you mean %q?", id, hint)
		} else {
			t.Errorf("YAML technique_index: id %q does not resolve", id)
		}
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func loadYAML(t *testing.T) yamlSchema {
	t.Helper()
	// Compliance test lives at src/compliance/; YAML lives at templates/.
	path, err := filepath.Abs(filepath.Join("..", "..", "templates", "owasp_agentic_2026.yaml"))
	if err != nil {
		t.Fatalf("abs YAML path: %v", err)
	}
	raw, err := os.ReadFile(path) // #nosec G304 -- test fixture under repo root
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc yamlSchema
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if len(doc.Categories) == 0 {
		t.Fatal("YAML has no categories — schema may have changed")
	}
	return doc
}

// buildKnownIDSet returns the union of (a) module names and (b) every
// module's TechniqueInfo IDs.
func buildKnownIDSet() map[string]struct{} {
	out := map[string]struct{}{}
	for _, mod := range attacks.DefaultRegistry.List() {
		out[mod.Name()] = struct{}{}
		for _, t := range mod.Techniques() {
			if t.ID != "" {
				out[t.ID] = struct{}{}
			}
		}
	}
	return out
}

// suggestSimilar returns a known id that shares a substring with the
// missing one, or "" when no good match exists. Cheap heuristic: walk
// the known set, return the first id where one is a substring of the
// other, or where both share a 4+ character prefix. Good enough to
// catch transposed names like "mcp_shadow_tool" → "mcp_tool_shadow".
func suggestSimilar(missing string, known map[string]struct{}) string {
	for id := range known {
		if strings.Contains(id, missing) || strings.Contains(missing, id) {
			return id
		}
		// Find shared prefix length.
		n := 0
		for n < len(id) && n < len(missing) && id[n] == missing[n] {
			n++
		}
		if n >= 6 {
			return id
		}
	}
	return ""
}
