// Package access provides access control and security auditing functionality
package access

// Role constants
const (
	RoleAdmin      = "admin"
	RoleUser       = "user"
	RoleManager    = "manager"
	RoleOperator   = "operator"
	RoleAuditor    = "auditor"
	RoleGuest      = "guest"
	RoleAutomation = "automation"
)

// Permission constants
const (
	PermissionAll = "*"

	// System permissions
	PermissionSystemAdmin   = "system.admin"
	PermissionSystemConfig  = "system.config"
	PermissionSystemView    = "system.view"
	PermissionSystemMonitor = "system.monitor"
	PermissionSystemAudit   = "system.audit"
	PermissionSystemBackup  = "system.backup"
	PermissionSystemRestore = "system.restore"

	// User permissions
	PermissionUserView        = "user.view"
	PermissionUserCreate      = "user.create"
	PermissionUserRead        = "user.read"
	PermissionUserUpdate      = "user.update"
	PermissionUserDelete      = "user.delete"
	PermissionUserRoleAssign  = "user.role.assign"
	PermissionUserManageRoles = "user.manage.roles"

	// Role permissions
	PermissionRoleView   = "role.view"
	PermissionRoleCreate = "role.create"
	PermissionRoleUpdate = "role.update"
	PermissionRoleDelete = "role.delete"

	// Template permissions
	PermissionTemplateView    = "template.view"
	PermissionTemplateCreate  = "template.create"
	PermissionTemplateRead    = "template.read"
	PermissionTemplateUpdate  = "template.update"
	PermissionTemplateDelete  = "template.delete"
	PermissionTemplateExecute = "template.execute"
	PermissionTemplateApprove = "template.approve"
	PermissionTemplateUse     = "template.use"

	// Security permissions
	PermissionSecurityView          = "security.view"
	PermissionSecurityConfig        = "security.config"
	PermissionSecurityTest          = "security.test"
	PermissionSecurityIncident      = "security.incident"
	PermissionSecurityAudit         = "security.audit"
	PermissionSecurityVulnerability = "security.vulnerability"
	PermissionSecurityScan          = "security.scan"

	// Audit permissions
	PermissionAuditView   = "audit.view"
	PermissionAuditExport = "audit.export"

	// Vulnerability permissions
	PermissionVulnerabilityView = "vulnerability.view"

	// Prompt permissions
	PermissionPromptView = "prompt.view"
	PermissionPromptUse  = "prompt.use"

	// Test permissions
	PermissionTestExecute = "test.execute"

	// Report permissions
	PermissionReportView     = "report.view"
	PermissionReportCreate   = "report.create"
	PermissionReportRead     = "report.read"
	PermissionReportUpdate   = "report.update"
	PermissionReportDelete   = "report.delete"
	PermissionReportExport   = "report.export"
	PermissionReportGenerate = "report.generate"

	// Security incident permissions
	PermissionSecurityIncidentView   = "security.incident.view"
	PermissionSecurityIncidentCreate = "security.incident.create"
	PermissionSecurityIncidentUpdate = "security.incident.update"
	PermissionSecurityIncidentDelete = "security.incident.delete"

	// Security vulnerability permissions
	PermissionSecurityVulnerabilityView   = "security.vulnerability.view"
	PermissionSecurityVulnerabilityCreate = "security.vulnerability.create"
	PermissionSecurityVulnerabilityUpdate = "security.vulnerability.update"
	PermissionSecurityVulnerabilityDelete = "security.vulnerability.delete"

	// User management permissions
	PermissionUserList          = "user.list"
	PermissionUserResetPassword = "user.reset_password"
	PermissionUserLock          = "user.lock"
	PermissionUserUnlock        = "user.unlock"
	PermissionUserManageMFA     = "user.manage_mfa"

	// Role management permissions
	PermissionRoleList   = "role.list"
	PermissionRoleAssign = "role.assign"
)

// Severity string constants for API handler use
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// Status string constants for API handler use
const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"
	StatusPatched    = "patched"
)

// AllPermissions contains all defined permissions
var AllPermissions = []string{
	PermissionSystemAdmin,
	PermissionSystemConfig,
	PermissionSystemView,
	PermissionSystemMonitor,
	PermissionSystemAudit,
	PermissionSystemBackup,
	PermissionSystemRestore,
	PermissionUserView,
	PermissionUserCreate,
	PermissionUserRead,
	PermissionUserUpdate,
	PermissionUserDelete,
	PermissionUserRoleAssign,
	PermissionUserManageRoles,
	PermissionRoleView,
	PermissionRoleCreate,
	PermissionRoleUpdate,
	PermissionRoleDelete,
	PermissionTemplateView,
	PermissionTemplateCreate,
	PermissionTemplateRead,
	PermissionTemplateUpdate,
	PermissionTemplateDelete,
	PermissionTemplateExecute,
	PermissionTemplateApprove,
	PermissionTemplateUse,
	PermissionSecurityView,
	PermissionSecurityConfig,
	PermissionSecurityTest,
	PermissionSecurityIncident,
	PermissionSecurityAudit,
	PermissionSecurityVulnerability,
	PermissionSecurityScan,
	PermissionAuditView,
	PermissionVulnerabilityView,
	PermissionPromptView,
	PermissionPromptUse,
	PermissionTestExecute,
	PermissionReportView,
	PermissionReportCreate,
	PermissionReportRead,
	PermissionReportUpdate,
	PermissionReportDelete,
	PermissionReportExport,
	PermissionReportGenerate,
}
