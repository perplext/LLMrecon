# Version Checking Mechanism Design

## Overview

The Version Checking Mechanism is a critical component of the LLMreconing Tool's update system. It allows the tool to securely check for updates to the core binary, templates, and provider modules from both GitHub and GitLab sources.

## Requirements

1. **Semantic Versioning Support**
   - Follow SemVer principles (major.minor.patch)
   - Support pre-release identifiers and build metadata
   - Provide version comparison logic

2. **API Endpoints for Version Queries**
   - Design RESTful API endpoints for version checking
   - Support both GitHub and GitLab as update sources
   - Allow checking for specific components (core, templates, modules)

3. **Security Features**
   - Timestamp verification to prevent replay attacks
   - Request/response signing using HMAC-SHA256
   - Support for required updates with minimum version enforcement
   - Secure communication over HTTPS

4. **Version Comparison Logic**
   - Determine if updates are available
   - Classify update types (major, minor, patch)
   - Support for compatibility checking between components

## Implementation Details

### Version Check Service

The `VersionCheckService` is the main component responsible for checking available updates. It:

1. Prepares a version check request with current version information
2. Signs the request using HMAC-SHA256 if a secret key is provided
3. Sends the request to the update server
4. Verifies the response signature and timestamp
5. Processes the response to determine available updates

### Security Measures

1. **Timestamp Verification**
   - Each request/response includes a timestamp
   - The service verifies that the timestamp is within an acceptable range (5 minutes by default)
   - This prevents replay attacks where an attacker might capture and replay an old response

2. **Request/Response Signing**
   - Both requests and responses can be signed using HMAC-SHA256
   - The signature is calculated over the JSON representation of the request/response
   - The service verifies the signature before processing the response

3. **Required Updates**
   - The update server can mark certain updates as required
   - Required updates include a minimum version that must be installed
   - The service identifies required updates and flags them accordingly

### API Endpoints

The version checking mechanism uses the following API endpoint:

- `POST /api/v1/check-version` - Checks for available updates

The request includes:
- Client ID
- Current versions of components
- Components to check
- Timestamp
- Request signature (optional)

The response includes:
- Available versions for requested components
- Server timestamp
- Response signature (optional)

## Integration with Existing Update System

The new version checking mechanism integrates with the existing update system by:

1. Enhancing the `UpdateInfo` struct to include a `Required` field
2. Providing a more secure alternative to the existing `VersionChecker`
3. Supporting both the existing version manifest format and the new API-based format
4. Maintaining backward compatibility with existing update sources

## Testing Strategy

1. **Unit Tests**
   - Test version parsing and comparison
   - Test request/response signing and verification
   - Test timestamp validation

2. **Integration Tests**
   - Test with mock update servers
   - Test with different response scenarios (no updates, regular updates, required updates)
   - Test error handling for network failures, invalid signatures, etc.

3. **Security Tests**
   - Test replay attack prevention
   - Test signature verification
   - Test handling of malformed responses
