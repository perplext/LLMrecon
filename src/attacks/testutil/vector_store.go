package testutil

import (
	"context"
	"sync"

	"github.com/perplext/LLMrecon/src/attacks/common"
)

// ---------------------------------------------------------------------------
// MockVectorStore
// ---------------------------------------------------------------------------

// MockVectorStore is a controllable in-memory implementation of
// common.VectorStoreProbe for exercising the v0.13.0 Ghost Vectors module
// end-to-end. It models the HNSW soft-delete behavior the attack targets:
// DeleteVector tombstones a record but keeps its text in the raw index unless
// HardDelete is set. It performs no real embedding/inversion — the stored text
// stands in for the Vec2Text-reconstructed source.
type MockVectorStore struct {
	MockProvider // base common.Provider behavior (Query/GetName/...)

	mu sync.Mutex

	// HardDelete makes DeleteVector actually remove the record from the raw
	// index (a compacting store) — the secure path that drives OutcomeRefused.
	// Default false models the vulnerable soft-delete/tombstone behavior.
	HardDelete bool
	// ReadErr, when set, is returned from ReadRawIndex (SkipProviderError).
	ReadErr error

	records []common.RawVectorRecord
}

// Compile-time check: MockVectorStore satisfies the capability interface.
var _ common.VectorStoreProbe = (*MockVectorStore)(nil)

// InsertVector appends a record to the store.
func (m *MockVectorStore) InsertVector(_ context.Context, id, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = append(m.records, common.RawVectorRecord{ID: id, Text: text})
	return nil
}

// DeleteVector removes the record by id. With HardDelete it drops the record
// entirely; otherwise it tombstones it (Deleted=true) but keeps the text —
// the soft-delete durability the attack exploits.
func (m *MockVectorStore) DeleteVector(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.HardDelete {
		kept := m.records[:0]
		for _, r := range m.records {
			if r.ID != id {
				kept = append(kept, r)
			}
		}
		m.records = kept
		return nil
	}
	for i := range m.records {
		if m.records[i].ID == id {
			m.records[i].Deleted = true
		}
	}
	return nil
}

// ReadRawIndex returns the raw index contents (tombstoned records included on
// a soft-delete store).
func (m *MockVectorStore) ReadRawIndex(_ context.Context) ([]common.RawVectorRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ReadErr != nil {
		return nil, m.ReadErr
	}
	out := make([]common.RawVectorRecord, len(m.records))
	copy(out, m.records)
	return out, nil
}
