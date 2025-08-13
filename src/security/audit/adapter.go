package audit

import (

	"github.com/perplext/LLMrecon/src/audit"
)

// AuditLoggerAdapter adapts our CredentialAuditLogger to the existing audit.AuditLogger interface
type AuditLoggerAdapter struct {
	// credentialLogger is the credential audit logger
	credentialLogger *CredentialAuditLogger
	// auditLogger is the standard audit logger
	auditLogger *audit.AuditLogger
}

// NewAuditLoggerAdapter creates a new audit logger adapter
func NewAuditLoggerAdapter(credentialLogger *CredentialAuditLogger, writer io.Writer, user string) *AuditLoggerAdapter {
	return &AuditLoggerAdapter{
		credentialLogger: credentialLogger,
		auditLogger:      audit.NewAuditLogger(writer, user),
	}
}

// LogCredentialAccess logs a credential access event
func (a *AuditLoggerAdapter) LogCredentialAccess(credentialID, service, operation string) error {
	// Log to credential logger
	if a.credentialLogger != nil {
		if err := a.credentialLogger.LogCredentialAccess(credentialID, service, operation); err != nil {
			return err
		}
	}

	// Log to standard audit logger
	if a.auditLogger != nil {
		a.auditLogger.LogEventWithStatus(
			"credential_access",
			"vault",
			credentialID,
			"success",
			map[string]interface{}{
				"service":   service,
				"operation": operation,
			},
		)
	}

	return nil
}

// LogCredentialError logs a credential error event
func (a *AuditLoggerAdapter) LogCredentialError(credentialID, service, operation string, err error) error {
	// Log to credential logger
	if a.credentialLogger != nil {
		if logErr := a.credentialLogger.LogCredentialError(credentialID, service, operation, err); logErr != nil {
			return logErr
		}
	}

	// Log to standard audit logger
	if a.auditLogger != nil {
		a.auditLogger.LogEventWithStatus(
			"credential_error",
			"vault",
			credentialID,
			"failure",
			map[string]interface{}{
				"service":   service,
				"operation": operation,
				"error":     err.Error(),
			},
		)
	}

	return nil
}

// LogAlert logs an alert event
func (a *AuditLoggerAdapter) LogAlert(message, alertType string, metadata map[string]string) error {
	// Log to credential logger
	if a.credentialLogger != nil {
		if err := a.credentialLogger.LogAlert(message, alertType, metadata); err != nil {
			return err
		}
	}

	// Convert metadata to interface map
	metadataInterface := make(map[string]interface{})
	for k, v := range metadata {
		metadataInterface[k] = v
	}

	// Log to standard audit logger
	if a.auditLogger != nil {
		a.auditLogger.LogEventWithStatus(
			"alert",
			"vault",
			"",
			alertType,
			map[string]interface{}{
				"message":  message,
				"metadata": metadataInterface,
			},
		)
	}

	return nil
}

// LogKeyOperation logs a key operation event
func (a *AuditLoggerAdapter) LogKeyOperation(operation, keyID, details string) error {
	// Log to standard audit logger
	if a.auditLogger != nil {
		a.auditLogger.LogEventWithStatus(
			"key_operation",
			"keystore",
			keyID,
			"success",
			map[string]interface{}{
				"operation": operation,
				"details":   details,
			},
		)
	}
	return nil
}

// GetAuditEvents returns audit events from the credential logger
func (a *AuditLoggerAdapter) GetAuditEvents(limit int, filter map[string]string) ([]CredentialAuditEvent, error) {
	if a.credentialLogger != nil {
		return a.credentialLogger.GetAuditEvents(limit, filter)
	}
	return []CredentialAuditEvent{}, nil
}

// RotateLogFile rotates the credential audit log file
func (a *AuditLoggerAdapter) RotateLogFile() error {
	if a.credentialLogger != nil {
		return a.credentialLogger.RotateLogFile()
	}
	return nil
}

// GetStandardAuditLogger returns the standard audit logger
func (a *AuditLoggerAdapter) GetStandardAuditLogger() *audit.AuditLogger {
	return a.auditLogger
}

// GetCredentialAuditLogger returns the credential audit logger
func (a *AuditLoggerAdapter) GetCredentialAuditLogger() *CredentialAuditLogger {
	return a.credentialLogger
}

// NewNullAuditLoggerAdapter creates a new audit logger adapter with null loggers
// This is useful for testing or when audit logging is not required
func NewNullAuditLoggerAdapter() *AuditLoggerAdapter {
	return &AuditLoggerAdapter{
		credentialLogger: nil,
		auditLogger:      nil,
	}
}
