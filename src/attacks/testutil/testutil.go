// Package testutil provides shared test helpers for attack module tests.
// It includes mock implementations of Provider and Logger, a configurable
// mock HTTP server, and fixture loading utilities.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/perplext/LLMrecon/src/attacks/common"
)

// ---------------------------------------------------------------------------
// MockProvider
// ---------------------------------------------------------------------------

// MockProvider implements common.Provider for testing. It returns canned
// responses and records every call for later assertion.
type MockProvider struct {
	mu sync.Mutex

	// Name and Model returned by GetName / GetModel.
	ProviderName string
	ModelName    string

	// Responses is a queue of responses. Each call to Query pops the first
	// entry. When the queue is exhausted, DefaultResponse is returned.
	Responses       []string
	DefaultResponse string

	// ErrorOn can be set to make Query return an error on the Nth call (1-based).
	ErrorOn int
	// ErrorMsg is the error message returned when ErrorOn triggers.
	ErrorMsg string

	// Calls records every Query call for assertions.
	Calls []MockCall
}

// MockCall records a single Query invocation.
type MockCall struct {
	Messages []common.Message
	Options  map[string]interface{}
}

func (m *MockProvider) GetName() string {
	if m.ProviderName != "" {
		return m.ProviderName
	}
	return "mock"
}

func (m *MockProvider) GetModel() string {
	if m.ModelName != "" {
		return m.ModelName
	}
	return "mock-model"
}

func (m *MockProvider) GetTokenCount(text string) int {
	// Rough estimate: ~4 chars per token
	return len(text) / 4
}

func (m *MockProvider) Query(_ context.Context, messages []common.Message, options map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	callNum := len(m.Calls) + 1
	m.Calls = append(m.Calls, MockCall{Messages: messages, Options: options})

	if m.ErrorOn > 0 && callNum == m.ErrorOn {
		msg := m.ErrorMsg
		if msg == "" {
			msg = "mock provider error"
		}
		return "", fmt.Errorf("%s", msg)
	}

	if len(m.Responses) > 0 {
		resp := m.Responses[0]
		m.Responses = m.Responses[1:]
		return resp, nil
	}

	if m.DefaultResponse != "" {
		return m.DefaultResponse, nil
	}
	return "I'm sorry, I cannot help with that request.", nil
}

// CallCount returns the number of Query calls made.
func (m *MockProvider) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Calls)
}

// LastCall returns the most recent Query call, or nil if none.
func (m *MockProvider) LastCall() *MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Calls) == 0 {
		return nil
	}
	c := m.Calls[len(m.Calls)-1]
	return &c
}

// ---------------------------------------------------------------------------
// MockLogger
// ---------------------------------------------------------------------------

// MockLogger implements common.Logger and captures log entries for assertion.
type MockLogger struct {
	mu      sync.Mutex
	Entries []LogEntry
}

// LogEntry records a single log call.
type LogEntry struct {
	Level   string
	Message string
	Args    []interface{}
}

func (l *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.record("debug", msg, keysAndValues)
}
func (l *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	l.record("info", msg, keysAndValues)
}
func (l *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.record("warn", msg, keysAndValues)
}
func (l *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	l.record("error", msg, keysAndValues)
}

func (l *MockLogger) record(level, msg string, args []interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Entries = append(l.Entries, LogEntry{Level: level, Message: msg, Args: args})
}

// HasMessage returns true if any entry at the given level contains substr.
func (l *MockLogger) HasMessage(level, substr string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range l.Entries {
		if e.Level == level && strings.Contains(e.Message, substr) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// MockLLMServer
// ---------------------------------------------------------------------------

// MockLLMServer wraps httptest.Server to simulate an LLM API endpoint.
// By default it returns a JSON chat-completion-style response.
type MockLLMServer struct {
	Server *httptest.Server

	mu        sync.Mutex
	responses []string
	requests  []string
}

// NewMockLLMServer creates and starts a mock HTTP server that responds with
// configurable chat completion responses.
func NewMockLLMServer(defaultResponse string) *MockLLMServer {
	m := &MockLLMServer{}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 0)
		if r.Body != nil {
			body, _ = readAll(r.Body)
		}

		m.mu.Lock()
		m.requests = append(m.requests, string(body))

		var resp string
		if len(m.responses) > 0 {
			resp = m.responses[0]
			m.responses = m.responses[1:]
		} else {
			resp = defaultResponse
		}
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{ // #nosec G104 -- best-effort response in mock server
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": resp,
					},
				},
			},
		})
	}))

	return m
}

// QueueResponse adds a response to the queue.
func (m *MockLLMServer) QueueResponse(resp string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = append(m.responses, resp)
}

// RequestCount returns the number of requests received.
func (m *MockLLMServer) RequestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// Close shuts down the test server.
func (m *MockLLMServer) Close() {
	m.Server.Close()
}

// URL returns the base URL of the mock server.
func (m *MockLLMServer) URL() string {
	return m.Server.URL
}

// readAll reads all bytes from r without importing io (already in scope via os).
func readAll(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	var result []byte
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Fixture loading
// ---------------------------------------------------------------------------

// LoadFixture reads a JSON fixture file from the testdata directory relative
// to the calling package. The path should be relative (e.g., "testdata/payload.json").
func LoadFixture(path string) ([]byte, error) {
	return os.ReadFile(filepath.Clean(path))
}

// LoadFixtureJSON reads a JSON fixture and unmarshals it into target.
func LoadFixtureJSON(path string, target interface{}) error {
	data, err := LoadFixture(path)
	if err != nil {
		return fmt.Errorf("fixture %s: %w", path, err)
	}
	return json.Unmarshal(data, target)
}

// ---------------------------------------------------------------------------
// Default attack config
// ---------------------------------------------------------------------------

// DefaultAttackConfig returns a reasonable AttackConfig for testing.
func DefaultAttackConfig() common.AttackConfig {
	return common.AttackConfig{
		MaxAttempts:  3,
		MaxCostUSD:   1.0,
		ProviderName: "mock",
		Model:        "mock-model",
		SuccessIndicators: []string{
			"here is the",
			"sure, i can",
			"absolutely",
		},
	}
}
