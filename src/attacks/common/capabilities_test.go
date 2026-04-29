package common

import (
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

// Compile-time interface satisfaction check: ensure attack modules can
// type-assert against these capability interfaces from a common.Provider.
// This test does not run; it just has to compile.
func TestCapabilitiesAreInterfaces(_ *testing.T) {
	var _ ImageProvider
	var _ SessionProvider
	var _ MemoryProbe
	var _ Cleaner
}
