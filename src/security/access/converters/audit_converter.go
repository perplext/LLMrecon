// Package converters provides conversion functions between different data models
package converters

import (

	"github.com/google/uuid"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// AuditActionToString converts an AuditAction to a string
func AuditActionToString(action models.AuditAction) string {
	return string(action)
}

// StringToAuditAction converts a string to an AuditAction
func StringToAuditAction(action string) models.AuditAction {
	return models.AuditAction(action)
}

// CreateAuditLog creates a new audit log entry with common fields populated
func CreateAuditLog(action models.AuditAction, resourceType, resource, description string) *models.AuditLog {
	return &models.AuditLog{
		ID:           uuid.New().String(),
		Action:       AuditActionToString(action),
		ResourceType: resourceType,
		Resource:     resource,
		Description:  description,
		Timestamp:    time.Now(),
	}
}

// CreateLoginAuditLog creates a login audit log entry
func CreateLoginAuditLog(username, status, ipAddress, userAgent string) *models.AuditLog {
	log := CreateAuditLog(models.AuditActionLogin, "auth", "login", "Login attempt for user: "+username)
	log.Status = status
	log.IPAddress = ipAddress
	log.UserAgent = userAgent
	return log
}

// CreateLogoutAuditLog creates a logout audit log entry
func CreateLogoutAuditLog(userID, username, ipAddress, userAgent string) *models.AuditLog {
	log := CreateAuditLog(models.AuditActionLogout, "auth", "logout", "Logout for user: "+username)
	log.UserID = userID
	log.IPAddress = ipAddress
	log.UserAgent = userAgent
	return log
}

// CreatePasswordUpdateAuditLog creates an audit log entry for password updates
func CreatePasswordUpdateAuditLog(userID, action, ipAddress, userAgent string) *models.AuditLog {
	log := CreateAuditLog(models.AuditActionUpdate, "auth", "user", "Password "+action+" for user")
	log.UserID = userID
	log.ResourceID = userID
	log.IPAddress = ipAddress
	log.UserAgent = userAgent
	return log
}
