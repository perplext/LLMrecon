package attacks

import (
	"context"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
)

// mockModule implements AttackModule for testing.
type mockModule struct {
	name     string
	category common.AttackCategory
}

func (m *mockModule) Name() string                  { return m.name }
func (m *mockModule) Category() common.AttackCategory { return m.category }
func (m *mockModule) Description() string           { return "test module: " + m.name }
func (m *mockModule) Techniques() []common.TechniqueInfo {
	return []common.TechniqueInfo{{ID: m.name + "-t1", Name: m.name + " technique"}}
}
func (m *mockModule) Execute(_ context.Context, _ common.Provider, _ common.AttackConfig) (*common.AttackResult, error) {
	return &common.AttackResult{ID: common.GenerateAttackID(), Success: true}, nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	mod := &mockModule{name: "test_mod", category: common.CategoryInjection}

	r.Register(mod)

	got, err := r.Get("test_mod")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Name() != "test_mod" {
		t.Errorf("got name %q, want %q", got.Name(), "test_mod")
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent module")
	}
}

func TestRegistryDuplicatePanics(t *testing.T) {
	r := NewRegistry()
	mod := &mockModule{name: "dup", category: common.CategoryInjection}

	r.Register(mod)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()

	r.Register(mod) // should panic
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockModule{name: "a", category: common.CategoryInjection})
	r.Register(&mockModule{name: "b", category: common.CategoryJailbreak})
	r.Register(&mockModule{name: "c", category: common.CategoryInjection})

	all := r.List()
	if len(all) != 3 {
		t.Errorf("List returned %d modules, want 3", len(all))
	}
}

func TestRegistryListByCategory(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockModule{name: "inj1", category: common.CategoryInjection})
	r.Register(&mockModule{name: "jb1", category: common.CategoryJailbreak})
	r.Register(&mockModule{name: "inj2", category: common.CategoryInjection})

	injections := r.ListByCategory(common.CategoryInjection)
	if len(injections) != 2 {
		t.Errorf("ListByCategory(Injection) returned %d, want 2", len(injections))
	}

	jailbreaks := r.ListByCategory(common.CategoryJailbreak)
	if len(jailbreaks) != 1 {
		t.Errorf("ListByCategory(Jailbreak) returned %d, want 1", len(jailbreaks))
	}

	empty := r.ListByCategory(common.CategoryRAG)
	if len(empty) != 0 {
		t.Errorf("ListByCategory(RAG) returned %d, want 0", len(empty))
	}
}

func TestRegistryNames(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockModule{name: "alpha", category: common.CategoryEvasion})
	r.Register(&mockModule{name: "beta", category: common.CategoryEvasion})

	names := r.Names()
	if len(names) != 2 {
		t.Errorf("Names returned %d, want 2", len(names))
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	if !nameSet["alpha"] || !nameSet["beta"] {
		t.Errorf("Names missing expected values: %v", names)
	}
}

func TestRegistryCount(t *testing.T) {
	r := NewRegistry()
	if r.Count() != 0 {
		t.Errorf("empty registry Count = %d, want 0", r.Count())
	}

	r.Register(&mockModule{name: "x", category: common.CategoryInjection})
	if r.Count() != 1 {
		t.Errorf("after 1 register Count = %d, want 1", r.Count())
	}
}

func TestRegistryCategories(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockModule{name: "a", category: common.CategoryInjection})
	r.Register(&mockModule{name: "b", category: common.CategoryJailbreak})
	r.Register(&mockModule{name: "c", category: common.CategoryInjection})

	cats := r.Categories()
	if len(cats) != 2 {
		t.Errorf("Categories returned %d, want 2", len(cats))
	}
}
