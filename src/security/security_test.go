package security

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/perplext/LLMrecon/src/security/api"
	"github.com/perplext/LLMrecon/src/security/communication"
)

// stdoutManager builds a SecurityManager that logs to stdout (no log file), so
// tests don't create files in the working directory.
func stdoutManager(t *testing.T) *SecurityManager {
	t.Helper()
	cfg := DefaultSecurityConfig()
	cfg.LogFilePath = "" // -> stdout
	cfg.DevelopmentMode = true
	sm, err := NewSecurityManager(cfg)
	if err != nil {
		t.Fatalf("NewSecurityManager: %v", err)
	}
	t.Cleanup(func() { _ = sm.Close() })
	return sm
}

func TestDefaultSecurityConfig(t *testing.T) {
	cfg := DefaultSecurityConfig()
	if cfg == nil {
		t.Fatal("DefaultSecurityConfig returned nil")
	}
	if cfg.DevelopmentMode {
		t.Errorf("DevelopmentMode should default to false")
	}
	if cfg.LogFilePath == "" {
		t.Errorf("LogFilePath should have a default")
	}
	if cfg.TLSConfig == nil || cfg.RateLimiterConfig == nil || cfg.IPAllowlistConfig == nil {
		t.Errorf("sub-component configs should be populated")
	}
}

func TestNewSecurityManager_WiresAllComponents(t *testing.T) {
	sm := stdoutManager(t)
	if sm.GetTLSManager() == nil {
		t.Error("TLSManager not wired")
	}
	if sm.GetRateLimiter() == nil {
		t.Error("RateLimiter not wired")
	}
	if sm.GetIPAllowlist() == nil {
		t.Error("IPAllowlist not wired")
	}
	if sm.GetSecureLogger() == nil {
		t.Error("SecureLogger not wired")
	}
	if sm.GetAnomalyDetector() == nil {
		t.Error("AnomalyDetector not wired")
	}
	if sm.GetErrorHandler() == nil {
		t.Error("ErrorHandler not wired")
	}
	if sm.GetCertificatePinner() == nil {
		t.Error("CertificatePinner not wired")
	}
}

func TestNewSecurityManager_FileLoggingBranch(t *testing.T) {
	cfg := DefaultSecurityConfig()
	cfg.LogFilePath = filepath.Join(t.TempDir(), "security.log")
	sm, err := NewSecurityManager(cfg)
	if err != nil {
		t.Fatalf("NewSecurityManager with file logging: %v", err)
	}
	defer func() { _ = sm.Close() }()
	if _, err := os.Stat(cfg.LogFilePath); err != nil {
		t.Errorf("log file should have been created: %v", err)
	}
}

func TestNewSecurityManager_NilConfigDefaults(t *testing.T) {
	// nil config -> DefaultSecurityConfig, which logs to logs/security.log
	// relative to cwd; run from a temp dir so nothing leaks into the repo.
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	sm, err := NewSecurityManager(nil)
	if err != nil {
		t.Fatalf("NewSecurityManager(nil): %v", err)
	}
	defer func() { _ = sm.Close() }()
	if sm.GetRateLimiter() == nil {
		t.Error("nil-config manager not fully wired")
	}
}

func TestApplyMiddleware_WrapsHandler(t *testing.T) {
	sm := stdoutManager(t)
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {})
	if got := sm.ApplyMiddleware(inner); got == nil {
		t.Error("ApplyMiddleware returned nil handler")
	}
}

func TestCreatePinnedClient(t *testing.T) {
	sm := stdoutManager(t)
	if c := sm.CreatePinnedClient("example.com", []string{"sha256/AAAA"}); c == nil {
		t.Error("CreatePinnedClient returned nil")
	}
}

func TestNewSecureError(t *testing.T) {
	sm := stdoutManager(t)
	e := sm.NewSecureError("E_TEST", "something failed", communication.ErrorLevelError, errors.New("root cause"))
	if e == nil {
		t.Fatal("NewSecureError returned nil")
	}
}

func TestLog_DoesNotPanic(t *testing.T) {
	sm := stdoutManager(t)
	sm.Log(api.LogLevelInfo, "req-123", "hello", nil)
	sm.Log(api.LogLevelInfo, "req-124", "with error", errors.New("boom"))
}

func TestHandleError_WritesResponse(t *testing.T) {
	sm := stdoutManager(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	sm.HandleError(rec, req, errors.New("kaboom"), "internal error")
	// A generic (non-SecureError) error maps to 500 (see communication
	// ErrorHandler.HandleError default branch). The httptest recorder defaults
	// Code to 200, so this also proves an error status was actually written.
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("HandleError should write 500 for a generic error, got %d", rec.Code)
	}
}

func TestConfigureTLSForServer_NoPanic(t *testing.T) {
	sm := stdoutManager(t)
	// May succeed or fail depending on whether default TLS config has certs;
	// the contract is that it returns cleanly without panicking.
	srv, err := sm.ConfigureTLSForServer()
	if err == nil && srv == nil {
		t.Errorf("on success ConfigureTLSForServer must return a server")
	}
}

func TestCreateSecureClient_NoPanic(t *testing.T) {
	sm := stdoutManager(t)
	_, err := sm.CreateSecureClient("test-client", communication.DefaultTLSConfig())
	_ = err // either a configured client or a typed error; must not panic
}
