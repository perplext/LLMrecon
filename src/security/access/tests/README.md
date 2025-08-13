# Access Control System Tests

This directory contains comprehensive tests for the LLMrecon Tool's access control and security auditing system.

## Test Structure

The test suite is organized into the following components:

- **Authentication Tests**: Tests for login, logout, token validation, and MFA verification
- **RBAC Tests**: Tests for role-based access control, permissions, and role hierarchy
- **User Management Tests**: Tests for user creation, updating, deletion, and password management
- **Session Management Tests**: Tests for session creation, validation, expiration, and refresh
- **Security Management Tests**: Tests for security incidents and vulnerabilities
- **Audit Logging Tests**: Tests for audit event logging, querying, and exporting
- **Boundary Enforcement Tests**: Tests for context boundary enforcement and prompt injection protection
- **API Tests**: Tests for the RESTful API endpoints

## Running Tests

### Running All Tests

To run all tests, use the following command from the project root:

```bash
go test -v ./src/security/access/tests/...
```

### Running Specific Test Groups

To run a specific group of tests, use the `-run` flag with the test name pattern:

```bash
# Run only authentication tests
go test -v ./src/security/access/tests -run TestAuthentication

# Run only API tests
go test -v ./src/security/access/tests -run TestAPI
```

### Running Individual Tests

To run a specific test case, use the `-run` flag with the test case name pattern:

```bash
# Run only the login success test
go test -v ./src/security/access/tests -run TestAuthentication/Login_AdminUser_Success
```

## Test Coverage

To generate test coverage reports, use the `-cover` flag:

```bash
go test -v -cover ./src/security/access/tests/...
```

For a detailed HTML coverage report:

```bash
go test -coverprofile=coverage.out ./src/security/access/tests/...
go tool cover -html=coverage.out -o coverage.html
```

## Test Helpers

The `test_helpers.go` file provides a `TestContext` structure that creates an isolated test environment with:

- In-memory SQLite database
- Initialized access control manager
- Test users and roles
- Helper methods for assertions and test setup

## Adding New Tests

When adding new tests:

1. Follow the existing test structure and naming conventions
2. Use the `TestContext` for test setup and teardown
3. Group related tests using `t.Run()`
4. Add assertions using the helper methods in `TestContext`
5. Ensure proper cleanup of resources

## Continuous Integration

These tests are automatically run as part of the CI/CD pipeline to ensure the access control system remains secure and functional with each change.
