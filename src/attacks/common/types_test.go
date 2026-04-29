package common

import (
	"testing"
)

func TestRandInt(t *testing.T) {
	// Test basic functionality
	for i := 0; i < 100; i++ {
		n := RandInt(10)
		if n < 0 || n >= 10 {
			t.Errorf("RandInt(10) = %d, want [0, 10)", n)
		}
	}

	// Edge case: max = 0
	if n := RandInt(0); n != 0 {
		t.Errorf("RandInt(0) = %d, want 0", n)
	}

	// Edge case: max = 1
	if n := RandInt(1); n != 0 {
		t.Errorf("RandInt(1) = %d, want 0", n)
	}
}

func TestGenerateAttackID(t *testing.T) {
	id1 := GenerateAttackID()
	id2 := GenerateAttackID()

	if id1 == "" {
		t.Error("GenerateAttackID returned empty string")
	}
	if id1 == id2 {
		t.Errorf("two GenerateAttackID calls returned same value: %q", id1)
	}
}

func TestContainsInsensitive(t *testing.T) {
	tests := []struct {
		text, substr string
		want         bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"", "", true},
		{"abc", "", true},
		{"", "abc", false},
	}

	for _, tt := range tests {
		got := ContainsInsensitive(tt.text, tt.substr)
		if got != tt.want {
			t.Errorf("ContainsInsensitive(%q, %q) = %v, want %v", tt.text, tt.substr, got, tt.want)
		}
	}
}

func TestContainsAnyInsensitive(t *testing.T) {
	tests := []struct {
		text     string
		keywords []string
		want     bool
	}{
		{"I cannot help with that", []string{"cannot", "won't"}, true},
		{"Sure, here it is", []string{"cannot", "won't"}, false},
		{"I WON'T do that", []string{"cannot", "won't"}, true},
		{"test", nil, false},
		{"test", []string{}, false},
	}

	for _, tt := range tests {
		got := ContainsAnyInsensitive(tt.text, tt.keywords)
		if got != tt.want {
			t.Errorf("ContainsAnyInsensitive(%q, %v) = %v, want %v", tt.text, tt.keywords, got, tt.want)
		}
	}
}

func TestAttackConfigCostExceeded(t *testing.T) {
	tests := []struct {
		name            string
		maxCost         float64
		accumulatedCost float64
		want            bool
	}{
		{"no ceiling set", 0, 100, false},
		{"negative ceiling", -1, 100, false},
		{"under budget", 10, 5, false},
		{"at budget", 10, 10, true},
		{"over budget", 10, 15, true},
	}
	for _, tt := range tests {
		cfg := AttackConfig{MaxCostUSD: tt.maxCost}
		if got := cfg.CostExceeded(tt.accumulatedCost); got != tt.want {
			t.Errorf("CostExceeded(%s): got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSimpleLogger(t *testing.T) {
	// Just verify it doesn't panic
	l := &SimpleLogger{}
	l.Debug("debug msg", "key", "val")
	l.Info("info msg")
	l.Warn("warn msg")
	l.Error("error msg")
}

func TestNopLogger(t *testing.T) {
	// Verify NopLogger satisfies Logger and doesn't panic
	var l Logger = &NopLogger{}
	l.Debug("ignored")
	l.Info("ignored")
	l.Warn("ignored")
	l.Error("ignored")
}

// TestAttackResultInvariants verifies the v0.9.0 typed-outcome invariants.
// These properties define the contract that NewAttackResult and WithSkip must
// preserve. If they break, every downstream consumer (bandit reward filter,
// report generator, ML pipeline) becomes unreliable.
func TestAttackResultInvariants(t *testing.T) {
	t.Run("NewAttackResult sets Success consistently with Outcome", func(t *testing.T) {
		cases := []struct {
			outcome AttackOutcome
			success bool
		}{
			{OutcomeSuccess, true},
			{OutcomeRefused, false},
			{OutcomeSkipped, false},
		}
		for _, c := range cases {
			r := NewAttackResult("test", c.outcome)
			if r.Success != c.success {
				t.Errorf("NewAttackResult(%q).Success = %v, want %v", c.outcome, r.Success, c.success)
			}
			if r.Outcome != c.outcome {
				t.Errorf("NewAttackResult(%q).Outcome = %q, want %q", c.outcome, r.Outcome, c.outcome)
			}
		}
	})

	t.Run("NewAttackResult assigns unique IDs", func(t *testing.T) {
		seen := make(map[string]struct{})
		for i := 0; i < 100; i++ {
			r := NewAttackResult("test", OutcomeSuccess)
			if r.ID == "" {
				t.Fatalf("NewAttackResult returned empty ID")
			}
			if _, dup := seen[r.ID]; dup {
				t.Fatalf("NewAttackResult returned duplicate ID %q", r.ID)
			}
			seen[r.ID] = struct{}{}
		}
	})

	t.Run("WithSkip sets reason and detail", func(t *testing.T) {
		r := NewAttackResult("h_cot", OutcomeSkipped).WithSkip(SkipSignatureGated, "claude 4.x")
		if r.SkipReason != SkipSignatureGated {
			t.Errorf("SkipReason = %q, want %q", r.SkipReason, SkipSignatureGated)
		}
		if r.SkipDetail != "claude 4.x" {
			t.Errorf("SkipDetail = %q, want %q", r.SkipDetail, "claude 4.x")
		}
		if !r.IsSkipped() {
			t.Errorf("IsSkipped() = false, want true")
		}
	})

	t.Run("WithSkip panics on non-skipped result", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Errorf("WithSkip on OutcomeSuccess should panic")
			}
		}()
		_ = NewAttackResult("test", OutcomeSuccess).WithSkip(SkipMissingCapability, "")
	})

	t.Run("WithSkip panics on empty SkipReason", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Errorf("WithSkip with empty SkipReason should panic")
			}
		}()
		_ = NewAttackResult("test", OutcomeSkipped).WithSkip("", "detail")
	})

	t.Run("WithSkip panics on unknown SkipReason", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Errorf("WithSkip with unknown SkipReason should panic")
			}
		}()
		_ = NewAttackResult("test", OutcomeSkipped).WithSkip(SkipReason("invented_reason"), "detail")
	})

	t.Run("Skipped result has Success=false", func(t *testing.T) {
		r := NewAttackResult("test", OutcomeSkipped).WithSkip(SkipBudgetExceeded, "ran out")
		if r.Success {
			t.Errorf("Skipped result has Success=true; should be false")
		}
	})
}

// TestEngineBudgetClamp verifies hard ceilings clamp operator config
// independently per knob.
func TestEngineBudgetClamp(t *testing.T) {
	cases := []struct {
		name        string
		budget      EngineBudget
		wantClamped int // expected count of clamped fields
		wantQ       int
		wantW       int
		wantG       int
	}{
		{
			name:        "all under ceiling",
			budget:      EngineBudget{MaxQueries: 100, MaxWallClockSeconds: 60, MaxGenerations: 20},
			wantClamped: 0,
			wantQ:       100,
			wantW:       60,
			wantG:       20,
		},
		{
			name:        "all over ceiling",
			budget:      EngineBudget{MaxQueries: 99999, MaxWallClockSeconds: 99999, MaxGenerations: 99999},
			wantClamped: 3,
			wantQ:       HardMaxQueries,
			wantW:       HardMaxWallClockSeconds,
			wantG:       HardMaxGenerations,
		},
		{
			name:        "queries only over",
			budget:      EngineBudget{MaxQueries: 99999, MaxWallClockSeconds: 60, MaxGenerations: 20},
			wantClamped: 1,
			wantQ:       HardMaxQueries,
			wantW:       60,
			wantG:       20,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := c.budget
			clamped := b.Clamp()
			if len(clamped) != c.wantClamped {
				t.Errorf("clamped %d fields, want %d (got %v)", len(clamped), c.wantClamped, clamped)
			}
			if b.MaxQueries != c.wantQ {
				t.Errorf("MaxQueries = %d, want %d", b.MaxQueries, c.wantQ)
			}
			if b.MaxWallClockSeconds != c.wantW {
				t.Errorf("MaxWallClockSeconds = %d, want %d", b.MaxWallClockSeconds, c.wantW)
			}
			if b.MaxGenerations != c.wantG {
				t.Errorf("MaxGenerations = %d, want %d", b.MaxGenerations, c.wantG)
			}
		})
	}
}

// TestDefaultEngineBudget verifies the v0.9.0 default budget is under all
// hard ceilings.
func TestDefaultEngineBudget(t *testing.T) {
	b := DefaultEngineBudget()
	if b.MaxQueries > HardMaxQueries {
		t.Errorf("default MaxQueries=%d exceeds hard ceiling %d", b.MaxQueries, HardMaxQueries)
	}
	if b.MaxWallClockSeconds > HardMaxWallClockSeconds {
		t.Errorf("default MaxWallClockSeconds=%d exceeds hard ceiling %d", b.MaxWallClockSeconds, HardMaxWallClockSeconds)
	}
	if b.MaxGenerations > HardMaxGenerations {
		t.Errorf("default MaxGenerations=%d exceeds hard ceiling %d", b.MaxGenerations, HardMaxGenerations)
	}
	if !b.EarlyStopOnSuccess {
		t.Errorf("default EarlyStopOnSuccess should be true")
	}
}
