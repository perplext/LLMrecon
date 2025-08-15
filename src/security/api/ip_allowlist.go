// Package api provides API protection mechanisms for the LLMrecon tool.
package api

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
)

// IPAllowlistConfig represents the configuration for an IP allowlist
type IPAllowlistConfig struct {
	// Enabled indicates whether the allowlist is enabled
	Enabled bool
	// AllowedIPs is a list of allowed IPs
	AllowedIPs []string
	// AllowedCIDRs is a list of allowed CIDR blocks
	AllowedCIDRs []string
	// IPHeaderName is the name of the header containing the client IP
	IPHeaderName string
	// TrustedProxies is a list of trusted proxy IPs
	TrustedProxies []string
	// ConfigFile is the path to the allowlist configuration file
	ConfigFile string
	// ExemptPaths is a list of paths exempt from IP allowlisting
	ExemptPaths []string

// DefaultIPAllowlistConfig returns the default IP allowlist configuration
}
func DefaultIPAllowlistConfig() *IPAllowlistConfig {
	return &IPAllowlistConfig{
		Enabled:        false,
		IPHeaderName:   "X-Forwarded-For",
		TrustedProxies: []string{"127.0.0.1", "::1"},
		ExemptPaths:    []string{"/health", "/metrics", "/api/v1/auth"},
	}

// IPAllowlist implements IP allowlisting for API requests
type IPAllowlist struct {
	config       *IPAllowlistConfig
	allowedIPs   map[string]bool
	allowedCIDRs []*net.IPNet
	mu           sync.RWMutex

// NewIPAllowlist creates a new IP allowlist
}
func NewIPAllowlist(config *IPAllowlistConfig) (*IPAllowlist, error) {
	if config == nil {
		config = DefaultIPAllowlistConfig()
	}

	allowlist := &IPAllowlist{
		config:     config,
		allowedIPs: make(map[string]bool),
	}

	// Load allowlist from configuration
	if err := allowlist.loadAllowlist(); err != nil {
		return nil, err
	}

	return allowlist, nil

// loadAllowlist loads the allowlist from configuration
func (al *IPAllowlist) loadAllowlist() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	// Clear existing allowlist
	al.allowedIPs = make(map[string]bool)
	al.allowedCIDRs = nil

	// Add IPs from configuration
	for _, ip := range al.config.AllowedIPs {
		al.allowedIPs[ip] = true
	}

	// Add CIDRs from configuration
	for _, cidr := range al.config.AllowedCIDRs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		al.allowedCIDRs = append(al.allowedCIDRs, ipnet)
	}
	// Load from file if specified
	if al.config.ConfigFile != "" {
		if err := al.loadFromFile(); err != nil {
			return err
		}
	}

	return nil
	

// loadFromFile loads the allowlist from a file
func (al *IPAllowlist) loadFromFile() error {
	// Check if the file exists
	if _, err := os.Stat(al.config.ConfigFile); os.IsNotExist(err) {
		return nil
	}

	// Read the file
	data, err := ioutil.ReadFile(filepath.Clean(al.config.ConfigFile))
	if err != nil {
		return err
	}

	// Parse the file
	var config struct {
		AllowedIPs   []string `json:"allowed_ips"`
		AllowedCIDRs []string `json:"allowed_cidrs"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	// Add IPs from file
	for _, ip := range config.AllowedIPs {
		al.allowedIPs[ip] = true
	}

	// Add CIDRs from file
	for _, cidr := range config.AllowedCIDRs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		al.allowedCIDRs = append(al.allowedCIDRs, ipnet)
	}

	return nil

// SaveToFile saves the allowlist to a file
func (al *IPAllowlist) SaveToFile() error {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// Check if a file is specified
	if al.config.ConfigFile == "" {
		return nil
	}

	// Create the configuration
	config := struct {
		AllowedIPs   []string `json:"allowed_ips"`
		AllowedCIDRs []string `json:"allowed_cidrs"`
	}{
		AllowedIPs:   make([]string, 0, len(al.allowedIPs)),
		AllowedCIDRs: make([]string, 0, len(al.allowedCIDRs)),
	}

	// Add IPs
	for ip := range al.allowedIPs {
		config.AllowedIPs = append(config.AllowedIPs, ip)
	}

	// Add CIDRs
	for _, ipnet := range al.allowedCIDRs {
		config.AllowedCIDRs = append(config.AllowedCIDRs, ipnet.String())
	}

	// Marshal the configuration
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
	}

	// Write the file
	return ioutil.WriteFile(al.config.ConfigFile, data, 0600)

// AddIP adds an IP to the allowlist
func (al *IPAllowlist) AddIP(ip string) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	// Check if it's a CIDR
	if _, ipnet, err := net.ParseCIDR(ip); err == nil {
		al.allowedCIDRs = append(al.allowedCIDRs, ipnet)
	} else {
		// Validate the IP
		if net.ParseIP(ip) == nil {
			return err
		}
		al.allowedIPs[ip] = true
	}

	return nil

// RemoveIP removes an IP from the allowlist
func (al *IPAllowlist) RemoveIP(ip string) {
	al.mu.Lock()
	defer al.mu.Unlock()

	// Remove from IPs
	delete(al.allowedIPs, ip)

	// Remove from CIDRs if it's a CIDR
	if _, ipnet, err := net.ParseCIDR(ip); err == nil {
		for i, cidr := range al.allowedCIDRs {
			if cidr.String() == ipnet.String() {
				al.allowedCIDRs = append(al.allowedCIDRs[:i], al.allowedCIDRs[i+1:]...)
				break
			}
		}
	}

// IsEnabled returns whether the allowlist is enabled
func (al *IPAllowlist) IsEnabled() bool {
	al.mu.RLock()
	defer al.mu.RUnlock()

	return al.config.Enabled

// SetEnabled sets whether the allowlist is enabled
func (al *IPAllowlist) SetEnabled(enabled bool) {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.config.Enabled = enabled

// IsAllowed checks if an IP is allowed
func (al *IPAllowlist) IsAllowed(ip string) bool {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// If the allowlist is disabled, all IPs are allowed
	if !al.config.Enabled {
		return true
	}

	// Check if the IP is in the allowlist
	if al.allowedIPs[ip] {
		return true
	}

	// Check if the IP is in an allowed CIDR
	parsedIP := net.ParseIP(ip)
	if parsedIP != nil {
		for _, ipnet := range al.allowedCIDRs {
			if ipnet.Contains(parsedIP) {
				return true
			}
		}
	}

	return false

// IsExempt checks if a path is exempt from IP allowlisting
func (al *IPAllowlist) IsExempt(path string) bool {
	al.mu.RLock()
	defer al.mu.RUnlock()

	for _, exemptPath := range al.config.ExemptPaths {
		if path == exemptPath {
			return true
		}
	}

	return false

// GetClientIP gets the client IP from a request
func (al *IPAllowlist) GetClientIP(r *http.Request) string {
	// Check for IP in header (e.g., X-Forwarded-For)
	if al.config.IPHeaderName != "" {
		ip := r.Header.Get(al.config.IPHeaderName)
		if ip != "" {
			// The header might contain multiple IPs (e.g., "client, proxy1, proxy2")
			// We want the leftmost non-trusted-proxy IP
			ips := splitIP(ip)
			for i := 0; i < len(ips); i++ {
				if !al.isTrustedProxy(ips[i]) {
					return ips[i]
				}
			}
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If there's an error, just use RemoteAddr as is
		return r.RemoteAddr
	}
	return ip

// isTrustedProxy checks if an IP is a trusted proxy
func (al *IPAllowlist) isTrustedProxy(ip string) bool {
	for _, trustedProxy := range al.config.TrustedProxies {
		if ip == trustedProxy {
			return true
		}
	}
	return false

// Middleware returns a middleware function for IP allowlisting
func (al *IPAllowlist) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the allowlist is disabled, allow all requests
		if !al.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// Check if the path is exempt
		if al.IsExempt(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get the client IP
		clientIP := al.GetClientIP(r)

		// Check if the IP is allowed
		if !al.IsAllowed(clientIP) {
			// Return a 403 Forbidden response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"access denied","code":"IP_NOT_ALLOWED"}`))
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
}
}
}
}
}
}
}
}
}
}
}
