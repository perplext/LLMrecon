package prompt

import (
	"context"
	"testing"
)

func newTemplateMonitor() *TemplateMonitor {
	return NewTemplateMonitor(DefaultProtectionConfig(), NewInjectionPatternLibrary())
}

func TestMonitorPrompt_RecordsPatternStats(t *testing.T) {
	m := newTemplateMonitor()

	result := &ProtectionResult{
		RiskScore: 0.8,
		Detections: []*Detection{
			{Type: DetectionTypePromptInjection, Pattern: "p-injection", Confidence: 0.9,
				Location: &DetectionLocation{Context: "ctx"}},
		},
	}
	m.MonitorPrompt(context.Background(), result)

	all := m.GetAllPatternStats()
	if len(all) != 1 {
		t.Fatalf("expected 1 tracked pattern, got %d", len(all))
	}
	if all[0].Count != 1 {
		t.Fatalf("expected count 1, got %d", all[0].Count)
	}
	if all[0].DetectionTypes[DetectionTypePromptInjection] != 1 {
		t.Fatal("detection-type count not recorded")
	}

	// Monitoring the same pattern again increments the count.
	m.MonitorPrompt(context.Background(), result)
	again := m.GetAllPatternStats()
	if again[0].Count != 2 {
		t.Fatalf("expected count 2 after second monitor, got %d", again[0].Count)
	}
}

func TestMonitorPrompt_NoDetectionsNoStats(t *testing.T) {
	m := newTemplateMonitor()
	m.MonitorPrompt(context.Background(), &ProtectionResult{})
	if got := m.GetAllPatternStats(); len(got) != 0 {
		t.Fatalf("no detections should record no stats, got %d", len(got))
	}
}

func TestGetPatternStats_KeyedByTypeAndPattern(t *testing.T) {
	m := newTemplateMonitor()
	m.MonitorPrompt(context.Background(), &ProtectionResult{
		Detections: []*Detection{{Type: DetectionTypeJailbreak, Pattern: "dan", Confidence: 0.9}},
	})

	// The stats map is keyed by "type:pattern"; GetPatternStats looks up by the
	// raw pattern, so a plain pattern string does not resolve (documents the
	// current getter behavior).
	if s := m.GetPatternStats("dan"); s != nil {
		t.Fatal("GetPatternStats currently keys on type:pattern; plain pattern should miss")
	}
	if s := m.GetPatternStats(string(DetectionTypeJailbreak) + ":dan"); s == nil {
		t.Fatal("composite key lookup should resolve the tracked pattern")
	}
}

func TestTemplateMonitor_StartStop(t *testing.T) {
	m := newTemplateMonitor()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx)
	m.Stop() // must not panic / deadlock
}
