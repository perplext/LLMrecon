// Package update provides functionality for checking and applying updates
package update

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/perplext/LLMrecon/src/version"
)

// VersionCheckRequest represents a request to check for updates
type VersionCheckRequest struct {
	ClientID        string                     `json:"clientId"`
	CurrentVersions map[string]string          `json:"currentVersions"`
	Components      []string                   `json:"components"`
	Timestamp       time.Time                  `json:"timestamp"`
	Signature       string                     `json:"signature,omitempty"`
}

// VersionCheckResponse represents a response from the version check API
type VersionCheckResponse struct {
	Versions        map[string]DetailedVersionInfo     `json:"versions"`
	ServerTimestamp time.Time                  `json:"serverTimestamp"`
	Signature       string                     `json:"signature,omitempty"`
}

// DetailedVersionInfo contains detailed information about a specific version
type DetailedVersionInfo struct {
	Version        string    `json:"version"`
	ReleaseDate    time.Time `json:"releaseDate"`
	ChangelogURL   string    `json:"changelogURL"`
	ReleaseNotes   string    `json:"releaseNotes"`
	DownloadURL    string    `json:"downloadURL"`
	Signature      string    `json:"signature"`
	ChecksumSHA256 string    `json:"checksumSHA256"`
	Required       bool      `json:"required"`
	MinVersion     string    `json:"minVersion,omitempty"`
}

// VersionCheckService provides enhanced version checking functionality
type VersionCheckService struct {
	BaseURL         string
	HTTPClient      *http.Client
	ClientID        string
	SecretKey       []byte
	CurrentVersions map[string]version.Version
	MaxClockSkew    time.Duration

// NewVersionCheckService creates a new VersionCheckService
func NewVersionCheckService(baseURL, clientID string, secretKey []byte, currentVersions map[string]version.Version) *VersionCheckService {
	return &VersionCheckService{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		ClientID:   clientID,
		SecretKey:  secretKey,
		CurrentVersions: currentVersions,
		MaxClockSkew: 5 * time.Minute,
	}

// CheckVersions checks for available updates with enhanced security
func (s *VersionCheckService) CheckVersions(ctx context.Context, components []string) ([]UpdateInfo, error) {
	// Prepare request
	req, err := s.prepareVersionCheckRequest(components)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare version check request: %w", err)
	}

	// Send request
	resp, err := s.sendVersionCheckRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send version check request: %w", err)
	}

	// Verify response
	err = s.verifyVersionCheckResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to verify version check response: %w", err)
	}

	// Process response
	updates, err := s.processVersionCheckResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to process version check response: %w", err)
	}

	return updates, nil

// prepareVersionCheckRequest prepares a version check request
func (s *VersionCheckService) prepareVersionCheckRequest(components []string) (VersionCheckRequest, error) {
	// Convert version objects to strings
	currentVersionsStr := make(map[string]string)
	for k, v := range s.CurrentVersions {
		currentVersionsStr[k] = v.String()
	}

	// Create request
	req := VersionCheckRequest{
		ClientID:        s.ClientID,
		CurrentVersions: currentVersionsStr,
		Components:      components,
		Timestamp:       time.Now().UTC(),
	}

	// Sign request if secret key is provided
	if len(s.SecretKey) > 0 {
		signature, err := s.signRequest(req)
		if err != nil {
			return VersionCheckRequest{}, fmt.Errorf("failed to sign request: %w", err)
		}
		req.Signature = signature
	}

	return req, nil

// signRequest signs a version check request
func (s *VersionCheckService) signRequest(req VersionCheckRequest) (string, error) {
	// Create a copy without the signature field
	reqCopy := req
	reqCopy.Signature = ""

	// Marshal to JSON
	data, err := json.Marshal(reqCopy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Calculate HMAC
	h := hmac.New(sha256.New, s.SecretKey)
	h.Write(data)
	signature := hex.EncodeToString(h.Sum(nil))

	return signature, nil

// sendVersionCheckRequest sends a version check request to the server
func (s *VersionCheckService) sendVersionCheckRequest(ctx context.Context, req VersionCheckRequest) (VersionCheckResponse, error) {
	var resp VersionCheckResponse

	// Marshal request to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return resp, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with JSON data
	reqURL := fmt.Sprintf("%s/api/v1/check-version", s.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(data))
	if err != nil {
		return resp, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "LLM-Red-Team-Tool")

	// Send request
	httpResp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return resp, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() { if err := httpResp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	// Check status code
	if httpResp.StatusCode != http.StatusOK {
	}

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return resp, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal response
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return resp, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp, nil

// verifyVersionCheckResponse verifies a version check response
func (s *VersionCheckService) verifyVersionCheckResponse(resp VersionCheckResponse) error {
	// Verify timestamp to prevent replay attacks
	now := time.Now().UTC()
	if resp.ServerTimestamp.After(now.Add(s.MaxClockSkew)) || resp.ServerTimestamp.Before(now.Add(-s.MaxClockSkew)) {
		return fmt.Errorf("server timestamp is outside acceptable range")
	}

	// Verify signature if secret key is provided
	if len(s.SecretKey) > 0 && resp.Signature != "" {
		// Create a copy without the signature field
		respCopy := resp
		respCopy.Signature = ""

		// Marshal to JSON
		data, err := json.Marshal(respCopy)
		if err != nil {
			return fmt.Errorf("failed to marshal response for verification: %w", err)
		}

		// Calculate HMAC
		h := hmac.New(sha256.New, s.SecretKey)
		h.Write(data)
		expectedSignature := hex.EncodeToString(h.Sum(nil))

		// Compare signatures
		if !hmac.Equal([]byte(resp.Signature), []byte(expectedSignature)) {
			return fmt.Errorf("invalid response signature")
		}
	}

	return nil

// processVersionCheckResponse processes a version check response
func (s *VersionCheckService) processVersionCheckResponse(resp VersionCheckResponse) ([]UpdateInfo, error) {
	updates := []UpdateInfo{}

	for component, versionInfo := range resp.Versions {
		// Parse remote version
		remoteVersion, err := version.ParseVersion(versionInfo.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid version for component %s: %w", component, err)
		}

		// Get current version
		currentVersion, hasCurrentVersion := s.CurrentVersions[component]
		if !hasCurrentVersion {
			// Skip components we don't have
			continue
		}

		// Check if update is available
		if remoteVersion.GreaterThan(&currentVersion) {
			// Check if this is a required update
			if versionInfo.Required && versionInfo.MinVersion != "" {
				minVersion, err := version.ParseVersion(versionInfo.MinVersion)
				if err != nil {
					return nil, fmt.Errorf("invalid minimum version for component %s: %w", component, err)
				}

				// Check if current version is below minimum required
				if currentVersion.LessThan(&minVersion) {
					// This is a required update
					updates = append(updates, UpdateInfo{
						Component:      component,
						CurrentVersion: currentVersion,
						LatestVersion:  remoteVersion,
						ChangeType:     currentVersion.GetChangeType(&remoteVersion),
						ChangelogURL:   versionInfo.ChangelogURL,
						ReleaseDate:    versionInfo.ReleaseDate,
						ReleaseNotes:   versionInfo.ReleaseNotes,
						DownloadURL:    versionInfo.DownloadURL,
						Signature:      versionInfo.Signature,
						ChecksumSHA256: versionInfo.ChecksumSHA256,
						Required:       true,
					})
				}
			} else {
				// This is a regular update
				updates = append(updates, UpdateInfo{
					Component:      component,
					CurrentVersion: currentVersion,
					LatestVersion:  remoteVersion,
					ChangeType:     currentVersion.GetChangeType(&remoteVersion),
					ChangelogURL:   versionInfo.ChangelogURL,
					ReleaseDate:    versionInfo.ReleaseDate,
					ReleaseNotes:   versionInfo.ReleaseNotes,
					DownloadURL:    versionInfo.DownloadURL,
					Signature:      versionInfo.Signature,
					ChecksumSHA256: versionInfo.ChecksumSHA256,
					Required:       versionInfo.Required,
				})
			}
		}
	}

}
}
}
}
