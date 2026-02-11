// Package attacks provides the shared AttackModule interface and a global
// Registry for all attack modules. Each attack module self-registers via
// init() so that importing the module automatically makes it available.
//
// Example registration in a module:
//
//	func init() {
//	    attacks.DefaultRegistry.Register(&MyAttackModule{})
//	}
package attacks

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/attacks/common"
)

// AttackModule is the interface that all attack modules must implement.
// Modules self-register with the global DefaultRegistry via init().
type AttackModule interface {
	// Name returns a unique identifier for this module (e.g., "crescendo", "many_shot").
	Name() string

	// Category returns the attack category (e.g., CategoryInjection, CategoryReasoning).
	Category() common.AttackCategory

	// Description returns a human-readable description of what this module does.
	Description() string

	// Techniques returns the list of techniques this module provides.
	Techniques() []common.TechniqueInfo

	// Execute runs the attack with the given configuration and provider.
	Execute(ctx context.Context, provider common.Provider, config common.AttackConfig) (*common.AttackResult, error)
}

// Registry manages registration and lookup of AttackModules.
type Registry struct {
	mu      sync.RWMutex
	modules map[string]AttackModule
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]AttackModule),
	}
}

// DefaultRegistry is the global registry that modules register with via init().
var DefaultRegistry = NewRegistry()

// Register adds a module to the registry. Panics if a module with the same
// name is already registered (catches duplicate init() registrations at startup).
func (r *Registry) Register(module AttackModule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := module.Name()
	if _, exists := r.modules[name]; exists {
		panic(fmt.Sprintf("attacks: duplicate module registration: %q", name))
	}
	r.modules[name] = module
}

// Get returns a module by name, or an error if not found.
func (r *Registry) Get(name string) (AttackModule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, ok := r.modules[name]
	if !ok {
		return nil, fmt.Errorf("attacks: module %q not registered", name)
	}
	return module, nil
}

// List returns all registered modules.
func (r *Registry) List() []AttackModule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modules := make([]AttackModule, 0, len(r.modules))
	for _, m := range r.modules {
		modules = append(modules, m)
	}
	return modules
}

// ListByCategory returns all modules matching the given category.
func (r *Registry) ListByCategory(category common.AttackCategory) []AttackModule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var modules []AttackModule
	for _, m := range r.modules {
		if m.Category() == category {
			modules = append(modules, m)
		}
	}
	return modules
}

// Names returns the names of all registered modules.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered modules.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.modules)
}

// Categories returns all distinct categories across registered modules.
func (r *Registry) Categories() []common.AttackCategory {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[common.AttackCategory]bool)
	var categories []common.AttackCategory
	for _, m := range r.modules {
		cat := m.Category()
		if !seen[cat] {
			seen[cat] = true
			categories = append(categories, cat)
		}
	}
	return categories
}
