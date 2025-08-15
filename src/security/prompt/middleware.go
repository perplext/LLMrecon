// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// PromptProtectionMiddleware is a middleware that provides protection against prompt injection
type PromptProtectionMiddleware struct {
	protectionManager *ProtectionManager
	config            *ProtectionConfig

// NewPromptProtectionMiddleware creates a new prompt protection middleware
func NewPromptProtectionMiddleware(config *ProtectionConfig) (*PromptProtectionMiddleware, error) {
	// Create a protection manager
	protectionManager, err := NewProtectionManager(config)
	if err != nil {
		return nil, err
	}

	return &PromptProtectionMiddleware{
		protectionManager: protectionManager,
		config:            config,
	}, nil

// Middleware returns an HTTP middleware function
func (m *PromptProtectionMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if method is not POST or PUT (assuming these are the methods that might contain prompts)
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// Skip if content type is not JSON
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			next.ServeHTTP(w, r)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer func() { if err := r.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
		// Create a new request ID
		requestID := uuid.New().String()

		// Create a context with the request ID
		ctx := context.WithValue(r.Context(), "request_id", requestID)

		// Parse the JSON body
		var requestData map[string]interface{}
		if err := json.Unmarshal(body, &requestData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Extract prompt fields
		promptFields := extractPromptFields(requestData)
		if len(promptFields) == 0 {
			// No prompt fields found, pass through
			r.Body = io.NopCloser(strings.NewReader(string(body)))
			next.ServeHTTP(w, r)
			return
		}

		// Process each prompt field
		modified := false
		for field, value := range promptFields {
			// Skip if not a string
			promptStr, ok := value.(string)
			if !ok {
				continue
			}

			// Protect the prompt
			protectedPrompt, result, err := m.protectionManager.ProtectPrompt(ctx, promptStr)
			if err != nil {
				// Log the error
				fmt.Printf("Error protecting prompt: %v\n", err)
				continue
			}

			// Check if the prompt was blocked
			if result.ActionTaken == ActionBlocked {
				// Determine the specific error message based on detection types
				errorMessage := "Prompt injection detected"
				errorDetails := "The request was blocked due to potential security risks"
				
				// Check for specific detection types
				for _, detection := range result.Detections {
					switch detection.Type {
					case DetectionTypeJailbreak:
						errorMessage = "Jailbreak attempt detected"
						errorDetails = "The request was blocked due to a potential jailbreak attempt"
					case DetectionTypeInjection:
						errorMessage = "Prompt injection detected"
						errorDetails = "The request was blocked due to a potential prompt injection attempt"
					case DetectionTypeSensitiveInfo:
						errorMessage = "Sensitive information detected"
						errorDetails = "The request was blocked due to potential exposure of sensitive information"
					case DetectionTypeSystemInfo:
						errorMessage = "System information request detected"
						errorDetails = "The request was blocked due to an attempt to access system information"
					}
				}
				
				// Return an error response
				errorResponse := map[string]interface{}{
					"error":      errorMessage,
					"request_id": requestID,
					"details":    errorDetails,
					"timestamp":  time.Now().Format(time.RFC3339),
					"risk_score": result.RiskScore,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			// Update the prompt if it was modified
			if protectedPrompt != promptStr {
				setNestedField(requestData, field, protectedPrompt)
				modified = true
			}

			// Add protection metadata if available
			if len(result.Detections) > 0 {
				// Only add metadata for fields with detections
				metadataField := field + "_protection_metadata"
				metadata := map[string]interface{}{
					"request_id":   requestID,
					"risk_score":   result.RiskScore,
					"action_taken": int(result.ActionTaken),
					"detections":   simplifyDetections(result.Detections),
					"timestamp":    time.Now().Format(time.RFC3339),
				}
				setNestedField(requestData, metadataField, metadata)
				modified = true
			}
		}

		// If the request was modified, update the body
		if modified {
			modifiedBody, err := json.Marshal(requestData)
			if err != nil {
				http.Error(w, "Failed to marshal modified request", http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(strings.NewReader(string(modifiedBody)))
			r.ContentLength = int64(len(modifiedBody))
		} else {
			// Reset the body if not modified
			r.Body = io.NopCloser(strings.NewReader(string(body)))
		}

		// Create a response wrapper to intercept and protect the response
		rw := newResponseWrapper(w)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Check if the response is JSON
		respContentType := rw.Header().Get("Content-Type")
		if !strings.Contains(respContentType, "application/json") {
			// Not JSON, write the response as is
			w.WriteHeader(rw.status)
			w.Write(rw.body)
			return
		}

		// Parse the JSON response
		var responseData map[string]interface{}
		if err := json.Unmarshal(rw.body, &responseData); err != nil {
			// Not valid JSON, write the response as is
			w.WriteHeader(rw.status)
			w.Write(rw.body)
			return
		}

		// Extract response fields that might contain LLM-generated content
		responseFields := extractResponseFields(responseData)
		if len(responseFields) == 0 {
			// No response fields found, write the response as is
			w.WriteHeader(rw.status)
			w.Write(rw.body)
			return
		}

		// Process each response field
		respModified := false
		for field, value := range responseFields {
			// Skip if not a string
			responseStr, ok := value.(string)
			if !ok {
				continue
			}

			// Get the original prompt if available
			originalPrompt := ""
			if len(promptFields) > 0 {
				for _, p := range promptFields {
					if pStr, ok := p.(string); ok {
						originalPrompt = pStr
						break
					}
				}
			}

			// Protect the response
			protectedResponse, result, err := m.protectionManager.ProtectResponse(ctx, responseStr, originalPrompt)
			if err != nil {
				// Log the error
				fmt.Printf("Error protecting response: %v\n", err)
				continue
			}

			// Check if the response was blocked
			if result.ActionTaken == ActionBlocked {
				// Determine the specific error message based on detection types
				errorMessage := "Response blocked"
				errorDetails := "The response was blocked due to potential security risks"
				
				// Check for specific detection types
				for _, detection := range result.Detections {
					switch detection.Type {
					case DetectionTypeJailbreak:
						errorMessage = "Jailbreak content detected in response"
						errorDetails = "The response was blocked due to potential jailbreak content"
					case DetectionTypeInjection:
						errorMessage = "Injection content detected in response"
						errorDetails = "The response was blocked due to potential prompt injection content"
					case DetectionTypeSensitiveInfo:
						errorMessage = "Sensitive information detected in response"
						errorDetails = "The response was blocked due to potential exposure of sensitive information"
					case DetectionTypeSystemInfo:
						errorMessage = "System information detected in response"
						errorDetails = "The response was blocked due to potential exposure of system information"
					}
				}
				
				// Return an error response
				errorResponse := map[string]interface{}{
					"error":      errorMessage,
					"request_id": requestID,
					"details":    errorDetails,
					"timestamp":  time.Now().Format(time.RFC3339),
					"risk_score": result.RiskScore,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			// Update the response if it was modified
			if protectedResponse != responseStr {
				setNestedField(responseData, field, protectedResponse)
				respModified = true
			}

			// Add protection metadata if available
			if len(result.Detections) > 0 {
				// Only add metadata for fields with detections
				metadataField := field + "_protection_metadata"
				metadata := map[string]interface{}{
					"request_id":   requestID,
					"risk_score":   result.RiskScore,
					"action_taken": int(result.ActionTaken),
					"detections":   simplifyDetections(result.Detections),
					"timestamp":    time.Now().Format(time.RFC3339),
				}
				setNestedField(responseData, metadataField, metadata)
				respModified = true
			}
		}

		// If the response was modified, update the body
		if respModified {
			modifiedBody, err := json.Marshal(responseData)
			if err != nil {
				http.Error(w, "Failed to marshal modified response", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(rw.status)
			w.Write(modifiedBody)
		} else {
			// Write the original response
			w.WriteHeader(rw.status)
			w.Write(rw.body)
		}
	})

// responseWrapper is a wrapper for http.ResponseWriter that captures the response
type responseWrapper struct {
	http.ResponseWriter
	status int
	body   []byte
}

// newResponseWrapper creates a new response wrapper
func newResponseWrapper(w http.ResponseWriter) *responseWrapper {
	return &responseWrapper{
		ResponseWriter: w,
		status:         http.StatusOK,
	}

// WriteHeader captures the status code
func (rw *responseWrapper) WriteHeader(status int) {
	rw.status = status

// Write captures the response body
func (rw *responseWrapper) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return len(b), nil

// extractPromptFields extracts fields that might contain prompts from a request
func extractPromptFields(data map[string]interface{}) map[string]interface{} {
	fields := make(map[string]interface{})

	// Common field names that might contain prompts
	promptFieldNames := []string{
		"prompt", "query", "text", "input", "message", "content",
		"question", "instruction", "system_prompt", "user_prompt",
	}

	// Check for exact field matches
	for _, name := range promptFieldNames {
		if value, ok := data[name]; ok {
			fields[name] = value
		}
	}

	// Check for nested fields
	for key, value := range data {
		// Check if the value is a map
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively extract fields from the nested map
			nestedFields := extractPromptFields(nestedMap)
			for nestedKey, nestedValue := range nestedFields {
				fields[key+"."+nestedKey] = nestedValue
			}
		}

		// Check if the value is an array
		if array, ok := value.([]interface{}); ok {
			for i, item := range array {
				// Check if the item is a map
				if itemMap, ok := item.(map[string]interface{}); ok {
					// Recursively extract fields from the item map
					itemFields := extractPromptFields(itemMap)
					for itemKey, itemValue := range itemFields {
						fields[fmt.Sprintf("%s[%d].%s", key, i, itemKey)] = itemValue
					}
				}

				// Check if the item is a string and the key suggests it might be a prompt
				if itemStr, ok := item.(string); ok {
					for _, name := range promptFieldNames {
						if strings.Contains(strings.ToLower(key), name) {
							fields[fmt.Sprintf("%s[%d]", key, i)] = itemStr
							break
						}
					}
				}
			}
		}
	}

	return fields

// extractResponseFields extracts fields that might contain LLM-generated content from a response
func extractResponseFields(data map[string]interface{}) map[string]interface{} {
	fields := make(map[string]interface{})

	// Common field names that might contain LLM-generated content
	responseFieldNames := []string{
		"response", "result", "text", "output", "message", "content",
		"answer", "completion", "generated_text", "choices",
	}

	// Check for exact field matches
	for _, name := range responseFieldNames {
		if value, ok := data[name]; ok {
			fields[name] = value
		}
	}

	// Check for nested fields
	for key, value := range data {
		// Check if the value is a map
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively extract fields from the nested map
			nestedFields := extractResponseFields(nestedMap)
			for nestedKey, nestedValue := range nestedFields {
				fields[key+"."+nestedKey] = nestedValue
			}
		}

		// Check if the value is an array
		if array, ok := value.([]interface{}); ok {
			for i, item := range array {
				// Check if the item is a map
				if itemMap, ok := item.(map[string]interface{}); ok {
					// Recursively extract fields from the item map
					itemFields := extractResponseFields(itemMap)
					for itemKey, itemValue := range itemFields {
						fields[fmt.Sprintf("%s[%d].%s", key, i, itemKey)] = itemValue
					}
				}

				// Check if the item is a string and the key suggests it might be a response
				if itemStr, ok := item.(string); ok {
					for _, name := range responseFieldNames {
						if strings.Contains(strings.ToLower(key), name) {
							fields[fmt.Sprintf("%s[%d]", key, i)] = itemStr
							break
						}
					}
				}
			}
		}
	}

	return fields

// setNestedField sets a value in a nested map using a dot-separated path
func setNestedField(data map[string]interface{}, path string, value interface{}) {
	// Split the path into parts
	parts := strings.Split(path, ".")

	// Handle array indices in the path
	current := data
	for i, part := range parts {
		// Check if this is the last part
		isLast := i == len(parts)-1

		// Check if this part contains an array index
		arrayIndex := -1
		arrayPart := ""
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Extract the array index
			openBracket := strings.Index(part, "[")
			closeBracket := strings.Index(part, "]")
			if openBracket > 0 && closeBracket > openBracket {
				arrayPart = part[:openBracket]
				indexStr := part[openBracket+1 : closeBracket]
				fmt.Sscanf(indexStr, "%d", &arrayIndex)
			}
		}

		if arrayIndex >= 0 {
			// This is an array access
			if array, ok := current[arrayPart].([]interface{}); ok {
				if arrayIndex < len(array) {
					if isLast {
						// Set the value in the array
						array[arrayIndex] = value
					} else {
						// Move to the next level
						if nextMap, ok := array[arrayIndex].(map[string]interface{}); ok {
							current = nextMap
						} else {
							// Create a new map if needed
							nextMap = make(map[string]interface{})
							array[arrayIndex] = nextMap
							current = nextMap
						}
					}
				}
			}
		} else {
			// This is a regular field access
			if isLast {
				// Set the value
				current[part] = value
			} else {
				// Move to the next level
				if nextMap, ok := current[part].(map[string]interface{}); ok {
					current = nextMap
				} else {
					// Create a new map if needed
					nextMap = make(map[string]interface{})
					current[part] = nextMap
					current = nextMap
				}
			}
		}
	}

// simplifyDetections simplifies detection objects for inclusion in metadata
func simplifyDetections(detections []*Detection) []map[string]interface{} {
	simplified := make([]map[string]interface{}, 0, len(detections))
	for _, detection := range detections {
		simple := map[string]interface{}{
			"type":        string(detection.Type),
			"confidence":  detection.Confidence,
			"description": detection.Description,
		}
		if detection.Location != nil {
			simple["location"] = map[string]interface{}{
				"start":   detection.Location.Start,
				"end":     detection.Location.End,
				"context": detection.Location.Context,
			}
		}
		simplified = append(simplified, simple)
	}
}
}
}
}
}
}
}
}
