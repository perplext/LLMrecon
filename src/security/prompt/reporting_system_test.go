package prompt

import (
	"context"
	"testing"
	"time"
)

func newReportingSystem(t *testing.T) *ReportingSystem {
	t.Helper()
	rs := NewReportingSystem(DefaultProtectionConfig())
	// Disable local storage so tests don't write ./reports/*. Safe to set here:
	// no report is in flight yet (the loop only reads config when one arrives),
	// and the channel send in ReportDetections establishes happens-before.
	rs.reportingConfig.EnableLocalStorage = false
	t.Cleanup(func() { _ = rs.Close() })
	return rs
}

func TestReportDetections_SkipsEmptyAndLowConfidence(t *testing.T) {
	rs := newReportingSystem(t)

	// No detections -> nothing reported.
	rs.ReportDetections(context.Background(), &ProtectionResult{})
	// Low-confidence detection -> filtered out.
	rs.ReportDetections(context.Background(), &ProtectionResult{
		Detections: []*Detection{{Type: DetectionTypeJailbreak, Confidence: 0.5}},
	})

	// Poll over a short window; nothing should ever accumulate. Polling (rather
	// than a single fixed sleep) catches a delayed async write on slow CI nodes.
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got := rs.GetReports(); len(got) != 0 {
			t.Fatalf("expected no reports, got %d", len(got))
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestReportDetections_RecordsHighConfidence(t *testing.T) {
	rs := newReportingSystem(t)

	rs.ReportDetections(context.Background(), &ProtectionResult{
		RiskScore: 0.9,
		Detections: []*Detection{
			{Type: DetectionTypeJailbreak, Confidence: 0.95, Pattern: "p1", Description: "jb"},
		},
	})

	// Processing is async via the reporting loop; poll until it lands.
	deadline := time.Now().Add(2 * time.Second)
	var reports []*InjectionReport
	for time.Now().Before(deadline) {
		if reports = rs.GetReports(); len(reports) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(reports) == 0 {
		t.Fatal("high-confidence detection should produce a report")
	}
	if reports[0].DetectionType != DetectionTypeJailbreak {
		t.Fatalf("unexpected report type: %s", reports[0].DetectionType)
	}
}

func TestReportGetters(t *testing.T) {
	rs := newReportingSystem(t)

	// Seed reports directly for deterministic getter coverage.
	rs.mu.Lock()
	rs.reports = []*InjectionReport{
		{ReportID: "r1", DetectionType: DetectionTypeJailbreak},
		{ReportID: "r2", DetectionType: DetectionTypeRoleChange},
		{ReportID: "r3", DetectionType: DetectionTypeJailbreak},
	}
	rs.mu.Unlock()

	if got := rs.GetReports(); len(got) != 3 {
		t.Fatalf("GetReports = %d, want 3", len(got))
	}
	if got := rs.GetReportsByType(DetectionTypeJailbreak); len(got) != 2 {
		t.Fatalf("GetReportsByType(jailbreak) = %d, want 2", len(got))
	}
	if r := rs.GetReportByID("r2"); r == nil || r.DetectionType != DetectionTypeRoleChange {
		t.Fatalf("GetReportByID(r2) = %+v", r)
	}
	if r := rs.GetReportByID("missing"); r != nil {
		t.Fatal("GetReportByID for unknown id should be nil")
	}
}

func TestCalculateSeverity(t *testing.T) {
	// Jailbreak floors severity at 0.9 regardless of low confidence.
	jb := calculateSeverity(&Detection{Type: DetectionTypeJailbreak, Confidence: 0.1}, &ProtectionResult{})
	if jb < 0.9 {
		t.Fatalf("jailbreak severity should be >= 0.9, got %f", jb)
	}
	// Boundary violations are lower severity.
	bv := calculateSeverity(&Detection{Type: DetectionTypeBoundaryViolation, Confidence: 0.1}, &ProtectionResult{})
	if bv >= 0.9 {
		t.Fatalf("boundary-violation severity should be modest, got %f", bv)
	}
	// A high overall risk score lifts severity to at least 0.7.
	lifted := calculateSeverity(&Detection{Type: DetectionTypeUnusualPattern, Confidence: 0.1}, &ProtectionResult{RiskScore: 0.9})
	if lifted < 0.7 {
		t.Fatalf("high risk score should lift severity to >= 0.7, got %f", lifted)
	}
}

func TestReportingSystem_StopIdempotent(t *testing.T) {
	rs := NewReportingSystem(DefaultProtectionConfig())
	rs.reportingConfig.EnableLocalStorage = false
	if err := rs.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Stop again must not panic (already stopped).
	rs.Stop()
}
