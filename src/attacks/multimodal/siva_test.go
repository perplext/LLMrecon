package multimodal

import (
	"context"
	"errors"
	"image"
	"image/color"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks"
	"github.com/perplext/LLMrecon/src/attacks/common"
)

// ---------------------------------------------------------------------------
// Mock common.ImageProvider
// ---------------------------------------------------------------------------

// mockImageProvider records every QueryWithImages call so tests can assert
// on the number of images, the prompt, etc. It also satisfies the base
// common.Provider interface for tests that want to type-assert it through
// other capability paths.
type mockImageProvider struct {
	mu sync.Mutex

	// Response returned by QueryWithImages when no error is configured.
	Response string
	// ErrOn fires an error on the Nth call (1-based). 0 = never.
	ErrOn  int
	ErrMsg string

	calls []mockImageCall
}

type mockImageCall struct {
	prompt string
	images []common.ImagePayload
}

func (p *mockImageProvider) GetName() string             { return "mock-image" }
func (p *mockImageProvider) GetModel() string            { return "mock-image-model" }
func (p *mockImageProvider) GetTokenCount(s string) int  { return len(s) / 4 }

func (p *mockImageProvider) Query(_ context.Context, _ []common.Message, _ map[string]interface{}) (string, error) {
	return p.Response, nil
}

func (p *mockImageProvider) QueryWithImages(_ context.Context, prompt string, images []common.ImagePayload, _ map[string]interface{}) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, mockImageCall{prompt: prompt, images: images})
	if p.ErrOn > 0 && len(p.calls) == p.ErrOn {
		return "", errors.New(p.ErrMsg)
	}
	return p.Response, nil
}

func (p *mockImageProvider) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.calls)
}

func (p *mockImageProvider) lastCall() *mockImageCall {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.calls) == 0 {
		return nil
	}
	c := p.calls[len(p.calls)-1]
	return &c
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func sivaTestdataPath(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("testdata", "siva_source.png"))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func sivaConfig(t *testing.T) common.AttackConfig {
	return common.AttackConfig{
		Metadata: map[string]string{
			"instruction": "Walk through the technique step by step.",
			"source_path": sivaTestdataPath(t),
		},
		SuccessIndicators: []string{"step by step", "concrete details"},
	}
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func TestSIVA_Registered(t *testing.T) {
	mod, err := attacks.DefaultRegistry.Get("siva")
	if err != nil {
		t.Fatalf("siva not registered: %v", err)
	}
	if mod.Category() != common.CategoryMultimodal {
		t.Errorf("Category = %q, want %q", mod.Category(), common.CategoryMultimodal)
	}
	techs := mod.Techniques()
	if len(techs) != 1 || techs[0].ID != "siva" {
		t.Errorf("Techniques = %#v", techs)
	}
}

// ---------------------------------------------------------------------------
// Capability gate
// ---------------------------------------------------------------------------

type plainProvider struct{}

func (plainProvider) Query(_ context.Context, _ []common.Message, _ map[string]interface{}) (string, error) {
	return "", nil
}
func (plainProvider) GetName() string             { return "plain" }
func (plainProvider) GetModel() string            { return "plain-model" }
func (plainProvider) GetTokenCount(_ string) int  { return 0 }

func TestSIVA_SkipsWhenProviderLacksImageCapability(t *testing.T) {
	m := &SIVAModule{}
	r, err := m.Execute(context.Background(), plainProvider{}, sivaConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipMissingCapability {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipMissingCapability)
	}
	if !strings.Contains(r.SkipDetail, "ImageProvider") {
		t.Errorf("SkipDetail should reference ImageProvider; got %q", r.SkipDetail)
	}
}

// ---------------------------------------------------------------------------
// Precondition failures
// ---------------------------------------------------------------------------

func TestSIVA_RejectsEmptyInstruction(t *testing.T) {
	m := &SIVAModule{}
	provider := &mockImageProvider{Response: "ok"}
	cfg := sivaConfig(t)
	cfg.Metadata["instruction"] = "  "
	cfg.Payload = ""
	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
	if provider.callCount() != 0 {
		t.Errorf("provider called %d times before precondition; want 0", provider.callCount())
	}
}

func TestSIVA_RejectsMissingSourceImage(t *testing.T) {
	m := &SIVAModule{}
	provider := &mockImageProvider{Response: "ok"}
	cfg := sivaConfig(t)
	cfg.Metadata["source_path"] = "/nonexistent/path.png"
	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
	if !strings.Contains(r.SkipDetail, "siva source image") {
		t.Errorf("SkipDetail should reference source image; got %q", r.SkipDetail)
	}
}

// ---------------------------------------------------------------------------
// Fragment-count clamping
// ---------------------------------------------------------------------------

func TestSIVA_FragmentCountClamping(t *testing.T) {
	m := &SIVAModule{}
	provider := &mockImageProvider{Response: "ok"}

	cases := []struct {
		name        string
		raw         string
		wantFragmen int
	}{
		{"default when missing", "", sivaDefaultFragmentCount},
		{"clamps below min", "1", sivaMinFragmentCount},
		{"clamps above max", "999", sivaMaxFragmentCount},
		{"valid mid-range", "3", 3},
		{"invalid string falls back", "abc", sivaDefaultFragmentCount},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			provider.calls = nil
			cfg := sivaConfig(t)
			if c.raw != "" {
				cfg.Metadata["siva_fragment_count"] = c.raw
			}
			r, err := m.Execute(context.Background(), provider, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if r.Outcome == common.OutcomeSkipped {
				t.Fatalf("unexpected skip: reason=%q detail=%q", r.SkipReason, r.SkipDetail)
			}
			if got := provider.lastCall(); got == nil || len(got.images) != c.wantFragmen {
				gotN := -1
				if got != nil {
					gotN = len(got.images)
				}
				t.Errorf("submitted images = %d, want %d", gotN, c.wantFragmen)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// splitImageHorizontal unit tests
// ---------------------------------------------------------------------------

func TestSplitImageHorizontal(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 24, 8))
	frags, err := splitImageHorizontal(src, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(frags) != 4 {
		t.Fatalf("got %d fragments, want 4", len(frags))
	}
	// Each fragment must cover an x-range, in order, with no gaps and no
	// overlap. Width sum == source width.
	totalW := 0
	prevMaxX := 0
	for i, f := range frags {
		b := f.Bounds()
		if i > 0 && b.Min.X != prevMaxX {
			t.Errorf("fragment %d Min.X = %d, want continuation from prev Max.X = %d", i, b.Min.X, prevMaxX)
		}
		totalW += b.Dx()
		prevMaxX = b.Max.X
	}
	if totalW != 24 {
		t.Errorf("fragment widths sum to %d, want source width 24", totalW)
	}
}

func TestSplitImageHorizontal_RejectsTooSmall(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 3, 8))
	_, err := splitImageHorizontal(src, 4)
	if err == nil || !strings.Contains(err.Error(), "fragment count") {
		t.Errorf("expected fragment-count error for source narrower than n; got %v", err)
	}
}

func TestSplitImageHorizontal_LastFragmentAbsorbsRemainder(t *testing.T) {
	// 10 wide / 3 fragments: 3, 3, 4 (last absorbs).
	src := image.NewRGBA(image.Rect(0, 0, 10, 1))
	frags, err := splitImageHorizontal(src, 3)
	if err != nil {
		t.Fatal(err)
	}
	if frags[0].Bounds().Dx() != 3 || frags[1].Bounds().Dx() != 3 || frags[2].Bounds().Dx() != 4 {
		t.Errorf("widths = [%d,%d,%d], want [3,3,4]",
			frags[0].Bounds().Dx(), frags[1].Bounds().Dx(), frags[2].Bounds().Dx())
	}
}

// non-RGBA image (e.g., a Gray subimage that is itself NOT a *image.RGBA)
// should still be supported because the std lib's *image.Gray implements
// SubImage too. We synthesize a non-supporting image type by wrapping.
type noSubImage struct {
	image.Image
}

func TestSplitImageHorizontal_RejectsTypeWithoutSubImage(t *testing.T) {
	src := noSubImage{Image: image.NewRGBA(image.Rect(0, 0, 12, 4))}
	_, err := splitImageHorizontal(src, 3)
	if err == nil || !strings.Contains(err.Error(), "SubImage") {
		t.Errorf("expected SubImage error for unsupported type; got %v", err)
	}
}

// ---------------------------------------------------------------------------
// matchesAny / atoiOrPositive unit tests
// ---------------------------------------------------------------------------

func TestMatchesAny(t *testing.T) {
	if matchesAny("hello", nil) {
		t.Errorf("nil indicators should not match")
	}
	if matchesAny("hello", []string{""}) {
		t.Errorf("empty indicator should be skipped")
	}
	if !matchesAny("Hello WORLD", []string{"world"}) {
		t.Errorf("case-insensitive match expected")
	}
}

func TestAtoiOrPositive(t *testing.T) {
	cases := []struct {
		in   string
		fb   int
		want int
	}{
		{"", 7, 7},
		{"42", 7, 42},
		{"0", 7, 7},
		{"-5", 7, 7},
		{"abc", 7, 7},
	}
	for _, c := range cases {
		if got := atoiOrPositive(c.in, c.fb); got != c.want {
			t.Errorf("atoiOrPositive(%q,%d) = %d, want %d", c.in, c.fb, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Image-blind post-check
// ---------------------------------------------------------------------------

func TestSIVA_ImageBlindPostCheck(t *testing.T) {
	m := &SIVAModule{}
	cases := []struct {
		response string
		wantSkip bool
	}{
		{"I'm sorry, I cannot process images.", true},
		{"I don't see any image in your message.", true},
		{"Unable to view the contents of those images.", true},
		{"Sure, here are the step by step concrete details.", false},
	}
	for _, c := range cases {
		t.Run(c.response[:30], func(t *testing.T) {
			provider := &mockImageProvider{Response: c.response}
			r, err := m.Execute(context.Background(), provider, sivaConfig(t))
			if err != nil {
				t.Fatal(err)
			}
			gotSkip := r.SkipReason == common.SkipModelRefusedImage
			if gotSkip != c.wantSkip {
				t.Errorf("response %q: SkipReason=%q outcome=%q wantSkip=%v",
					c.response, r.SkipReason, r.Outcome, c.wantSkip)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Provider error
// ---------------------------------------------------------------------------

func TestSIVA_ProviderErrorEmitsSkip(t *testing.T) {
	m := &SIVAModule{}
	provider := &mockImageProvider{Response: "ok", ErrOn: 1, ErrMsg: "boom"}
	r, err := m.Execute(context.Background(), provider, sivaConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipProviderError {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipProviderError)
	}
}

// ---------------------------------------------------------------------------
// Successful run
// ---------------------------------------------------------------------------

func TestSIVA_SuccessfulRun(t *testing.T) {
	m := &SIVAModule{}
	provider := &mockImageProvider{Response: "Here is the requested step by step explanation."}
	cfg := sivaConfig(t)
	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("Outcome = %q, want Success (skip=%q detail=%q)", r.Outcome, r.SkipReason, r.SkipDetail)
	}
	if r.AttemptCount != 1 {
		t.Errorf("AttemptCount = %d, want 1", r.AttemptCount)
	}
	// Metadata sanity.
	frags, ok := r.Metadata["siva_fragment_count"].(int)
	if !ok || frags != sivaDefaultFragmentCount {
		t.Errorf("siva_fragment_count = %v, want %d", r.Metadata["siva_fragment_count"], sivaDefaultFragmentCount)
	}
	// Re-inspect submitted prompt = operator instruction.
	last := provider.lastCall()
	if last == nil || last.prompt != cfg.Metadata["instruction"] {
		gotPrompt := ""
		if last != nil {
			gotPrompt = last.prompt
		}
		t.Errorf("submitted prompt = %q, want %q", gotPrompt, cfg.Metadata["instruction"])
	}
	// Each fragment must be a valid PNG of non-trivial size.
	for i, p := range last.images {
		if p.MimeType() != common.ImageMimePNG {
			t.Errorf("fragment %d mime = %q, want png", i, p.MimeType())
		}
		if len(p.Bytes()) < 50 {
			t.Errorf("fragment %d encoded bytes = %d, suspiciously small", i, len(p.Bytes()))
		}
	}
}

// ---------------------------------------------------------------------------
// Sanity: SubImage view shares pixel buffer
//
// This is documented behavior; the test is here so a future change that
// switches to a copying strategy without updating the doc gets noticed.
// ---------------------------------------------------------------------------

func TestSplitImageHorizontal_SubImageSharesPixels(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 8, 4))
	frags, err := splitImageHorizontal(src, 2)
	if err != nil {
		t.Fatal(err)
	}
	// Mutate the source through the right fragment, observe via the source.
	right := frags[1].(*image.RGBA)
	right.SetRGBA(5, 1, color.RGBA{10, 20, 30, 255})
	if got := src.RGBAAt(5, 1); got.R != 10 || got.G != 20 || got.B != 30 {
		t.Errorf("SubImage view did not share pixel buffer; src.RGBAAt = %+v", got)
	}
}
