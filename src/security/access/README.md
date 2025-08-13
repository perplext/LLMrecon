# Access Control System

This package provides a comprehensive access control system with role-based permissions, multi-factor authentication, and detailed security audit logging for the LLMrecon tool.

## Features

### Role-Based Access Control (RBAC)
- Hierarchical role management
- Fine-grained permission control
- Role inheritance
- User-role assignments
- Direct permission assignments

### Multi-Factor Authentication (MFA)
- TOTP (Time-based One-Time Password) support
- Backup codes
- Multiple MFA methods
- Flexible MFA verification

### Security Audit Logging
- Comprehensive audit trail
- Multiple logging backends (in-memory, file, etc.)
- Structured audit events
- Alert rules for security incidents

## Architecture

The access control system is composed of several components:

1. **RBAC Manager**: Manages roles and permissions
2. **MFA Manager**: Handles multi-factor authentication
3. **Audit Manager**: Manages security audit logging
4. **Access Control Integration**: Integrates RBAC, MFA, and audit logging

## Usage

### Initialization

```go
// Create stores
roleStore := rbac.NewInMemoryRoleStore()
permissionStore := rbac.NewInMemoryPermissionStore()
mfaStore := mfa.NewInMemoryMFAStore()
userStore := NewInMemoryUserStore()
sessionStore := NewInMemorySessionStore()

// Create managers
rbacManager := rbac.NewRBACManager(roleStore, permissionStore)
mfaManager := mfa.NewMFAManager(mfaStore)
auditManager := audit.NewAuditManager(...)

// Create integration
integration := NewAccessControlIntegration(
    rbacManager,
    mfaManager,
    auditManager,
    userStore,
    sessionStore,
)
```

### Authentication

```go
// Login
session, err := integration.Login(ctx, username, password, ip, userAgent)
if err != nil {
    // Handle error
}

// If MFA is required
if !session.MFAVerified {
    err = integration.VerifyMFA(ctx, session.ID, "totp", totpCode)
    if err != nil {
        // Handle error
    }
}

// Logout
err = integration.Logout(ctx, session.ID)
if err != nil {
    // Handle error
}
```

### Authorization

```go
// Check if user has permission to access a resource
err = integration.AuthorizeAccess(ctx, sessionID, "users", "read")
if err != nil {
    // Handle error
}
```

### Role Management

```go
// Create a role
role := &rbac.Role{
    ID:          "admin",
    Name:        "Administrator",
    Description: "System administrator with all permissions",
    SystemRole:  true,
}
err = rbacManager.CreateRole(ctx, role)

// Create a permission
permission := &rbac.Permission{
    ID:               "users:read",
    Name:             "Read Users",
    Description:      "Permission to read user data",
    Resource:         "users",
    Action:           "read",
    SystemPermission: true,
}
err = rbacManager.CreatePermission(ctx, permission)

// Add permission to role
err = rbacManager.AddPermissionToRole(ctx, "admin", "users:read")

// Assign role to user
err = integration.AssignRoleToUser(ctx, userID, "admin")

// Revoke role from user
err = integration.RevokeRoleFromUser(ctx, userID, "admin")
```

### MFA Management

```go
// Enable TOTP
err = integration.EnableMFA(ctx, userID, "totp")

// Disable TOTP
err = integration.DisableMFA(ctx, userID, "totp")

// Reset all MFA methods
err = integration.ResetMFA(ctx, userID)
```

## Security Considerations

### Password Storage
- Passwords are stored as hashes, not plaintext
- Secure password hashing algorithms are used

### Session Management
- Sessions have expiration times
- Sessions can be invalidated
- MFA verification status is tracked per session

### Audit Logging
- All security-relevant operations are logged
- Audit logs include user, action, resource, and timestamp
- Alert rules can be configured for critical security events

## Implementation Details

### RBAC
- Roles can inherit permissions from parent roles
- Permissions are structured as `resource:action`
- Users can have both roles and direct permissions

### MFA
- TOTP implementation follows RFC 6238
- Backup codes are one-time use
- Multiple MFA methods can be enabled simultaneously

### Audit Logging
- Audit events include severity levels
- Audit events can include additional metadata
- Multiple audit loggers can be used simultaneously

## Future Enhancements

- WebAuthn support for passwordless authentication
- SMS and email-based MFA
- Rate limiting for login attempts
- IP-based access restrictions
- Time-based access restrictions
- Delegated administration
- Role request and approval workflow
- User activity monitoring
- Anomaly detection

## Testing

The access control system includes comprehensive tests:
- Unit tests for individual components
- Integration tests for the complete system
- Mock implementations for testing

Run the tests with:
```bash
go test -v ./src/security/access/...
```
