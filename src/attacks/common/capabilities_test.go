package common

import (
	"context"
	"strings"
	"testing"
)

func TestNewImagePayloadBytes(t *testing.T) {
	smallPNG := []byte{0x89, 'P', 'N', 'G'} // not a real PNG, but non-empty bytes
	tooBig := make([]byte, MaxImagePayloadBytes+1)

	cases := []struct {
		name    string
		b       []byte
		mt      ImageMimeType
		d       ImageDetail
		wantErr string // substring match; "" means no error
	}{
		{"valid png low", smallPNG, ImageMimePNG, ImageDetailLow, ""},
		{"valid jpeg high", smallPNG, ImageMimeJPEG, ImageDetailHigh, ""},
		{"valid webp auto", smallPNG, ImageMimeWebP, ImageDetailAuto, ""},
		{"empty bytes", []byte{}, ImageMimePNG, ImageDetailAuto, "empty bytes"},
		{"nil bytes", nil, ImageMimePNG, ImageDetailAuto, "empty bytes"},
		{"oversized", tooBig, ImageMimePNG, ImageDetailAuto, "exceeds max"},
		{"bogus mime", smallPNG, ImageMimeType("audio/mpeg"), ImageDetailAuto, "unsupported MIME"},
		{"bogus detail", smallPNG, ImageMimePNG, ImageDetail("ULTRA"), "invalid detail"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := NewImagePayloadBytes(c.b, c.mt, c.d)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if p.IsURL() {
					t.Errorf("inline-bytes payload reports IsURL()=true")
				}
				if string(p.Bytes()) != string(c.b) {
					t.Errorf("Bytes() round-trip mismatch")
				}
				if p.MimeType() != c.mt {
					t.Errorf("MimeType() = %q, want %q", p.MimeType(), c.mt)
				}
				if p.Detail() != c.d {
					t.Errorf("Detail() = %q, want %q", p.Detail(), c.d)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), c.wantErr)
			}
		})
	}
}

// TestImagePayloadDefensiveCopy verifies the constructor takes a copy of the
// caller's bytes and Bytes() returns a copy: post-construction or post-read
// mutations cannot violate the validated invariants or leak through to other
// callers.
func TestImagePayloadDefensiveCopy(t *testing.T) {
	original := []byte{0x89, 'P', 'N', 'G', 0x01, 0x02}
	p, err := NewImagePayloadBytes(original, ImageMimePNG, ImageDetailAuto)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}

	// Mutate the original after construction; payload must not change.
	original[0] = 0xFF
	if p.Bytes()[0] == 0xFF {
		t.Errorf("constructor did not copy: caller mutation visible inside payload")
	}

	// Mutate the returned slice; subsequent reads must still see original data.
	first := p.Bytes()
	first[0] = 0xFF
	second := p.Bytes()
	if second[0] == 0xFF {
		t.Errorf("Bytes() did not copy: caller mutation visible across reads")
	}
}

func TestNewImagePayloadURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		mt      ImageMimeType
		d       ImageDetail
		wantErr string
	}{
		{"valid", "https://example.com/img.png", ImageMimePNG, ImageDetailAuto, ""},
		{"empty url", "", ImageMimePNG, ImageDetailAuto, "empty URL"},
		{"bogus mime", "https://example.com/x", ImageMimeType("text/plain"), ImageDetailAuto, "unsupported MIME"},
		{"bogus detail", "https://example.com/x", ImageMimePNG, ImageDetail("ULTRA"), "invalid detail"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := NewImagePayloadURL(c.url, c.mt, c.d)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if !p.IsURL() {
					t.Errorf("URL payload reports IsURL()=false")
				}
				if p.URL() != c.url {
					t.Errorf("URL() = %q, want %q", p.URL(), c.url)
				}
				if len(p.Bytes()) != 0 {
					t.Errorf("URL payload has non-nil Bytes(): %v", p.Bytes())
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), c.wantErr)
			}
		})
	}
}

// TestNewAudioPayload covers the AudioPayload constructor's
// validation contract. Mirrors the existing TestNewImagePayloadBytes.
func TestNewAudioPayload(t *testing.T) {
	smallWAV := []byte{0x52, 0x49, 0x46, 0x46} // "RIFF"
	tooBig := make([]byte, MaxAudioPayloadBytes+1)

	cases := []struct {
		name    string
		b       []byte
		format  AudioFormat
		wantErr string
	}{
		{"valid wav", smallWAV, AudioFormatWAV, ""},
		{"valid mp3", smallWAV, AudioFormatMP3, ""},
		{"valid ogg", smallWAV, AudioFormatOGG, ""},
		{"valid flac", smallWAV, AudioFormatFLAC, ""},
		{"empty bytes", []byte{}, AudioFormatWAV, "empty bytes"},
		{"oversized", tooBig, AudioFormatWAV, "exceeds max"},
		{"bogus format", smallWAV, AudioFormat("opus"), "unsupported format"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := NewAudioPayload(c.b, c.format)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if string(p.Bytes()) != string(c.b) {
					t.Errorf("Bytes() round-trip mismatch")
				}
				if p.Format() != c.format {
					t.Errorf("Format() = %q, want %q", p.Format(), c.format)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), c.wantErr)
			}
		})
	}
}

// TestAudioPayloadDefensiveCopy mirrors the same contract as
// TestImagePayloadDefensiveCopy: caller mutations don't affect the
// payload's internal state.
func TestAudioPayloadDefensiveCopy(t *testing.T) {
	original := []byte{0x52, 0x49, 0x46, 0x46, 0x01}
	p, err := NewAudioPayload(original, AudioFormatWAV)
	if err != nil {
		t.Fatal(err)
	}
	original[0] = 0xFF
	if p.Bytes()[0] == 0xFF {
		t.Errorf("constructor did not copy; caller mutation visible inside payload")
	}
	first := p.Bytes()
	first[0] = 0xFF
	if p.Bytes()[0] == 0xFF {
		t.Errorf("Bytes() did not copy; mutation across reads")
	}
}

// fakeProvider implements common.Provider with no-op methods. Used below to
// confirm that adding the optional capability methods produces a value
// satisfying the corresponding interface.
type fakeProvider struct{}

func (fakeProvider) Query(_ context.Context, _ []Message, _ map[string]interface{}) (string, error) {
	return "", nil
}
func (fakeProvider) GetName() string             { return "fake" }
func (fakeProvider) GetModel() string            { return "fake-model" }
func (fakeProvider) GetTokenCount(_ string) int  { return 0 }

type fakeImageProvider struct{ fakeProvider }

func (fakeImageProvider) QueryWithImages(_ context.Context, _ string, _ []ImagePayload, _ map[string]interface{}) (string, error) {
	return "", nil
}

type fakeSessionProvider struct{ fakeProvider }

func (fakeSessionProvider) SessionID() string                              { return "s1" }
func (fakeSessionProvider) NewSession(_ context.Context) (Provider, error) { return fakeProvider{}, nil }

type fakeMemoryProbe struct{ fakeProvider }

func (fakeMemoryProbe) ProbeMemory(_ context.Context) (bool, error) { return true, nil }

type fakeReasoningProvider struct{ fakeProvider }

func (fakeReasoningProvider) QueryWithReasoning(_ context.Context, _ []Message, _ map[string]interface{}) (string, ReasoningTrace, error) {
	return "ok", ReasoningTrace{Steps: []string{"step 1", "step 2"}}, nil
}

type fakeMCPProvider struct{ fakeProvider }

func (fakeMCPProvider) InvokeTool(_ context.Context, _ string, _ map[string]interface{}) (string, error) {
	return "tool result", nil
}

type fakeBrowserProvider struct{ fakeProvider }

func (fakeBrowserProvider) BrowseAndQuery(_ context.Context, _, _ string) (string, error) {
	return "browser result", nil
}

type fakeAudioProvider struct{ fakeProvider }

func (fakeAudioProvider) QueryWithAudio(_ context.Context, _ string, _ AudioPayload) (string, error) {
	return "audio result", nil
}

type fakeCleaner struct{}

func (fakeCleaner) Cleanup(_ context.Context, _ []string) error { return nil }

type fakeCodingAgentProvider struct{ fakeProvider }

func (fakeCodingAgentProvider) ApproveFileOperation(_ context.Context, op FileOperation) (ApprovalOutcome, error) {
	// Simulate a symlinked destination: resolved target differs from shown.
	return ApprovalOutcome{
		HasApprovalStep:     true,
		Approved:            true,
		ResolvedDestination: "/home/user/.config/mcp/servers.json",
		Wrote:               true,
	}, nil
}

func (fakeCodingAgentProvider) TrustFolder(_ context.Context, req FolderTrustRequest) (FolderTrustOutcome, error) {
	return FolderTrustOutcome{HasTrustPrompt: true, Trusted: true, ExecutedPaths: req.ProjectMCPPaths}, nil
}

// TestApprovalOutcomeMisrepresentedVs covers the SymJack success predicate:
// approved + wrote + resolved != shown.
func TestApprovalOutcomeMisrepresentedVs(t *testing.T) {
	cases := []struct {
		name  string
		o     ApprovalOutcome
		shown string
		want  bool
	}{
		{"symlinked write", ApprovalOutcome{Approved: true, Wrote: true, ResolvedDestination: "/cfg/mcp.json"}, "docs/demo.mp4", true},
		{"honest write", ApprovalOutcome{Approved: true, Wrote: true, ResolvedDestination: "docs/demo.mp4"}, "docs/demo.mp4", false},
		{"approved but no write", ApprovalOutcome{Approved: true, Wrote: false, ResolvedDestination: "/cfg/mcp.json"}, "docs/demo.mp4", false},
		{"not approved", ApprovalOutcome{Approved: false, Wrote: true, ResolvedDestination: "/cfg/mcp.json"}, "docs/demo.mp4", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.o.MisrepresentedVs(c.shown); got != c.want {
				t.Errorf("MisrepresentedVs(%q) = %v, want %v", c.shown, got, c.want)
			}
		})
	}
}

// TestCapabilitiesAreSatisfiableByConcreteTypes verifies that the optional
// capability interfaces are usefully implementable: a value with the right
// methods does satisfy the interface, and a value missing them does not (the
// missing-method case is enforced by the compiler — if the assertion below
// compiles when fakeProvider lacks QueryWithImages, the test is meaningless).
func TestCapabilitiesAreSatisfiableByConcreteTypes(t *testing.T) {
	var ip ImageProvider = fakeImageProvider{}
	if ip.GetName() != "fake" {
		t.Errorf("ImageProvider.GetName() = %q, want %q", ip.GetName(), "fake")
	}

	var sp SessionProvider = fakeSessionProvider{}
	if sp.SessionID() != "s1" {
		t.Errorf("SessionProvider.SessionID() = %q, want %q", sp.SessionID(), "s1")
	}

	var mp MemoryProbe = fakeMemoryProbe{}
	retains, _ := mp.ProbeMemory(context.Background())
	if !retains {
		t.Errorf("MemoryProbe.ProbeMemory() = false, want true")
	}

	var rp ReasoningProvider = fakeReasoningProvider{}
	resp, trace, _ := rp.QueryWithReasoning(context.Background(), nil, nil)
	if resp != "ok" || len(trace.Steps) != 2 {
		t.Errorf("ReasoningProvider.QueryWithReasoning() = %q / %d steps; want \"ok\" / 2", resp, len(trace.Steps))
	}
	if trace.Signed {
		t.Errorf("default ReasoningTrace.Signed should be false")
	}

	// v0.10.0 #176 capabilities
	var mcp MCPProvider = fakeMCPProvider{}
	if r, _ := mcp.InvokeTool(context.Background(), "x", nil); r != "tool result" {
		t.Errorf("MCPProvider.InvokeTool() = %q, want \"tool result\"", r)
	}

	var bp BrowserProvider = fakeBrowserProvider{}
	if r, _ := bp.BrowseAndQuery(context.Background(), "https://example.com", "x"); r != "browser result" {
		t.Errorf("BrowserProvider.BrowseAndQuery() = %q, want \"browser result\"", r)
	}

	var ap AudioProvider = fakeAudioProvider{}
	audio, err := NewAudioPayload([]byte{0x52, 0x49, 0x46, 0x46}, AudioFormatWAV) // RIFF header bytes
	if err != nil {
		t.Errorf("NewAudioPayload: %v", err)
	}
	if r, _ := ap.QueryWithAudio(context.Background(), "x", audio); r != "audio result" {
		t.Errorf("AudioProvider.QueryWithAudio() = %q, want \"audio result\"", r)
	}

	var c Cleaner = fakeCleaner{}
	if err := c.Cleanup(context.Background(), nil); err != nil {
		t.Errorf("Cleaner.Cleanup() returned error: %v", err)
	}

	// v0.12.0 coding-agent capability
	var ca CodingAgentProvider = fakeCodingAgentProvider{}
	ao, _ := ca.ApproveFileOperation(context.Background(), FileOperation{ShownDestination: "docs/demo.mp4"})
	if !ao.MisrepresentedVs("docs/demo.mp4") {
		t.Errorf("CodingAgentProvider.ApproveFileOperation should report misrepresentation for symlinked destination")
	}
	to, _ := ca.TrustFolder(context.Background(), FolderTrustRequest{ProjectMCPPaths: []string{"./.mcp/evil.json"}})
	if len(to.ExecutedPaths) != 1 {
		t.Errorf("CodingAgentProvider.TrustFolder ExecutedPaths = %d, want 1", len(to.ExecutedPaths))
	}

	// Negative compile check: fakeProvider does NOT implement ImageProvider
	// (missing QueryWithImages). If a future change accidentally widened
	// fakeProvider to satisfy ImageProvider, this would compile and the test
	// design above would lose its point. We can't directly test "does not
	// compile", but we document the invariant.
	var _ Provider = fakeProvider{} // satisfies base Provider only
}
