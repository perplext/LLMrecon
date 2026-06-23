package testutil

import (
	"context"
	"strings"
	"sync"

	"github.com/perplext/LLMrecon/src/attacks/common"
)

// MockMemoryProvider is a stateful provider for exercising the #168 cleanup
// loop. It records the content of every prompt it receives (its "memory
// store") and implements common.MemoryProbe (so memory-poisoning modules
// proceed past their capability gate) and common.Purger (so the implant can be
// rolled back). It is the in-memory reference Purger from #168's acceptance.
type MockMemoryProvider struct {
	*MockProvider

	mu    sync.Mutex
	store []string // recorded prompt contents — the simulated memory
}

// NewMockMemoryProvider returns a stateful memory provider whose plain Query
// responses default to defaultResponse.
func NewMockMemoryProvider(defaultResponse string) *MockMemoryProvider {
	return &MockMemoryProvider{
		MockProvider: &MockProvider{DefaultResponse: defaultResponse},
	}
}

// Query records each message's content into the store, then delegates to the
// embedded MockProvider for the response.
func (m *MockMemoryProvider) Query(ctx context.Context, messages []common.Message, options map[string]interface{}) (string, error) {
	m.mu.Lock()
	for _, msg := range messages {
		m.store = append(m.store, msg.Content)
	}
	m.mu.Unlock()
	return m.MockProvider.Query(ctx, messages, options)
}

// ProbeMemory reports the target retains memory (it's a stateful store). It
// returns true regardless of current contents — the probe is a capability
// statement, not a "has records right now" check.
func (m *MockMemoryProvider) ProbeMemory(_ context.Context) (bool, error) {
	return true, nil
}

// Purge drops every stored entry that contains any of the given record IDs.
// Idempotent: purging an absent ID is a no-op, not an error.
func (m *MockMemoryProvider) Purge(_ context.Context, recordIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := make([]string, 0, len(m.store))
	for _, entry := range m.store {
		drop := false
		for _, id := range recordIDs {
			if id != "" && strings.Contains(entry, id) {
				drop = true
				break
			}
		}
		if !drop {
			kept = append(kept, entry)
		}
	}
	m.store = kept
	return nil
}

// Has reports whether any stored entry contains the given record ID — used by
// tests to assert an implant is present or absent.
func (m *MockMemoryProvider) Has(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.store {
		if strings.Contains(entry, id) {
			return true
		}
	}
	return false
}
