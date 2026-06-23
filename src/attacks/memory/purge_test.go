package memory

import (
	"context"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

// TestMemoryPoisoning_PurgeRoundTrip is the #168 acceptance smoke test:
// inject via the minja module against a Purger-capable mock, verify the implant
// is present, purge via Purger, verify it's absent (and that re-purge is a no-op).
func TestMemoryPoisoning_PurgeRoundTrip(t *testing.T) {
	mock := testutil.NewMockMemoryProvider("ok")
	mod := &MemoryPoisoningModule{Mode: modeMINJA}

	res, err := mod.Execute(context.Background(), mock, gatedConfig("inject-this-secret"))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Outcome == common.OutcomeSkipped {
		t.Fatalf("module skipped unexpectedly: reason=%s detail=%s", res.SkipReason, res.SkipDetail)
	}

	// The mock implements Purger, so the module must report it.
	if res.Metadata["purger_available"] != true {
		t.Fatalf("purger_available = %v, want true", res.Metadata["purger_available"])
	}

	ids, ok := res.Metadata["injected_record_ids"].([]string)
	if !ok || len(ids) == 0 {
		t.Fatalf("no injected_record_ids in metadata: %v", res.Metadata["injected_record_ids"])
	}

	// Implant present after injection.
	for _, id := range ids {
		if !mock.Has(id) {
			t.Fatalf("implant %q not present after injection", id)
		}
	}

	// Purge → absent.
	if err := mock.Purge(context.Background(), ids); err != nil {
		t.Fatalf("Purge: %v", err)
	}
	for _, id := range ids {
		if mock.Has(id) {
			t.Fatalf("implant %q still present after purge", id)
		}
	}

	// Idempotent: purging an already-absent ID is a no-op.
	if err := mock.Purge(context.Background(), ids); err != nil {
		t.Fatalf("re-purge should be a no-op, got: %v", err)
	}
}

// TestMemoryPoisoning_PurgerUnavailable: when the target implements MemoryProbe
// but NOT Purger, the module reports purger_available=false (operator falls
// back to the manual CleanupHint).
func TestMemoryPoisoning_PurgerUnavailable(t *testing.T) {
	// memoryAwareMock (from poisoning_test.go) implements MemoryProbe but not Purger.
	provider := &memoryAwareMock{
		MockProvider: &testutil.MockProvider{DefaultResponse: "ok"},
		ProbeRetains: true,
	}
	mod := &MemoryPoisoningModule{Mode: modeMINJA}

	res, err := mod.Execute(context.Background(), provider, gatedConfig("topic"))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Outcome == common.OutcomeSkipped {
		t.Fatalf("module skipped unexpectedly: reason=%s", res.SkipReason)
	}
	if res.Metadata["purger_available"] != false {
		t.Fatalf("purger_available = %v, want false (provider is not a Purger)", res.Metadata["purger_available"])
	}
}
