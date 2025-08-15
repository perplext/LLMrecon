// Package common provides common constants and utilities for the security access control system
package common

import "time"


// AuditAction defines the type of action performed in an audit log
type AuditAction string

// AuditSeverity defines the severity level of an audit log
type AuditSeverity string

// Audit action constants
const (
	// User-related actions
	AuditActionUserCreate                  AuditAction = "user.create"
	AuditActionUserUpdate                  AuditAction = "user.update"
	AuditActionUserDelete                  AuditAction = "user.delete"
	AuditActionUserLogin                   AuditAction = "user.login"
	AuditActionUserLogout                  AuditAction = "user.logout"
	AuditActionUserLoginFailed             AuditAction = "user.login.failed"
	AuditActionUserPasswordChange          AuditAction = "user.password.change"
	AuditActionUserPasswordReset           AuditAction = "user.password.reset"
	AuditActionUserRoleAdd                 AuditAction = "user.role.add"
	AuditActionUserRoleRemove              AuditAction = "user.role.remove"
	AuditActionUserPermissionAdd           AuditAction = "user.permission.add"
	AuditActionUserPermissionRemove        AuditAction = "user.permission.remove"
	AuditActionUserLockout                 AuditAction = "user.lockout"
	AuditActionUserUnlock                  AuditAction = "user.unlock"
	AuditActionUserSuspend                 AuditAction = "user.suspend"
	AuditActionUserActivate                AuditAction = "user.activate"
	AuditActionUserProfileView             AuditAction = "user.profile.view"
	AuditActionUserProfileUpdate           AuditAction = "user.profile.update"

	// Role-related actions
	AuditActionRoleCreate                  AuditAction = "role.create"
	AuditActionRoleUpdate                  AuditAction = "role.update"
	AuditActionRoleDelete                  AuditAction = "role.delete"
	AuditActionRolePermissionAdd           AuditAction = "role.permission.add"
	AuditActionRolePermissionRemove        AuditAction = "role.permission.remove"
	AuditActionRoleAssign                  AuditAction = "role.assign"
	AuditActionRoleRevoke                  AuditAction = "role.revoke"
	AuditActionRoleChange                  AuditAction = "role.change"

	// Permission-related actions
	AuditActionPermissionCreate            AuditAction = "permission.create"
	AuditActionPermissionUpdate            AuditAction = "permission.update"
	AuditActionPermissionDelete            AuditAction = "permission.delete"
	AuditActionPermissionGrant             AuditAction = "permission.grant"
	AuditActionPermissionRevoke            AuditAction = "permission.revoke"

	// Session-related actions
	AuditActionSessionCreate               AuditAction = "session.create"
	AuditActionSessionUpdate               AuditAction = "session.update"
	AuditActionSessionDelete               AuditAction = "session.delete"
	AuditActionSessionExpire               AuditAction = "session.expire"
	AuditActionSessionInvalidate           AuditAction = "session.invalidate"
	AuditActionSessionRenew                AuditAction = "session.renew"

	// MFA-related actions
	AuditActionMfaEnable                   AuditAction = "mfa.enable"
	AuditActionMfaDisable                  AuditAction = "mfa.disable"
	AuditActionMfaVerify                   AuditAction = "mfa.verify"
	AuditActionMfaVerifyFailed             AuditAction = "mfa.verify.failed"
	AuditActionMfaMethodAdd                AuditAction = "mfa.method.add"
	AuditActionMfaMethodRemove             AuditAction = "mfa.method.remove"
	AuditActionMfaBackupCodesGenerate      AuditAction = "mfa.backup_codes.generate"
	AuditActionMfaBackupCodesUse           AuditAction = "mfa.backup_codes.use"

	// Security incident-related actions
	AuditActionSecurityIncidentCreate      AuditAction = "security.incident.create"
	AuditActionSecurityIncidentUpdate      AuditAction = "security.incident.update"
	AuditActionSecurityIncidentClose       AuditAction = "security.incident.close"
	AuditActionSecurityIncidentEscalate    AuditAction = "security.incident.escalate"
	AuditActionSecurityIncidentAssign      AuditAction = "security.incident.assign"
	AuditActionSecurityIncidentReopen      AuditAction = "security.incident.reopen"
	AuditActionSecurityIncidentComment     AuditAction = "security.incident.comment"
	AuditActionSecurityIncidentResolve     AuditAction = "security.incident.resolve"

	// Vulnerability-related actions
	AuditActionSecurityVulnerabilityCreate   AuditAction = "security.vulnerability.create"
	AuditActionSecurityVulnerabilityUpdate   AuditAction = "security.vulnerability.update"
	AuditActionSecurityVulnerabilityMitigate AuditAction = "security.vulnerability.mitigate"
	AuditActionSecurityVulnerabilityResolve  AuditAction = "security.vulnerability.resolve"
	AuditActionSecurityVulnerabilityEscalate AuditAction = "security.vulnerability.escalate"
	AuditActionSecurityVulnerabilityAssign   AuditAction = "security.vulnerability.assign"
	AuditActionSecurityVulnerabilityVerify   AuditAction = "security.vulnerability.verify"
	AuditActionSecurityVulnerabilityReopen   AuditAction = "security.vulnerability.reopen"

	// System-related actions
	AuditActionSystemStartup               AuditAction = "system.startup"
	AuditActionSystemShutdown              AuditAction = "system.shutdown"
	AuditActionSystemConfigChange          AuditAction = "system.config.change"
	AuditActionSystemBackup                AuditAction = "system.backup"
	AuditActionSystemRestore               AuditAction = "system.restore"
	AuditActionSystemUpgrade               AuditAction = "system.upgrade"
	AuditActionSystemMaintenance           AuditAction = "system.maintenance"
	AuditActionSystemAlert                 AuditAction = "system.alert"

	// Resource access actions
	AuditActionResourceCreate              AuditAction = "resource.create"
	AuditActionResourceRead                AuditAction = "resource.read"
	AuditActionResourceUpdate              AuditAction = "resource.update"
	AuditActionResourceDelete              AuditAction = "resource.delete"
	AuditActionResourceExecute             AuditAction = "resource.execute"
	AuditActionResourceAccess              AuditAction = "resource.access"
	AuditActionResourceAccessDenied        AuditAction = "resource.access.denied"

	// API-related actions
	AuditActionApiRequest                  AuditAction = "api.request"
	AuditActionApiResponse                 AuditAction = "api.response"
	AuditActionApiRateLimitExceeded        AuditAction = "api.rate_limit.exceeded"
	AuditActionApiKeyCreate                AuditAction = "api.key.create"
	AuditActionApiKeyUpdate                AuditAction = "api.key.update"
	AuditActionApiKeyDelete                AuditAction = "api.key.delete"
	AuditActionApiKeyRotate                AuditAction = "api.key.rotate"

	// Audit-related actions
	AuditActionAuditLogView                AuditAction = "audit.log.view"
	AuditActionAuditLogExport              AuditAction = "audit.log.export"
	AuditActionAuditLogPurge               AuditAction = "audit.log.purge"
	AuditActionAuditLogSearch              AuditAction = "audit.log.search"
	AuditActionAuditConfigChange           AuditAction = "audit.config.change"
	AuditActionAlert                       AuditAction = "audit.alert"
	AuditActionAlertRuleCreate             AuditAction = "audit.alert.rule.create"
	AuditActionAlertRuleUpdate             AuditAction = "audit.alert.rule.update"
	AuditActionAlertRuleDelete             AuditAction = "audit.alert.rule.delete"
	AuditActionAlertRuleEnable             AuditAction = "audit.alert.rule.enable"
	AuditActionAlertRuleDisable            AuditAction = "audit.alert.rule.disable"

	// Data-related actions
	AuditActionDataExport                  AuditAction = "data.export"
	AuditActionDataImport                  AuditAction = "data.import"
	AuditActionDataDelete                  AuditAction = "data.delete"
	AuditActionDataAnonymize               AuditAction = "data.anonymize"
	AuditActionDataEncrypt                 AuditAction = "data.encrypt"
	AuditActionDataDecrypt                 AuditAction = "data.decrypt"

	// Template-related actions
	AuditActionTemplateCreate              AuditAction = "template.create"
	AuditActionTemplateUpdate              AuditAction = "template.update"
	AuditActionTemplateDelete              AuditAction = "template.delete"
	AuditActionTemplateExecute             AuditAction = "template.execute"
	AuditActionTemplateApprove             AuditAction = "template.approve"
	AuditActionTemplateReject              AuditAction = "template.reject"

	// Login-related actions
	AuditActionLoginFailed                 AuditAction = "login.failed"
	AuditActionLoginSuccess                AuditAction = "login.success"
	AuditActionLogout                      AuditAction = "logout"
	AuditActionPasswordChange              AuditAction = "password.change"
	AuditActionPasswordReset               AuditAction = "password.reset"
	AuditActionPasswordExpired             AuditAction = "password.expired"
	AuditActionAccountLocked               AuditAction = "account.locked"
	AuditActionAccountUnlocked             AuditAction = "account.unlocked"
)

// Audit severity constants
const (
	AuditSeverityDebug     AuditSeverity = "debug"
	AuditSeverityInfo      AuditSeverity = "info"
	AuditSeverityNotice    AuditSeverity = "notice"
	AuditSeverityWarning   AuditSeverity = "warning"
	AuditSeverityError     AuditSeverity = "error"
	AuditSeverityCritical  AuditSeverity = "critical"
	AuditSeverityAlert     AuditSeverity = "alert"
	AuditSeverityEmergency AuditSeverity = "emergency"
	AuditSeverityLow       AuditSeverity = "low"
	AuditSeverityMedium    AuditSeverity = "medium"
	AuditSeverityHigh      AuditSeverity = "high"
)

// SeverityFromString converts a string to an AuditSeverity
func SeverityFromString(s string) AuditSeverity {
	switch s {
	case "debug":
		return AuditSeverityDebug
	case "info":
		return AuditSeverityInfo
	case "notice":
		return AuditSeverityNotice
	case "warning":
		return AuditSeverityWarning
	case "error":
		return AuditSeverityError
	case "critical":
		return AuditSeverityCritical
	case "alert":
		return AuditSeverityAlert
	case "emergency":
		return AuditSeverityEmergency
	case "low":
		return AuditSeverityLow
	case "medium":
		return AuditSeverityMedium
	case "high":
		return AuditSeverityHigh
	default:
		return AuditSeverityInfo
	}
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Action      AuditAction            `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	Result      string                 `json:"result"`
	Severity    AuditSeverity          `json:"severity"`
	Details     map[string]interface{} `json:"details"`
	IP          string                 `json:"ip"`
	UserAgent   string                 `json:"user_agent"`
	SessionID   string                 `json:"session_id"`
}
