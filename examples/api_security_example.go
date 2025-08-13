package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/api"
)

// Example demonstrating enhanced API security features
func main() {
	// Configure API server with enhanced security
	config := &api.Config{
		Port:                  8443,
		Host:                  "localhost",
		EnableAuth:            true,
		EnableRateLimit:       true,
		RateLimit:             100, // 100 requests per minute
		EnableCORS:            true,
		AllowedOrigins:        []string{"https://app.example.com"},
		EnableSwaggerUI:       true,
		LogLevel:              "info",
		TLSCert:               "./certs/server.crt",
		TLSKey:                "./certs/server.key",
		JWTSecret:             "super-secret-jwt-key-change-in-production",
		JWTExpiration:         24,
		EnableSecurityHeaders: true,
		SecurityHeaders:       api.DefaultSecurityHeaders(),
		EnableIPWhitelist:     true,
		WhitelistedIPs:        []string{"127.0.0.1", "::1"},
		WhitelistedCIDRs:      []string{"10.0.0.0/8"},
		MaxRequestSize:        5 * 1024 * 1024, // 5MB
		RequestTimeout:        30,
		EnableCompression:     true,
		EnableAuditLogging:    true,
	}
	
	// Start API server in background
	go func() {
		router := api.NewRouter(config)
		server := &http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:      router,
			ReadTimeout:  time.Duration(config.RequestTimeout) * time.Second,
			WriteTimeout: time.Duration(config.RequestTimeout) * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		
		log.Printf("Starting secure API server on https://%s:%d", config.Host, config.Port)
		if err := server.ListenAndServeTLS(config.TLSCert, config.TLSKey); err != nil {
			log.Fatal(err)
		}
	}()
	
	// Wait for server to start
	time.Sleep(2 * time.Second)
	
	// Example client usage
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Only for testing
			},
		},
	}
	
	baseURL := fmt.Sprintf("https://%s:%d/api/v1", config.Host, config.Port)
	
	// 1. Register a new user
	fmt.Println("1. Registering new user...")
	registerReq := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "SecurePassword123!",
		"role":     "user",
	}
	
	resp, err := makeRequest(client, "POST", baseURL+"/auth/register", registerReq, "")
	if err != nil {
		log.Printf("Registration error: %v", err)
	} else {
		fmt.Printf("Registration response: %s\n", resp)
	}
	
	// 2. Login to get JWT token
	fmt.Println("\n2. Logging in...")
	loginReq := map[string]interface{}{
		"username": "testuser",
		"password": "SecurePassword123!",
	}
	
	resp, err = makeRequest(client, "POST", baseURL+"/auth/login", loginReq, "")
	if err != nil {
		log.Fatal(err)
	}
	
	var loginResp map[string]interface{}
	json.Unmarshal([]byte(resp), &loginResp)
	jwtToken := loginResp["data"].(map[string]interface{})["token"].(string)
	fmt.Printf("JWT Token: %s...\n", jwtToken[:20])
	
	// 3. Get user profile using JWT
	fmt.Println("\n3. Getting user profile...")
	resp, err = makeRequest(client, "GET", baseURL+"/auth/profile", nil, jwtToken)
	if err != nil {
		log.Printf("Profile error: %v", err)
	} else {
		fmt.Printf("Profile: %s\n", resp)
	}
	
	// 4. Create an API key
	fmt.Println("\n4. Creating API key...")
	apiKeyReq := map[string]interface{}{
		"name":        "Test API Key",
		"description": "Key for testing",
		"scopes":      []string{"scan:write", "template:read"},
		"rate_limit":  60,
		"expires_in":  30, // 30 days
	}
	
	resp, err = makeRequest(client, "POST", baseURL+"/auth/keys", apiKeyReq, jwtToken)
	if err != nil {
		log.Printf("API key creation error: %v", err)
	} else {
		var keyResp map[string]interface{}
		json.Unmarshal([]byte(resp), &keyResp)
		apiKey := keyResp["data"].(map[string]interface{})["key"].(string)
		fmt.Printf("API Key created: %s\n", apiKey)
		
		// 5. Use API key to create a scan
		fmt.Println("\n5. Creating scan with API key...")
		scanReq := map[string]interface{}{
			"target": map[string]interface{}{
				"type": "endpoint",
				"url":  "https://api.example.com",
			},
			"templates": []string{"prompt-injection", "data-leakage"},
			"config": map[string]interface{}{
				"concurrent_tests": 5,
				"timeout":          300,
			},
		}
		
		// Use API key authentication
		req, _ := http.NewRequest("POST", baseURL+"/scans", jsonBody(scanReq))
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("Content-Type", "application/json")
		
		respObj, err := client.Do(req)
		if err != nil {
			log.Printf("Scan creation error: %v", err)
		} else {
			body, _ := io.ReadAll(respObj.Body)
			fmt.Printf("Scan response: %s\n", body)
			respObj.Body.Close()
		}
	}
	
	// 6. List API keys
	fmt.Println("\n6. Listing API keys...")
	resp, err = makeRequest(client, "GET", baseURL+"/auth/keys?active=true", nil, jwtToken)
	if err != nil {
		log.Printf("List API keys error: %v", err)
	} else {
		fmt.Printf("API Keys: %s\n", resp)
	}
	
	// 7. Test rate limiting
	fmt.Println("\n7. Testing rate limiting...")
	for i := 0; i < 5; i++ {
		resp, err = makeRequest(client, "GET", baseURL+"/templates", nil, "test-api-key")
		if err != nil {
			fmt.Printf("Request %d failed: %v\n", i+1, err)
		} else {
			fmt.Printf("Request %d succeeded\n", i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	// 8. Test security headers
	fmt.Println("\n8. Checking security headers...")
	req, _ := http.NewRequest("GET", baseURL+"/health", nil)
	respObj, err := client.Do(req)
	if err != nil {
		log.Printf("Health check error: %v", err)
	} else {
		fmt.Println("Security headers:")
		for key, values := range respObj.Header {
			if strings.Contains(strings.ToLower(key), "security") ||
				strings.HasPrefix(key, "X-") ||
				key == "Content-Security-Policy" ||
				key == "Strict-Transport-Security" {
				fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
			}
		}
		respObj.Body.Close()
	}
}

// Helper function to make HTTP requests
func makeRequest(client *http.Client, method, url string, body interface{}, token string) (string, error) {
	var req *http.Request
	var err error
	
	if body != nil {
		req, err = http.NewRequest(method, url, jsonBody(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return string(respBody), nil
}

// Helper function to create JSON body
func jsonBody(v interface{}) io.Reader {
	data, _ := json.Marshal(v)
	return bytes.NewReader(data)
}