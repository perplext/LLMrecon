package access

import (
	"context"
	"fmt"
	"net/http"
)

// ContextBoundaryType represents the type of boundary being enforced
type ContextBoundaryType string

const (
	// BoundaryTypeSession represents a session boundary
	BoundaryTypeSession ContextBoundaryType = "session"
	// BoundaryTypeUser represents a user boundary
	BoundaryTypeUser ContextBoundaryType = "user"
	// BoundaryTypeRole represents a role boundary
	BoundaryTypeRole ContextBoundaryType = "role"
	// BoundaryTypePermission represents a permission boundary
	BoundaryTypePermission ContextBoundaryType = "permission"
	// BoundaryTypeMFA represents an MFA boundary
	BoundaryTypeMFA ContextBoundaryType = "mfa"
)

// ContextBoundary represents a security boundary that must be enforced
type ContextBoundary struct {
	Type       ContextBoundaryType
	Value      string
	Required   bool
	Validation func(ctx context.Context, value string) bool
}

// Note: User, AuthManager, and RBACManager types are defined in other files
// This file uses the existing concrete types to avoid redeclaration conflicts

// EnhancedContextBoundaryEnforcer enforces security boundaries for requests
type EnhancedContextBoundaryEnforcer struct {
	sessionManager *SessionManager
	authManager    *AuthManager
	rbacManager    *RBACManagerImpl
	boundaries     map[ContextBoundaryType][]ContextBoundary
}

// NewEnhancedContextBoundaryEnforcer creates a new boundary enforcer
func NewEnhancedContextBoundaryEnforcer() *EnhancedContextBoundaryEnforcer {
	return &EnhancedContextBoundaryEnforcer{
		boundaries: make(map[ContextBoundaryType][]ContextBoundary),
	}
}

// SetSessionManager sets the session manager for the boundary enforcer
func (e *EnhancedContextBoundaryEnforcer) SetSessionManager(sessionManager *SessionManager) {
	e.sessionManager = sessionManager
}

// SetAuthManager sets the auth manager for the boundary enforcer
func (e *EnhancedContextBoundaryEnforcer) SetAuthManager(authManager *AuthManager) {
	e.authManager = authManager
}

// SetRBACManager sets the RBAC manager for the boundary enforcer
func (e *EnhancedContextBoundaryEnforcer) SetRBACManager(rbacManager *RBACManagerImpl) {
	e.rbacManager = rbacManager
}

// AddBoundary adds a new boundary to enforce
func (e *EnhancedContextBoundaryEnforcer) AddBoundary(boundary ContextBoundary) {
	if _, ok := e.boundaries[boundary.Type]; !ok {
		e.boundaries[boundary.Type] = []ContextBoundary{}
	}
	e.boundaries[boundary.Type] = append(e.boundaries[boundary.Type], boundary)
}

// EnforceBoundaries enforces all boundaries for a request
func (e *EnhancedContextBoundaryEnforcer) EnforceBoundaries(ctx context.Context) error {
	// Check session boundary
	if boundaries, ok := e.boundaries[BoundaryTypeSession]; ok {
		for _, boundary := range boundaries {
			if !boundary.Validation(ctx, boundary.Value) {
				if boundary.Required {
					return fmt.Errorf("session boundary validation failed")
				}
			}
		}
	}

	// Check user boundary
	if boundaries, ok := e.boundaries[BoundaryTypeUser]; ok {
		for _, boundary := range boundaries {
			if !boundary.Validation(ctx, boundary.Value) {
				if boundary.Required {
					return fmt.Errorf("user boundary validation failed")
				}
			}
		}
	}

	// Check role boundary
	if boundaries, ok := e.boundaries[BoundaryTypeRole]; ok {
		for _, boundary := range boundaries {
			if !boundary.Validation(ctx, boundary.Value) {
				if boundary.Required {
					return fmt.Errorf("role boundary validation failed")
				}
			}
		}
	}

	// Check permission boundary
	if boundaries, ok := e.boundaries[BoundaryTypePermission]; ok {
		for _, boundary := range boundaries {
			if !boundary.Validation(ctx, boundary.Value) {
				if boundary.Required {
					return fmt.Errorf("permission boundary validation failed")
				}
			}
		}
	}

	// Check MFA boundary
	if boundaries, ok := e.boundaries[BoundaryTypeMFA]; ok {
		for _, boundary := range boundaries {
			if !boundary.Validation(ctx, boundary.Value) {
				if boundary.Required {
					return fmt.Errorf("MFA boundary validation failed")
				}
			}
		}
	}

	return nil
}

// Middleware creates an HTTP middleware that enforces boundaries
func (e *EnhancedContextBoundaryEnforcer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := e.EnforceBoundaries(r.Context()); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireSession creates a boundary that requires a valid session
func (e *EnhancedContextBoundaryEnforcer) RequireSession() ContextBoundary {
	return ContextBoundary{
		Type:     BoundaryTypeSession,
		Required: true,
		Validation: func(ctx context.Context, value string) bool {
			session, err := e.sessionManager.GetSessionFromContext(ctx)
			return err == nil && session != nil && !session.IsExpired()
		},
	}
}

// RequireUser creates a boundary that requires a specific user
func (e *EnhancedContextBoundaryEnforcer) RequireUser(userID string) ContextBoundary {
	return ContextBoundary{
		Type:     BoundaryTypeUser,
		Value:    userID,
		Required: true,
		Validation: func(ctx context.Context, value string) bool {
			session, err := e.sessionManager.GetSessionFromContext(ctx)
			return err == nil && session != nil && session.UserID == value
		},
	}
}

// RequireRole creates a boundary that requires a specific role
func (e *EnhancedContextBoundaryEnforcer) RequireRole(role string) ContextBoundary {
	return ContextBoundary{
		Type:     BoundaryTypeRole,
		Value:    role,
		Required: true,
		Validation: func(ctx context.Context, value string) bool {
			session, err := e.sessionManager.GetSessionFromContext(ctx)
			if err != nil || session == nil {
				return false
			}

			user, err := e.authManager.GetUserByID(ctx, session.UserID)
			if err != nil {
				return false
			}

			hasRole, err := e.rbacManager.HasRole(ctx, user.ID, value)
			if err != nil {
				return false
			}
			return hasRole
		},
	}
}

// RequirePermission creates a boundary that requires a specific permission
func (e *EnhancedContextBoundaryEnforcer) RequirePermission(permission string) ContextBoundary {
	return ContextBoundary{
		Type:     BoundaryTypePermission,
		Value:    permission,
		Required: true,
		Validation: func(ctx context.Context, value string) bool {
			session, err := e.sessionManager.GetSessionFromContext(ctx)
			if err != nil || session == nil {
				return false
			}

			user, err := e.authManager.GetUserByID(ctx, session.UserID)
			if err != nil {
				return false
			}

			hasPermission, err := e.rbacManager.HasPermission(ctx, user.ID, value)
			if err != nil {
				return false
			}
			return hasPermission
		},
	}
}

// RequireMFA creates a boundary that requires MFA completion
func (e *EnhancedContextBoundaryEnforcer) RequireMFA() ContextBoundary {
	return ContextBoundary{
		Type:     BoundaryTypeMFA,
		Required: true,
		Validation: func(ctx context.Context, value string) bool {
			session, err := e.sessionManager.GetSessionFromContext(ctx)
			if err != nil || session == nil {
				return false
			}

			return session.MFACompleted
		},
	}
}
