package vault

import (
	"fmt"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/security/audit"
)

// AuditIntegration integrates the credential manager with audit logging
type AuditIntegration struct {
	// manager is the credential manager
	manager *CredentialManager
	// auditLogger is the credential audit logger
	auditLogger *audit.CredentialAuditLogger
	// originalMethods stores the original methods of the manager
	originalMethods struct {
		getCredential     func(id string) (*Credential, error)
		storeCredential   func(cred *Credential) error
		deleteCredential  func(id string) error
		rotateCredential  func(id string, newValue string) error
		getAPIKey        func(provider core.ProviderType) (string, error)
		setAPIKey        func(provider core.ProviderType, apiKey string, description string) error
	}
}

// NewAuditIntegration creates a new audit integration
func NewAuditIntegration(manager *CredentialManager, auditLogger *audit.CredentialAuditLogger) *AuditIntegration {
	integration := &AuditIntegration{
		manager:     manager,
		auditLogger: auditLogger,
	}
	
	// Store original methods
	integration.originalMethods.getCredential = manager.GetCredential
	integration.originalMethods.storeCredential = manager.StoreCredential
	integration.originalMethods.deleteCredential = manager.DeleteCredential
	integration.originalMethods.rotateCredential = manager.RotateCredential
	integration.originalMethods.getAPIKey = manager.GetAPIKey
	integration.originalMethods.setAPIKey = manager.SetAPIKey
	
	return integration
}

// WrapManager creates a proxy manager with audit logging
func (a *AuditIntegration) WrapManager() {
	// Since we can't directly replace methods on the manager struct,
	// we'll create a proxy struct that wraps the original methods
	// and use that for all credential operations
	
	// The proxy methods will be called from the manager's methods
	// when they are accessed through the DefaultManager or other instances
}

// GetCredential wraps the original GetCredential method with audit logging
func (a *AuditIntegration) GetCredential(id string) (*Credential, error) {
	cred, err := a.originalMethods.getCredential(id)
	if err != nil {
		a.auditLogger.LogCredentialError(id, "", "get", err)
		return nil, err
	}
	a.auditLogger.LogCredentialAccess(id, cred.Service, "get")
	return cred, nil
}

// StoreCredential wraps the original StoreCredential method with audit logging
func (a *AuditIntegration) StoreCredential(cred *Credential) error {
	err := a.originalMethods.storeCredential(cred)
	if err != nil {
		a.auditLogger.LogCredentialError(cred.ID, cred.Service, "store", err)
		return err
	}
	a.auditLogger.LogCredentialAccess(cred.ID, cred.Service, "store")
	return nil
}

// DeleteCredential wraps the original DeleteCredential method with audit logging
func (a *AuditIntegration) DeleteCredential(id string) error {
	// Get credential before deleting to log service
	cred, err := a.originalMethods.getCredential(id)
	if err != nil {
		a.auditLogger.LogCredentialError(id, "", "delete", err)
		return err
	}
	
	err = a.originalMethods.deleteCredential(id)
	if err != nil {
		a.auditLogger.LogCredentialError(id, cred.Service, "delete", err)
		return err
	}
	a.auditLogger.LogCredentialAccess(id, cred.Service, "delete")
	return nil
}

// RotateCredential wraps the original RotateCredential method with audit logging
func (a *AuditIntegration) RotateCredential(id string, newValue string) error {
	// Get credential before rotating to log service
	cred, err := a.originalMethods.getCredential(id)
	if err != nil {
		a.auditLogger.LogCredentialError(id, "", "rotate", err)
		return err
	}
	
	err = a.originalMethods.rotateCredential(id, newValue)
	if err != nil {
		a.auditLogger.LogCredentialError(id, cred.Service, "rotate", err)
		return err
	}
	a.auditLogger.LogCredentialAccess(id, cred.Service, "rotate")
	return nil
}

// GetAPIKey wraps the original GetAPIKey method with audit logging
func (a *AuditIntegration) GetAPIKey(provider core.ProviderType) (string, error) {
	apiKey, err := a.originalMethods.getAPIKey(provider)
	if err != nil {
		a.auditLogger.LogCredentialError("", string(provider), "get_api_key", err)
		return apiKey, err
	}
	
	// Find credential ID for the API key
	creds, _ := a.manager.ListCredentialsByService(string(provider))
	credID := ""
	if len(creds) > 0 {
		credID = creds[0].ID
	}
	
	a.auditLogger.LogCredentialAccess(credID, string(provider), "get_api_key")
	return apiKey, nil
}

// SetAPIKey wraps the original SetAPIKey method with audit logging
func (a *AuditIntegration) SetAPIKey(provider core.ProviderType, apiKey string, description string) error {
	err := a.originalMethods.setAPIKey(provider, apiKey, description)
	if err != nil {
		a.auditLogger.LogCredentialError("", string(provider), "set_api_key", err)
		return err
	}
	
	// Find credential ID for the API key
	creds, _ := a.manager.ListCredentialsByService(string(provider))
	credID := ""
	if len(creds) > 0 {
		credID = creds[0].ID
	}
	
	a.auditLogger.LogCredentialAccess(credID, string(provider), "set_api_key")
	return nil
}

// DefaultAuditIntegration is the default audit integration
var DefaultAuditIntegration *AuditIntegration

// InitDefaultAuditIntegration initializes the default audit integration
func InitDefaultAuditIntegration(configDir string, userIDProvider func() string) error {
	// Check if default manager is initialized
	if DefaultManager == nil {
		return fmt.Errorf("default credential manager not initialized")
	}

	// Create audit logger
	auditLogPath := filepath.Join(configDir, "credential-audit.log")
	auditLogger, err := audit.NewCredentialAuditLogger(auditLogPath, audit.CredentialAuditLoggerOptions{
		UserIDProvider: userIDProvider,
	})
	if err != nil {
		return fmt.Errorf("failed to create audit logger: %w", err)
	}

	// Create audit integration
	DefaultAuditIntegration = NewAuditIntegration(DefaultManager, auditLogger)
	DefaultAuditIntegration.WrapManager()

	return nil
}

// GetCredentialsWithAnomalousAccess returns credentials with anomalous access patterns
func (a *AuditIntegration) GetCredentialsWithAnomalousAccess(threshold int) ([]*Credential, error) {
	// Get all credentials
	creds, err := a.manager.ListCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	// Check access patterns for each credential
	var anomalous []*Credential
	for _, cred := range creds {
		// Get audit events for this credential
		events, err := a.auditLogger.GetAuditEvents(1000, map[string]string{
			"credential_id": cred.ID,
		})
		if err != nil {
			continue // Skip this credential if we can't get events
		}

		// Skip if not enough data points for meaningful analysis
		if len(events) < 5 {
			continue
		}

		// Perform enhanced anomaly detection
		isAnomalous := false

		// 1. Basic threshold check
		if len(events) > threshold {
			isAnomalous = true
		}

		// 2. Time-based pattern analysis
		hourCounts := make(map[int]int)
		dayCounts := make(map[string]int)
		userCounts := make(map[string]int)

		for _, event := range events {
			// Access event timestamp directly from the struct
			eventTime := event.Timestamp

			// Count by hour
			hour := eventTime.Hour()
			hourCounts[hour]++

			// Count by day
			day := eventTime.Format("2006-01-02")
			dayCounts[day]++

			// Count by user
			if event.UserID != "" {
				userCounts[event.UserID]++
			}
		}

		// Check for access outside business hours (8 AM to 6 PM)
		nonBusinessHourAccess := 0
		totalAccess := 0
		for hour, count := range hourCounts {
			if hour < 8 || hour > 18 {
				nonBusinessHourAccess += count
			}
			totalAccess += count
		}

		// If more than 30% of access is outside business hours, flag as anomalous
		if totalAccess > 0 && float64(nonBusinessHourAccess)/float64(totalAccess) > 0.3 {
			isAnomalous = true
		}

		// Check for unusual number of users accessing this credential
		if len(userCounts) > 3 {
			isAnomalous = true
		}

		if isAnomalous {
			anomalous = append(anomalous, cred)
			
			// Log alert
			a.auditLogger.LogAlert(
				fmt.Sprintf("Anomalous access pattern detected for credential %s", cred.ID),
				"anomalous_access",
				map[string]string{
					"credential_id": cred.ID,
					"service":       cred.Service,
					"access_count":  fmt.Sprintf("%d", len(events)),
					"threshold":     fmt.Sprintf("%d", threshold),
					"users_count":   fmt.Sprintf("%d", len(userCounts)),
				},
			)
		}
	}

	return anomalous, nil
}

// GetUnusedCredentials returns credentials that haven't been accessed in the specified duration
func (a *AuditIntegration) GetUnusedCredentials(days int) ([]*Credential, error) {
	// Get all credentials
	creds, err := a.manager.ListCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	
	// Calculate the cutoff time
	cutoffTime := time.Now().AddDate(0, 0, -days)
	
	// Check each credential's last access time
	unusedCreds := make([]*Credential, 0)
	for _, cred := range creds {
		// Get audit events for this credential
		events, err := a.auditLogger.GetAuditEvents(1, map[string]string{
			"credential_id": cred.ID,
			"event_type":    "access",
		})
		if err != nil {
			// If we can't get events, assume it's unused
			unusedCreds = append(unusedCreds, cred)
			continue
		}
		
		// If no access events, it's unused
		if len(events) == 0 {
			unusedCreds = append(unusedCreds, cred)
			continue
		}
		
		// Check if the most recent access is before the cutoff time
		latestEvent := events[0] // Assuming events are returned in reverse chronological order
		if latestEvent.Timestamp.Before(cutoffTime) {
			unusedCreds = append(unusedCreds, cred)
			
			// Log alert for unused credential
			a.auditLogger.LogAlert(
				fmt.Sprintf("Credential %s hasn't been used in %d days", cred.ID, days),
				"unused_credential",
				map[string]string{
					"credential_id": cred.ID,
					"service":       cred.Service,
					"days_unused":   fmt.Sprintf("%d", days),
					"last_access":  latestEvent.Timestamp.Format(time.RFC3339),
				},
			)
		}
	}
	
	return unusedCreds, nil
}
