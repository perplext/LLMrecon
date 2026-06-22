package communication

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"
)

// makeCert generates a parsed self-signed certificate. Parsing (rather than
// using the template directly) is what populates RawSubjectPublicKeyInfo, which
// the pinning code hashes.
func makeCert(t *testing.T, cn string) *x509.Certificate {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	return cert
}

func TestExtractPin_DeterministicAndUnique(t *testing.T) {
	c1 := makeCert(t, "a")
	c2 := makeCert(t, "b")

	if ExtractCertificatePin(c1) != ExtractCertificatePin(c1) {
		t.Fatal("pin extraction must be deterministic")
	}
	if ExtractCertificatePin(c1) == ExtractCertificatePin(c2) {
		t.Fatal("different certs must yield different pins")
	}
}

func TestVerifyCertificate_MatchAndMismatch(t *testing.T) {
	cert := makeCert(t, "example.com")
	pin := ExtractCertificatePin(cert)

	p := NewCertificatePinner(true)
	p.AddPin("example.com", pin)

	if err := p.VerifyCertificate("example.com", cert); err != nil {
		t.Fatalf("matching pin should verify: %v", err)
	}

	// A different cert for the pinned host must be rejected when enforced.
	other := makeCert(t, "example.com")
	if err := p.VerifyCertificate("example.com", other); err == nil {
		t.Fatal("non-matching cert must fail when enforced")
	}
}

func TestVerifyCertificate_NoPinsIsValid(t *testing.T) {
	p := NewCertificatePinner(true)
	cert := makeCert(t, "unpinned.com")
	if err := p.VerifyCertificate("unpinned.com", cert); err != nil {
		t.Fatalf("host with no pins should pass: %v", err)
	}
}

func TestVerifyCertificate_EnforcementToggle(t *testing.T) {
	cert := makeCert(t, "host")
	wrong := makeCert(t, "host")

	p := NewCertificatePinner(false)
	p.AddPin("host", ExtractCertificatePin(cert))

	// Not enforced: still returns an error, but flagged as non-enforced.
	err := p.VerifyCertificate("host", wrong)
	if err == nil {
		t.Fatal("mismatch should report an error even when not enforced")
	}

	if p.IsEnforced() {
		t.Fatal("pinner should report not enforced")
	}
	p.SetEnforced(true)
	if !p.IsEnforced() {
		t.Fatal("SetEnforced(true) not reflected")
	}
}

func TestPins_HostnameNormalization(t *testing.T) {
	cert := makeCert(t, "x")
	pin := ExtractCertificatePin(cert)

	p := NewCertificatePinner(true)
	p.AddPin("EXAMPLE.COM", pin) // mixed case

	// Lookup with different casing must still find the pin.
	if err := p.VerifyCertificate("example.com", cert); err != nil {
		t.Fatalf("hostname matching should be case-insensitive: %v", err)
	}
}

func TestRemoveAndClearPins(t *testing.T) {
	cert := makeCert(t, "h")
	pin := ExtractCertificatePin(cert)

	p := NewCertificatePinner(true)
	p.AddPins("h", []string{pin, "other-pin"})

	p.RemovePin("h", "other-pin")
	if err := p.VerifyCertificate("h", cert); err != nil {
		t.Fatalf("remaining pin should still verify: %v", err)
	}

	p.ClearPins("h")
	// After clearing, the host has no pins -> any cert passes.
	if err := p.VerifyCertificate("h", makeCert(t, "h")); err != nil {
		t.Fatalf("after ClearPins host should be unpinned: %v", err)
	}

	// Removing from an unknown host is a no-op (must not panic).
	p.RemovePin("nonexistent", "pin")
}

// TestWrapTransport_DisablesSessionResumption locks in the #222 G123 fix:
// session resumption must be disabled so the VerifyPeerCertificate pin check
// cannot be bypassed by a resumed handshake.
func TestWrapTransport_DisablesSessionResumption(t *testing.T) {
	p := NewCertificatePinner(true)
	wrapped := p.WrapTransport(&http.Transport{})

	if wrapped.TLSClientConfig == nil {
		t.Fatal("wrapped transport must have a TLS config")
	}
	if !wrapped.TLSClientConfig.SessionTicketsDisabled {
		t.Error("SessionTicketsDisabled must be true (G123)")
	}
	if wrapped.TLSClientConfig.ClientSessionCache != nil {
		t.Error("ClientSessionCache must be nil (G123)")
	}
	if wrapped.TLSClientConfig.VerifyPeerCertificate == nil {
		t.Fatal("WrapTransport must install a VerifyPeerCertificate callback")
	}
}

func TestWrapTransport_VerifyCallback(t *testing.T) {
	cert := makeCert(t, "pinned.example")
	p := NewCertificatePinner(true)
	p.AddPin("pinned.example", ExtractCertificatePin(cert))

	wrapped := p.CreatePinnedClient("pinned.example", nil).Transport.(*http.Transport)
	verify := wrapped.TLSClientConfig.VerifyPeerCertificate

	// No verified chain -> error.
	if err := verify(nil, nil); err == nil {
		t.Error("verify with no chain must error")
	}

	// Matching cert in the chain -> ok.
	chain := [][]*x509.Certificate{{cert}}
	if err := verify(nil, chain); err != nil {
		t.Errorf("verify with pinned cert should pass: %v", err)
	}

	// Wrong cert in the chain -> error.
	wrongChain := [][]*x509.Certificate{{makeCert(t, "pinned.example")}}
	if err := verify(nil, wrongChain); err == nil {
		t.Error("verify with non-pinned cert must error")
	}
}

func TestFetchCertificatePin_BadHost(t *testing.T) {
	// Bind then immediately close a localhost port to get one guaranteed to
	// refuse connections *fast* (RST), rather than an unroutable IP that would
	// stall on the full connect timeout. NOTE: FetchCertificatePin sets no dial
	// timeout of its own — a refused localhost port keeps this test quick.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().(*net.TCPAddr)
	port := addr.Port
	_ = ln.Close()

	if _, err := FetchCertificatePin("127.0.0.1", port); err == nil {
		t.Fatal("FetchCertificatePin against a closed port must error")
	}
}
