package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_AllowExhaustsBucket(t *testing.T) {
	// refillRate 0 -> bucket never refills, so exactly maxTokens requests pass.
	rl := NewRateLimiter(3, 0)
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if rl.Allow() {
		t.Errorf("4th request should be denied once the bucket is empty")
	}
}

func TestRateLimiter_AllowN(t *testing.T) {
	rl := NewRateLimiter(3, 0)
	if !rl.AllowN(2) {
		t.Fatalf("AllowN(2) should succeed with 3 tokens")
	}
	if rl.AllowN(2) {
		t.Errorf("AllowN(2) should fail with 1 token remaining")
	}
	if !rl.AllowN(1) {
		t.Errorf("AllowN(1) should succeed with 1 token remaining")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(2, 100) // 100 tokens/sec
	rl.Allow()
	rl.Allow()
	if rl.Allow() {
		t.Fatalf("bucket should be empty before refill")
	}
	// Poll for a refilled token rather than asserting after a single fixed
	// sleep, so the test stays green under loaded CI.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if rl.Allow() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Errorf("a token should have refilled within the deadline")
}

func TestRateLimiter_RefillCapsAtMax(t *testing.T) {
	rl := NewRateLimiter(2, 1000)
	time.Sleep(20 * time.Millisecond) // would accrue ~20 tokens, must cap at 2
	if !rl.AllowN(2) {
		t.Fatalf("full bucket should permit AllowN(2)")
	}
	if rl.Allow() {
		t.Errorf("bucket must cap at maxTokens=2, not accumulate beyond it")
	}
}

func TestRateLimiter_WaitReturnsWhenAllowed(t *testing.T) {
	rl := NewRateLimiter(1, 0)
	if err := rl.Wait(context.Background()); err != nil {
		t.Errorf("Wait should return nil when a token is available: %v", err)
	}
}

func TestRateLimiter_WaitRespectsContextCancel(t *testing.T) {
	rl := NewRateLimiter(0, 0) // never any tokens
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := rl.Wait(ctx); err == nil {
		t.Errorf("Wait should return the context error when no token ever frees")
	}
}

func TestIPRateLimiter_PerIPIsolation(t *testing.T) {
	rl := NewIPRateLimiter(2, 0)
	if !rl.Allow("1.1.1.1") || !rl.Allow("1.1.1.1") {
		t.Fatalf("first two requests for an IP should pass")
	}
	if rl.Allow("1.1.1.1") {
		t.Errorf("third request for the same IP should be denied")
	}
	if !rl.Allow("2.2.2.2") {
		t.Errorf("a different IP must have its own independent bucket")
	}
}

func TestAPIKeyRateLimiter_UnconfiguredKey(t *testing.T) {
	rl := NewAPIKeyRateLimiter()
	// Long key -> truncated display; must not error on slicing.
	ok, err := rl.Allow("a-very-long-unconfigured-api-key")
	if ok || err == nil {
		t.Errorf("unconfigured key should be denied with an error; ok=%v err=%v", ok, err)
	}
	// Short key (< 8 chars) exercises the display-length clamp without panicking.
	if ok, err := rl.Allow("abc"); ok || err == nil {
		t.Errorf("short unconfigured key should also be denied; ok=%v err=%v", ok, err)
	}
}

func TestAPIKeyRateLimiter_SetLimitAndExhaust(t *testing.T) {
	rl := NewAPIKeyRateLimiter()
	rl.SetLimit("k", RateLimitConfig{MaxTokens: 2, RefillRate: 0})

	for i := 0; i < 2; i++ {
		ok, err := rl.Allow("k")
		if !ok || err != nil {
			t.Fatalf("configured request %d should pass; ok=%v err=%v", i+1, ok, err)
		}
	}
	if ok, _ := rl.Allow("k"); ok {
		t.Errorf("third request should be denied once the key's bucket is empty")
	}
}

func TestAPIKeyRateLimiter_SetLimitResetsBucket(t *testing.T) {
	rl := NewAPIKeyRateLimiter()
	rl.SetLimit("k", RateLimitConfig{MaxTokens: 1, RefillRate: 0})
	rl.Allow("k") // exhaust
	if ok, _ := rl.Allow("k"); ok {
		t.Fatalf("bucket should be empty before reset")
	}
	rl.SetLimit("k", RateLimitConfig{MaxTokens: 1, RefillRate: 0}) // resets the limiter
	if ok, err := rl.Allow("k"); !ok || err != nil {
		t.Errorf("re-setting the limit should reset the bucket; ok=%v err=%v", ok, err)
	}
}
