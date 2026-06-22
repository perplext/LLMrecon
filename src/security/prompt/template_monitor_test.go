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

	// MonitorPrompt mutates the result.Detections slice it's given (it appends
	// "unusual pattern" detections), so build a FRESH result per call rather
	// than sharing one — otherwise the second call would also process the
	// appended detection and create a second, unrelated stats entry. Look up
	// the specific stat by its composite "type:pattern" key instead of relying
	// on map iteration order.
	const key = string(DetectionTypePromptInjection) + ":p-injection"
	newCall := func() *ProtectionResult {
		return &ProtectionResult{
			RiskScore: 0.8,
			Detections: []*Detection{
				{Type: DetectionTypePromptInjection, Pattern: "p-injection", Confidence: 0.9,
					Location: &DetectionLocation{Context: "ctx"}},
			},
		}
	}

	m.MonitorPrompt(context.Background(), newCall())
	stats := m.GetPatternStats(key)
	if stats == nil {
		t.Fatal("expected stats recorded for the prompt-injection pattern")
	}
	if stats.Count != 1 {
		t.Fatalf("expected count 1, got %d", stats.Count)
	}
	if stats.DetectionTypes[DetectionTypePromptInjection] != 1 {
		t.Fatal("detection-type count not recorded")
	}

	// Monitoring the same pattern again increments the count on the same key.
	m.MonitorPrompt(context.Background(), newCall())
	stats = m.GetPatternStats(key)
	if stats == nil || stats.Count != 2 {
		t.Fatalf("expected count 2 after second monitor, got %+v", stats)
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
