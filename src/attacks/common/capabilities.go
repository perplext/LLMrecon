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
	return ImagePayload{bytes: b, mimeType: mt, detail: d}, nil
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

// Bytes returns the inline image bytes (nil for URL-referenced payloads).
func (p ImagePayload) Bytes() []byte { return p.bytes }

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
