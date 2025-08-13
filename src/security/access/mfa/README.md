# Multi-Factor Authentication (MFA) System

This package implements a comprehensive Multi-Factor Authentication (MFA) system for the LLMrecon tool, supporting multiple authentication methods to enhance security.

## Supported MFA Methods

1. **TOTP (Time-based One-Time Password)**
   - Compatible with standard authenticator apps like Google Authenticator, Authy, and Microsoft Authenticator
   - Implements RFC 6238 standard
   - Includes QR code generation for easy setup

2. **Backup Codes**
   - Provides one-time use recovery codes
   - Automatically invalidates codes after use
   - Allows users to regenerate codes when needed

3. **WebAuthn/FIDO2**
   - Supports hardware security keys and platform authenticators
   - Implements the Web Authentication API standard
   - Provides phishing-resistant authentication

4. **SMS Verification**
   - Sends one-time codes via SMS
   - Includes code expiration for enhanced security
   - Supports phone number management

## Architecture

The MFA system is designed with a modular architecture:

```
mfa/
├── interface.go     # MFA Manager interface definition
├── manager.go       # Main MFA Manager implementation
├── mock_manager.go  # Mock implementation for testing
├── totp.go          # TOTP implementation
├── backup.go        # Backup codes implementation
├── webauthn.go      # WebAuthn implementation
├── sms.go           # SMS verification implementation
└── manager_test.go  # Tests for the MFA system
```

### Core Components

1. **MFAManager Interface**
   - Defines the contract for all MFA operations
   - Allows for easy mocking and testing
   - Enables future extension with new authentication methods

2. **Manager Implementation**
   - Centralizes MFA logic and method selection
   - Handles user MFA preferences and settings
   - Manages MFA verification workflow

3. **Method-Specific Implementations**
   - Each MFA method has its own implementation file
   - Follows best practices for each authentication type
   - Includes proper error handling and security measures

## Integration with Authentication System

The MFA system integrates with the existing authentication system through:

1. **AuthManager**
   - Extended to include MFA verification in the authentication flow
   - Maintains backward compatibility with existing code
   - Enforces MFA requirements based on user roles and settings

2. **MFA Middleware**
   - Provides HTTP middleware for enforcing MFA requirements
   - Supports conditional MFA based on roles, paths, or custom logic
   - Handles MFA verification redirects and session updates

3. **API Endpoints**
   - Exposes RESTful endpoints for managing MFA settings
   - Provides verification endpoints for each MFA method
   - Includes proper error handling and security measures

## Security Considerations

1. **Secret Management**
   - TOTP secrets are stored securely using encryption
   - Backup codes are hashed before storage
   - WebAuthn credentials use secure storage mechanisms

2. **Rate Limiting**
   - Verification attempts are rate-limited to prevent brute force attacks
   - Failed attempts are logged for security auditing

3. **Session Management**
   - MFA status is tracked in the user's session
   - Sessions are invalidated on suspicious activity
   - MFA completion is required for sensitive operations

## Usage Examples

### Enabling MFA for a User

```go
// Initialize MFA manager
mfaManager := mfa.NewMFAManager(db)

// Generate TOTP secret for a user
secret, err := mfaManager.GenerateTOTPSecret(userID)
if err != nil {
    // Handle error
}

// Generate QR code URL for the user
qrCodeURL := mfaManager.GenerateTOTPQRCodeURL(username, secret)

// After user verifies the code, enable MFA
user.MFAEnabled = true
user.MFAMethods = append(user.MFAMethods, common.AuthMethodTOTP)
user.MFASecret = secret
```

### Verifying MFA During Login

```go
// After successful password verification
if user.MFAEnabled {
    // Get verification code from user
    valid, err := mfaManager.VerifyMFACode(ctx, user.ID, method, code)
    if err != nil || !valid {
        // Handle invalid code
        return errors.New("invalid verification code")
    }
    
    // Mark MFA as completed in session
    session.MFACompleted = true
    authManager.UpdateSession(ctx, session)
}
```

### Using MFA Middleware

```go
// Create MFA middleware
mfaMiddleware := access.NewMFAMiddleware(authManager)

// Require MFA for all admin routes
adminRouter := mux.NewRouter()
adminRouter.Use(mfaMiddleware.RequireMFA)
adminRouter.HandleFunc("/admin/dashboard", adminDashboardHandler)

// Require MFA only for specific roles
sensitiveRouter := mux.NewRouter()
sensitiveRouter.Use(mfaMiddleware.MFAByRole([]access.Role{access.RoleAdmin, access.RoleManager}))
sensitiveRouter.HandleFunc("/sensitive/data", sensitiveDataHandler)

// Require MFA only for specific paths
router := mux.NewRouter()
router.Use(mfaMiddleware.MFAByPath([]string{"/settings/security", "/settings/api-keys"}))
```

## Testing

The MFA system includes comprehensive tests for all components:

- Unit tests for each MFA method
- Integration tests with the authentication system
- Mock implementations for testing without external dependencies

Run the tests with:

```bash
go test -v ./src/security/access/mfa/...
```
