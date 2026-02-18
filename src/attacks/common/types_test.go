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
