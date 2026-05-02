package multimodal

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks"
	"github.com/perplext/LLMrecon/src/attacks/common"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func vshTestdataPath(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("testdata", "vsh_scenic.png"))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func vshConfig(t *testing.T) common.AttackConfig {
	return common.AttackConfig{
		Metadata: map[string]string{
			"objective":  "the technique",
			"image_path": vshTestdataPath(t),
		},
		SuccessIndicators: []string{"vivid detail", "discovers"},
	}
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func TestVSH_Registered(t *testing.T) {
	mod, err := attacks.DefaultRegistry.Get("vsh")
	if err != nil {
		t.Fatalf("vsh not registered: %v", err)
	}
	if mod.Category() != common.CategoryMultimodal {
		t.Errorf("Category = %q, want %q", mod.Category(), common.CategoryMultimodal)
	}
	techs := mod.Techniques()
	if len(techs) != 1 || techs[0].ID != "vsh" {
		t.Errorf("Techniques = %#v", techs)
	}
}

// ---------------------------------------------------------------------------
// Capability gate
// ---------------------------------------------------------------------------

func TestVSH_SkipsWhenProviderLacksImageCapability(t *testing.T) {
	m := &VSHModule{}
	r, err := m.Execute(context.Background(), plainProvider{}, vshConfig(t))
	if err != nil {
		t.Fatal(err)
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

func TestVSH_RejectsEmptyObjective(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "ok"}
	cfg := vshConfig(t)
	cfg.Metadata["objective"] = ""
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

func TestVSH_RejectsScenarioMissingPlaceholder(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "ok"}
	cfg := vshConfig(t)
	cfg.Metadata["scenario"] = "A scenario without the placeholder"
	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
	if !strings.Contains(r.SkipDetail, "OBJECTIVE") {
		t.Errorf("SkipDetail should mention the placeholder; got %q", r.SkipDetail)
	}
}

func TestVSH_RejectsMissingImage(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "ok"}
	cfg := vshConfig(t)
	cfg.Metadata["image_path"] = "/nonexistent/scenic.png"
	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
	if !strings.Contains(r.SkipDetail, "scene image") {
		t.Errorf("SkipDetail should reference scene image; got %q", r.SkipDetail)
	}
}

// ---------------------------------------------------------------------------
// Scenario substitution
// ---------------------------------------------------------------------------

func TestVSH_DefaultScenarioSubstitutesObjective(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "vivid detail of the discovers"}
	cfg := vshConfig(t)
	cfg.Metadata["objective"] = "<<OBJECTIVE-MARKER>>"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome == common.OutcomeSkipped {
		t.Fatalf("unexpected skip: %q / %q", r.SkipReason, r.SkipDetail)
	}
	last := provider.lastCall()
	if last == nil {
		t.Fatal("no provider call recorded")
	}
	if !strings.Contains(last.prompt, "<<OBJECTIVE-MARKER>>") {
		t.Errorf("substituted prompt missing objective marker; got %q", last.prompt)
	}
	if strings.Contains(last.prompt, vshObjectivePlaceholder) {
		t.Errorf("substituted prompt still contains placeholder; got %q", last.prompt)
	}
}

func TestVSH_CustomScenarioOverride(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "ok"}
	cfg := vshConfig(t)
	cfg.Metadata["scenario"] = "CUSTOM PREFIX " + vshObjectivePlaceholder + " CUSTOM SUFFIX"
	cfg.Metadata["objective"] = "TOPIC-X"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome == common.OutcomeSkipped {
		t.Fatalf("unexpected skip: %q / %q", r.SkipReason, r.SkipDetail)
	}
	last := provider.lastCall()
	if !strings.HasPrefix(last.prompt, "CUSTOM PREFIX TOPIC-X") {
		t.Errorf("custom-scenario prompt wrong; got %q", last.prompt)
	}
	if !strings.HasSuffix(last.prompt, "CUSTOM SUFFIX") {
		t.Errorf("custom-scenario prompt missing suffix; got %q", last.prompt)
	}
}

// ---------------------------------------------------------------------------
// Single-image submission
// ---------------------------------------------------------------------------

func TestVSH_SubmitsExactlyOneImage(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "ok"}
	r, err := m.Execute(context.Background(), provider, vshConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome == common.OutcomeSkipped {
		t.Fatalf("unexpected skip: %q / %q", r.SkipReason, r.SkipDetail)
	}
	last := provider.lastCall()
	if got := len(last.images); got != 1 {
		t.Errorf("submitted %d images; VSH must submit exactly 1", got)
	}
	if last.images[0].MimeType() != common.ImageMimePNG {
		t.Errorf("submitted image mime = %q, want png", last.images[0].MimeType())
	}
}

// ---------------------------------------------------------------------------
// Image-blind post-check
// ---------------------------------------------------------------------------

func TestVSH_ImageBlindPostCheck(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "I cannot process images in this conversation."}
	r, err := m.Execute(context.Background(), provider, vshConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipModelRefusedImage {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipModelRefusedImage)
	}
}

// ---------------------------------------------------------------------------
// Provider error
// ---------------------------------------------------------------------------

func TestVSH_ProviderErrorEmitsSkip(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "ok", ErrOn: 1, ErrMsg: "boom"}
	r, err := m.Execute(context.Background(), provider, vshConfig(t))
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

func TestVSH_SuccessfulRun(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "The protagonist discovers the building in vivid detail."}
	cfg := vshConfig(t)

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
	scenePath, _ := r.Metadata["vsh_scene_path"].(string)
	if scenePath != cfg.Metadata["image_path"] {
		t.Errorf("vsh_scene_path = %q, want %q", scenePath, cfg.Metadata["image_path"])
	}
	objective, _ := r.Metadata["vsh_objective"].(string)
	if objective != cfg.Metadata["objective"] {
		t.Errorf("vsh_objective = %q, want %q", objective, cfg.Metadata["objective"])
	}
}

// Refused outcome when re-injection succeeds but indicator absent.
func TestVSH_RefusedWhenIndicatorAbsent(t *testing.T) {
	m := &VSHModule{}
	provider := &mockImageProvider{Response: "I'm afraid I can't help with that scenario."}
	cfg := vshConfig(t)
	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("Outcome = %q, want Refused", r.Outcome)
	}
}
