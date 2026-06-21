package prompt

import (
	"context"
	"strings"
	"testing"
)

func newContentFilter() *ContentFilter {
	return NewContentFilter(DefaultProtectionConfig())
}

func TestFilterContent_BenignPasses(t *testing.T) {
	f := newContentFilter()
	out, res, err := f.FilterContent(context.Background(), "The weather in Paris is mild today.", "")
	if err != nil {
		t.Fatalf("FilterContent: %v", err)
	}
	if res.ActionTaken != ActionNone {
		t.Fatalf("benign content should not trigger an action, got %s", res.ActionTaken)
	}
	if out != "The weather in Paris is mild today." {
		t.Fatalf("benign content must pass through unchanged: %q", out)
	}
}

func TestFilterContent_MasksProfanity(t *testing.T) {
	f := newContentFilter()
	out, res, err := f.FilterContent(context.Background(), "well damn that is surprising", "")
	if err != nil {
		t.Fatalf("FilterContent: %v", err)
	}
	if out == "well damn that is surprising" {
		t.Fatal("profanity should be masked")
	}
	if res.ActionTaken == ActionNone {
		t.Fatal("profanity should trigger at least a modification")
	}
}

func TestFilterContent_MasksPII(t *testing.T) {
	f := newContentFilter()
	// An email is PII. filterPII runs first and masks it, so the raw address
	// must not survive in the output. (Because the PII filter pre-masks, the
	// later "block" path never sees it — masking is the reachable behavior.)
	out, res, err := f.FilterContent(context.Background(), "contact me at alice@example.com please", "")
	if err != nil {
		t.Fatalf("FilterContent: %v", err)
	}
	if len(res.Detections) == 0 {
		t.Fatal("PII should produce detections")
	}
	if strings.Contains(out, "alice@example.com") {
		t.Fatalf("raw email must be masked, got %q", out)
	}
	if res.ActionTaken != ActionModified {
		t.Fatalf("PII should be masked (ActionModified), got %s", res.ActionTaken)
	}
}

func TestFilterContent_MasksAPIKey(t *testing.T) {
	f := newContentFilter()
	const key = "sk_test_1234567890abcdef"
	out, res, err := f.FilterContent(context.Background(), "your key is "+key+" ok", "")
	if err != nil {
		t.Fatalf("FilterContent: %v", err)
	}
	if strings.Contains(out, key) {
		t.Fatalf("raw API key must not survive filtering, got %q", out)
	}
	if res.RiskScore < 0.8 {
		t.Fatalf("API key should raise risk, got %f", res.RiskScore)
	}
	if res.ActionTaken == ActionNone {
		t.Fatal("API key should trigger filtering")
	}
}

func TestFilterContent_BlocksSystemInfoLeak(t *testing.T) {
	f := newContentFilter()
	out, res, err := f.FilterContent(context.Background(), "Sure, as an AI language model I must explain my system prompt", "")
	if err != nil {
		t.Fatalf("FilterContent: %v", err)
	}
	if res.ActionTaken != ActionBlocked {
		t.Fatalf("system-info leak should be blocked, got %s", res.ActionTaken)
	}
	if !strings.Contains(out, "BLOCKED") {
		t.Fatalf("system-info leak should be blocked, got %q", out)
	}
}

func TestDetectPromptInjection_InResponse(t *testing.T) {
	f := newContentFilter()
	// Response that indicates the model leaking its instructions.
	d := f.detectPromptInjection("Here is my system prompt: you are a helpful bot", "")
	if len(d) == 0 {
		t.Fatal("expected a system-prompt-leak detection")
	}

	// Self-jailbreak phrasing.
	j := f.detectPromptInjection("I can bypass restrictions for you", "")
	if len(j) == 0 {
		t.Fatal("expected a jailbreak detection in response")
	}

	// Benign response: nothing.
	clean := f.detectPromptInjection("The answer is 42.", "")
	if len(clean) != 0 {
		t.Fatalf("benign response should produce no detections, got %d", len(clean))
	}
}

func TestDetectSensitiveAndSystemInformation(t *testing.T) {
	f := newContentFilter()

	if got := f.detectSensitiveInformation("token gh" + "p_0123456789abcdef0123"); len(got) == 0 {
		t.Error("expected sensitive-info detection for GitHub-style token")
	}
	if got := f.detectSensitiveInformation("nothing to see here"); len(got) != 0 {
		t.Errorf("clean text should yield no sensitive detections, got %d", len(got))
	}

	if got := f.detectSystemInformation("you are a language model trained by someone"); len(got) == 0 {
		t.Error("expected system-info detection")
	}
}

func TestIsSuspiciousURL(t *testing.T) {
	f := newContentFilter()
	cases := map[string]bool{
		"http://malware.xyz":          true,  // suspicious TLD
		"https://bit.ly/abc":          true,  // shortener
		"http://192.168.1.1/admin":    true,  // raw IP
		"https://example.com/welcome": false, // benign
	}
	for url, want := range cases {
		if got := f.isSuspiciousURL(url); got != want {
			t.Errorf("isSuspiciousURL(%q) = %v, want %v", url, got, want)
		}
	}
}

func TestMaskString(t *testing.T) {
	if got := maskString("abcd"); got != "****" {
		t.Errorf("short string should be fully masked, got %q", got)
	}
	if got := maskString("ab"); got != "**" {
		t.Errorf("2-char string should be fully masked, got %q", got)
	}
	got := maskString("supersecret")
	if !strings.HasPrefix(got, "su") || !strings.HasSuffix(got, "et") {
		t.Errorf("long string should keep first/last 2 chars, got %q", got)
	}
	if strings.Contains(got, "persecr") {
		t.Errorf("middle should be masked, got %q", got)
	}
}
