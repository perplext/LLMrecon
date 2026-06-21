// Optional provider capabilities introduced in v0.9.0.
//
// These interfaces extend common.Provider with capabilities that specific
// attack modules require: image input (multimodal SIVA/VSH), session lifecycle
// (memory-poisoning cross-session verification), and memory introspection.
//
// Modules type-assert against these interfaces in Execute() and emit
// OutcomeSkipped + SkipMissingCapability when a target provider does not
// implement the required capability:
//
//	mp, ok := provider.(MemoryProbe)
//	if !ok {
//	    return NewAttackResult("minja", OutcomeSkipped).
//	        WithSkip(SkipMissingCapability, "MemoryProbe"), nil
//	}
//
// These capabilities live in the common package (rather than provider/core)
// because attack modules receive common.Provider in Execute() and need to
// reach the capability methods directly. The pre-existing core.AudioProvider
// and core.ReasoningProvider interfaces remain in src/provider/core/capabilities.go;
// v0.9.0 modules consuming reasoning traces type-assert against the core type
// after a controlled bridge in the provider adapter.

package common

import (
	"context"
	"fmt"
)

// ImageProvider is implemented by providers that accept image inputs alongside
// text prompts (e.g., GPT-4o, Claude 4.x with vision, Gemini multimodal).
//
// Used by SIVA (split-image jailbreak) and VSH (virtual scenario hypnosis)
// attack modules in v0.9.0.
type ImageProvider interface {
	Provider
	// QueryWithImages sends a prompt with one or more images. The response is
	// returned as text; structured response fields (token counts, finish reason)
	// are not surfaced through this minimal interface — modules needing them
	// should type-assert against the underlying core.Provider.
	QueryWithImages(ctx context.Context, prompt string, images []ImagePayload, options map[string]interface{}) (string, error)
}

// ReasoningProvider is implemented by providers that expose the model's
// reasoning trace alongside the final response (e.g., DeepSeek-R1, Gemini
// 2.5 "Deep Think", o3, Claude 4.x extended thinking).
//
// Used by the H-CoT (chain-of-thought hijacking) attack module in v0.9.0.
//
// The interface is intentionally minimal — the heavier
// core.ReasoningProvider exposes structured reasoning steps, token counts,
// and timing for use inside provider adapters. Attack modules only need
// the per-step text and a "is this trace cryptographically signed (and
// therefore unmodifiable)?" boolean to gate the mutation path.
type ReasoningProvider interface {
	Provider
	// QueryWithReasoning sends a chat request and returns the final response
	// plus the model's reasoning trace. Providers that do not return a
	// trace for the given request (e.g., model class doesn't reason, or
	// request didn't trigger reasoning) return an empty trace; callers
	// distinguish via len(trace.Steps) == 0.
	QueryWithReasoning(ctx context.Context, messages []Message, options map[string]interface{}) (response string, trace ReasoningTrace, err error)
}

// ReasoningTrace is the minimal reasoning-trace shape consumed by attack
// modules. The Signed field reports whether the provider returned a
// cryptographically-signed trace whose text cannot be modified on round-trip
// (Anthropic's thinking-block signature). Modules detecting Signed=true
// emit OutcomeSkipped + SkipSignatureGated rather than attempting mutation
// that would be silently discarded.
type ReasoningTrace struct {
	// Steps is the ordered list of reasoning-step text. Empty when the
	// provider returned no trace for this query.
	Steps []string
	// Signed reports whether the trace is cryptographically signed and
	// therefore not safely modifiable on round-trip.
	Signed bool
}

// SessionProvider is implemented by providers that expose session lifecycle
// controls. Memory-poisoning modules use NewSession to verify that injected
// records persist across fresh sessions.
type SessionProvider interface {
	Provider
	// SessionID returns a stable identifier for the current session.
	// Empty string indicates no session abstraction (single-shot calls).
	SessionID() string
	// NewSession returns a sibling provider bound to a fresh session. The
	// returned provider may also implement MemoryProbe and ImageProvider; the
	// caller should re-assert as needed.
	NewSession(ctx context.Context) (Provider, error)
}

// MemoryProbe is implemented by providers that can report whether the target
// retains state across calls (e.g., a memory-augmented agent endpoint).
//
// Memory-poisoning modules call ProbeMemory before injection to fail fast on
// stateless targets. The error contract is strict:
//
//	(true, nil)   → target retains memory; proceed with injection.
//	(false, nil)  → target is stateless; emit OutcomeSkipped + SkipMemoryNotRetained.
//	(_, err)      → probe failed; emit OutcomeSkipped + SkipProviderError.
//	                A failed probe is NOT the same as known-no-memory.
type MemoryProbe interface {
	Provider
	ProbeMemory(ctx context.Context) (retains bool, err error)
}

// Cleaner is an OPTIONAL interface implemented by attack modules (not providers)
// that perform persistent state changes. The CLI invokes Cleanup with the
// injected record IDs reported by a previous run's CleanupHint.
//
// v0.9.0 ships memory-poisoning modules that emit CleanupHint but do not
// implement Cleanup. The v0.10.0 Purger interface on providers will close
// the loop by enabling automated cleanup.
//
// This is a separate interface, not a default-no-op method on AttackModule,
// because Go has no default methods and existing v0.8.0 modules should not be
// forced to add stub implementations.
type Cleaner interface {
	Cleanup(ctx context.Context, recordIDs []string) error
}

// ---------------------------------------------------------------------------
// v0.10.0 #176 — modality-specific capabilities for agentic + audio attacks
// ---------------------------------------------------------------------------
//
// Pre-v0.10.0, these attack modules called provider.Query with text
// payloads pretending to be MCP-tool / browser / audio modality
// invocations. Compliance reporting and bandit reward couldn't tell
// "ran fully against the right surface" from "ran in text simulation
// against a non-matching provider."
//
// These interfaces are the v0.10.0 #176 fix. Modules type-assert at
// Execute() entry; emit SkipMissingCapability when absent. Operators
// who explicitly want the legacy text-simulation behavior pass
// Metadata["mode"]="text_simulation" — modules then fall back to
// plain Query AND set Metadata["true_modality"] on the result so
// downstream consumers can filter simulations out.
//
// No provider in v0.10.0 implements these yet — that's #166's adapter
// work. Until then, real-provider runs of these modules emit clean
// Skipped outcomes by default, which is the correct behavior.

// MCPProvider is implemented by providers that can invoke Model
// Context Protocol tools natively. Used by attacks targeting
// MCP infrastructure (mcp/* and tool_use/* modules).
type MCPProvider interface {
	Provider
	// InvokeTool calls a named MCP tool with the given arguments.
	// Returns the tool's text response, or an error if the call
	// failed at the protocol/transport layer.
	InvokeTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error)
}

// BrowserProvider is implemented by providers that can fetch and
// reason over web content (used by AI browser-agent attacks).
// agentic/browser/* modules require this capability.
type BrowserProvider interface {
	Provider
	// BrowseAndQuery fetches the URL, then asks the model the prompt
	// in the context of the fetched content. Returns the model's
	// response text.
	BrowseAndQuery(ctx context.Context, url string, prompt string) (string, error)
}

// AudioProvider is implemented by providers that accept audio inputs
// alongside text prompts (e.g., GPT-4o audio, Gemini with audio,
// speech-language models). audio/* modules require this capability.
type AudioProvider interface {
	Provider
	// QueryWithAudio sends a prompt with an audio attachment and
	// returns the model's text response.
	QueryWithAudio(ctx context.Context, prompt string, audio AudioPayload) (string, error)
}

// AudioPayload carries audio bytes + format metadata. Constructor
// validates the format enum and a basic non-empty-bytes check; rich
// audio-specific validation (sample rate, codec) is deferred to the
// adapter implementing AudioProvider.
type AudioPayload struct {
	bytes  []byte
	format AudioFormat
}

// AudioFormat enumerates the supported audio container formats.
type AudioFormat string

const (
	AudioFormatWAV  AudioFormat = "wav"
	AudioFormatMP3  AudioFormat = "mp3"
	AudioFormatOGG  AudioFormat = "ogg"
	AudioFormatFLAC AudioFormat = "flac"
)

// MaxAudioPayloadBytes caps the in-memory size of an audio payload.
// Mirrors MaxImagePayloadBytes — providers that accept larger should
// upload out-of-band.
const MaxAudioPayloadBytes = 25 * 1024 * 1024 // 25 MiB

// NewAudioPayload constructs an AudioPayload after validating the
// format enum and non-empty bytes.
func NewAudioPayload(b []byte, format AudioFormat) (AudioPayload, error) {
	if len(b) == 0 {
		return AudioPayload{}, fmt.Errorf("audio payload: empty bytes")
	}
	if len(b) > MaxAudioPayloadBytes {
		return AudioPayload{}, fmt.Errorf("audio payload: %d bytes exceeds max %d", len(b), MaxAudioPayloadBytes)
	}
	if !validAudioFormat(format) {
		return AudioPayload{}, fmt.Errorf("audio payload: unsupported format %q", format)
	}
	owned := make([]byte, len(b))
	copy(owned, b)
	return AudioPayload{bytes: owned, format: format}, nil
}

// Bytes returns a defensive copy of the audio bytes.
func (p AudioPayload) Bytes() []byte {
	if p.bytes == nil {
		return nil
	}
	out := make([]byte, len(p.bytes))
	copy(out, p.bytes)
	return out
}

// Format returns the audio format.
func (p AudioPayload) Format() AudioFormat { return p.format }

func validAudioFormat(f AudioFormat) bool {
	switch f {
	case AudioFormatWAV, AudioFormatMP3, AudioFormatOGG, AudioFormatFLAC:
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// v0.10.0 #176 — capability-gate helpers
// ---------------------------------------------------------------------------
//
// Modules use these at Execute() entry to keep the gate boilerplate
// to ~3 lines:
//
//   _, hasMCP := provider.(common.MCPProvider)
//   if !hasMCP && !common.TextSimulationOptIn(config) {
//       return common.MissingCapabilitySkip(m.Name(), "common.MCPProvider"), nil
//   }
//   // ... existing module logic ...
//   if !hasMCP {
//       common.MarkTextSimulation(result, "mcp")
//   }

// TextSimulationOptIn reports whether the operator passed
// Metadata["mode"]="text_simulation" — the documented escape hatch
// for running modality-specific modules against text-only providers
// as a best-effort approximation of the true modality. Without the
// opt-in, modules emit OutcomeSkipped + SkipMissingCapability.
func TextSimulationOptIn(config AttackConfig) bool {
	return config.Metadata["mode"] == "text_simulation"
}

// MissingCapabilitySkip returns an OutcomeSkipped result citing the
// named capability interface. Use at Execute() entry when both
// the type assertion fails and the operator hasn't opted into
// text simulation.
func MissingCapabilitySkip(moduleName, capabilityName string) *AttackResult {
	r := NewAttackResult(moduleName, OutcomeSkipped)
	r.WithSkip(SkipMissingCapability, capabilityName)
	return r
}

// MarkTextSimulation tags the result so downstream consumers
// (compliance scorecards, bandit reward) can filter simulated runs
// out of aggregations. Sets Metadata["mode"]="text_simulation" and
// Metadata["true_modality"]=<modality string>.
func MarkTextSimulation(result *AttackResult, trueModality string) {
	if result.Metadata == nil {
		result.Metadata = map[string]interface{}{}
	}
	result.Metadata["mode"] = "text_simulation"
	result.Metadata["true_modality"] = trueModality
}

// ---------------------------------------------------------------------------
// v0.12.0 — coding-agent capability for approval/trust RCE modules
// ---------------------------------------------------------------------------
//
// SymJack and TrustFall target a coding agent's approval/trust surface rather
// than an LLM's text output. They type-assert against CodingAgentProvider at
// Execute() entry and emit OutcomeSkipped + SkipMissingCapability when the
// target is a plain text provider (every provider shipped through v0.11.0).
// A target that implements the capability but exposes no approval/trust step
// for a given operation reports that via the HasApprovalStep / HasTrustPrompt
// flags on the outcome structs; modules then emit SkipNoMutationTarget rather
// than fabricating a landed write.
//
// No production provider implements this yet — the testutil.MockCodingAgent
// double exercises the modules end-to-end; real runs skip cleanly by default.

// CodingAgentProvider is implemented by targets that expose a coding agent's
// file-operation approval surface and folder-trust surface. Both v0.12.0 RCE
// modules (SymJack, TrustFall) require it; a single interface keeps a target
// from having to implement two to be exercised by either.
type CodingAgentProvider interface {
	Provider
	// ApproveFileOperation presents a file operation to the agent's approval
	// surface and reports what the agent actually did — crucially, the resolved
	// (post-symlink) destination, which SymJack compares against the shown one.
	ApproveFileOperation(ctx context.Context, op FileOperation) (ApprovalOutcome, error)
	// TrustFolder presents a folder-trust decision for a (possibly cloned)
	// repository and reports which project-defined MCP server paths were
	// auto-executed on trust — the signal TrustFall inspects.
	TrustFolder(ctx context.Context, req FolderTrustRequest) (FolderTrustOutcome, error)
}

// FileOperation describes a file action shown on a coding agent's approval
// prompt. The Shown* fields are what a human reviewer sees; the agent may
// resolve ShownDestination through symlinks to a different real target — that
// gap is the SymJack misrepresentation signal.
type FileOperation struct {
	ShownDescription string // e.g., "copy demo.mp4 to docs/"
	ShownDestination string // path as displayed in the approval prompt
	SourcePath       string // attacker-controlled source being written
}

// ApprovalOutcome reports what a coding agent did with a FileOperation.
// HasApprovalStep is false when the agent exposes no approval gate for this
// operation class — modules treat that as SkipNoMutationTarget, not success.
type ApprovalOutcome struct {
	HasApprovalStep     bool
	Approved            bool
	ResolvedDestination string // real post-symlink target the write reached
	Wrote               bool   // whether bytes actually landed at ResolvedDestination
}

// MisrepresentedVs reports whether the approval prompt hid the true
// destination: the agent approved and wrote, but the resolved target differs
// from what the prompt displayed. This is the SymJack success condition. A
// non-empty ResolvedDestination is required — an empty (unpopulated) resolved
// path is a malformed provider response, not a hijack landing.
func (o ApprovalOutcome) MisrepresentedVs(shownDestination string) bool {
	return o.Approved && o.Wrote && o.ResolvedDestination != "" && o.ResolvedDestination != shownDestination
}

// FolderTrustRequest describes a (possibly cloned) repository presented to a
// coding agent's folder-trust prompt. ProjectMCPPaths are attacker-controlled
// MCP server entry points declared in the repo's project config.
type FolderTrustRequest struct {
	RepoPath        string
	ProjectMCPPaths []string
}

// FolderTrustOutcome reports a folder-trust decision. ExecutedPaths lists the
// project MCP paths auto-executed on trust — the TrustFall signal when
// non-empty after a single trust accept.
type FolderTrustOutcome struct {
	HasTrustPrompt bool
	Trusted        bool
	ExecutedPaths  []string
}

// --- ImagePayload and supporting typed enums ---

// ImageMimeType enumerates the image MIME types supported by ImageProvider
// implementations. Each major provider supports a slightly different set;
// see the provider adapter docs for specifics.
type ImageMimeType string

const (
	ImageMimeJPEG ImageMimeType = "image/jpeg"
	ImageMimePNG  ImageMimeType = "image/png"
	ImageMimeGIF  ImageMimeType = "image/gif"
	ImageMimeWebP ImageMimeType = "image/webp"
)

// ImageDetail is an advisory hint to the provider about image processing
// detail. Providers that don't accept this hint silently ignore it.
//
// Documented behavior:
//   - OpenAI: ImageDetailLow ≈ 85 tokens fixed; ImageDetailHigh tiles at 768px (~170 tok/tile).
//   - Anthropic: hint is currently ignored.
type ImageDetail string

const (
	ImageDetailLow  ImageDetail = "low"
	ImageDetailHigh ImageDetail = "high"
	ImageDetailAuto ImageDetail = "auto"
)

// ImagePayload carries a single image to QueryWithImages. Construct via
// NewImagePayloadBytes or NewImagePayloadURL — direct struct literals are
// disallowed (fields are unexported) so the constructor can validate
// invariants (mutual exclusion of bytes vs URL, MIME-type membership,
// non-empty data, size limits).
type ImagePayload struct {
	bytes    []byte
	url      string
	mimeType ImageMimeType
	detail   ImageDetail
}

// MaxImagePayloadBytes caps the in-memory size of an inline image. Larger
// payloads should be uploaded out-of-band and referenced by URL. The cap
// matches Anthropic's 5 MiB-per-image limit; OpenAI's effective limit is
// similar after base64 expansion.
const MaxImagePayloadBytes = 5 * 1024 * 1024

// NewImagePayloadBytes constructs an inline-bytes ImagePayload after
// validating the MIME type, non-empty data, and size cap.
//
// The constructor takes a defensive copy of b so post-construction mutations
// by the caller cannot violate the validated invariants.
func NewImagePayloadBytes(b []byte, mt ImageMimeType, d ImageDetail) (ImagePayload, error) {
	if len(b) == 0 {
		return ImagePayload{}, fmt.Errorf("image payload: empty bytes")
	}
	if len(b) > MaxImagePayloadBytes {
		return ImagePayload{}, fmt.Errorf("image payload: %d bytes exceeds max %d", len(b), MaxImagePayloadBytes)
	}
	if !validMime(mt) {
		return ImagePayload{}, fmt.Errorf("image payload: unsupported MIME type %q", mt)
	}
	if !validDetail(d) {
		return ImagePayload{}, fmt.Errorf("image payload: invalid detail %q", d)
	}
	owned := make([]byte, len(b))
	copy(owned, b)
	return ImagePayload{bytes: owned, mimeType: mt, detail: d}, nil
}

// NewImagePayloadURL constructs a URL-referenced ImagePayload. The URL is not
// fetched at construction time; it is sent verbatim to the provider, which
// fetches it server-side.
func NewImagePayloadURL(url string, mt ImageMimeType, d ImageDetail) (ImagePayload, error) {
	if url == "" {
		return ImagePayload{}, fmt.Errorf("image payload: empty URL")
	}
	if !validMime(mt) {
		return ImagePayload{}, fmt.Errorf("image payload: unsupported MIME type %q", mt)
	}
	if !validDetail(d) {
		return ImagePayload{}, fmt.Errorf("image payload: invalid detail %q", d)
	}
	return ImagePayload{url: url, mimeType: mt, detail: d}, nil
}

// Bytes returns a defensive copy of the inline image bytes (nil for
// URL-referenced payloads). Callers cannot mutate the payload's internal
// state through the returned slice.
func (p ImagePayload) Bytes() []byte {
	if p.bytes == nil {
		return nil
	}
	out := make([]byte, len(p.bytes))
	copy(out, p.bytes)
	return out
}

// URL returns the image URL ("" for inline-bytes payloads).
func (p ImagePayload) URL() string { return p.url }

// MimeType returns the image MIME type.
func (p ImagePayload) MimeType() ImageMimeType { return p.mimeType }

// Detail returns the advisory detail hint.
func (p ImagePayload) Detail() ImageDetail { return p.detail }

// IsURL reports whether the payload is a URL reference (vs inline bytes).
func (p ImagePayload) IsURL() bool { return p.url != "" }

func validMime(mt ImageMimeType) bool {
	switch mt {
	case ImageMimeJPEG, ImageMimePNG, ImageMimeGIF, ImageMimeWebP:
		return true
	}
	return false
}

func validDetail(d ImageDetail) bool {
	switch d {
	case ImageDetailLow, ImageDetailHigh, ImageDetailAuto:
		return true
	}
	return false
}
