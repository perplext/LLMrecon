package access

import (
	"net/http"
)

// MFAMiddleware represents middleware for MFA verification
type MFAMiddleware struct {
	authManager *AuthManager
}

// NewMFAMiddleware creates a new MFA middleware
func NewMFAMiddleware(authManager *AuthManager) *MFAMiddleware {
	return &MFAMiddleware{
		authManager: authManager,
	}
}

// RequireMFA creates middleware that requires MFA verification
func (m *MFAMiddleware) RequireMFA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session from cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate session
		session, err := m.authManager.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if MFA is completed for the session
		if !session.MFACompleted {
			// Get user
			user, err := m.authManager.GetUserByID(r.Context(), session.UserID)
			if err != nil {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}

			// Check if user has MFA enabled
			if user.MFAEnabled {
				// Redirect to MFA verification page
				http.Redirect(w, r, "/mfa/verify", http.StatusSeeOther)
				return
			}
		}

		// Add MFA status to context
		ctx := r.Context()
		if session.MFACompleted {
			ctx = WithMFAStatus(ctx, MFAStatusCompleted)
		} else {
			ctx = WithMFAStatus(ctx, MFAStatusNotRequired)
		}

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalMFA creates middleware that adds MFA status to context but doesn't require it
func (m *MFAMiddleware) OptionalMFA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session from cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// No session, just continue
			next.ServeHTTP(w, r)
			return
		}

		// Validate session
		session, err := m.authManager.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			// Invalid session, just continue
			next.ServeHTTP(w, r)
			return
		}

		// Add MFA status to context
		ctx := r.Context()
		if session.MFACompleted {
			ctx = WithMFAStatus(ctx, MFAStatusCompleted)
		} else {
			// Get user
			user, err := m.authManager.GetUserByID(r.Context(), session.UserID)
			if err == nil && user.MFAEnabled {
				ctx = WithMFAStatus(ctx, MFAStatusRequired)
			} else {
				ctx = WithMFAStatus(ctx, MFAStatusNotRequired)
			}
		}

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MFAVerifyHandler handles MFA verification
func (m *MFAMiddleware) MFAVerifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate session
	session, err := m.authManager.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get code from form
	code := r.Form.Get("code")

	// Verify MFA code
	if err := m.authManager.VerifyMFA(r.Context(), session.ID, code); err != nil {
		http.Error(w, "Invalid verification code", http.StatusUnauthorized)
		return
	}

	// Update session to indicate MFA is completed
	session.MFACompleted = true
	updates := map[string]interface{}{
		"mfa_completed": true,
	}
	if err := m.authManager.UpdateSession(r.Context(), session.ID, updates); err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	// Redirect to original destination or home page
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		redirectURL = "/"
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// MFARequiredFunc is a function that determines if MFA is required for a request
type MFARequiredFunc func(r *http.Request) bool

// ConditionalMFA creates middleware that requires MFA only if the condition function returns true
func (m *MFAMiddleware) ConditionalMFA(condition MFARequiredFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if condition(r) {
			m.RequireMFA(next).ServeHTTP(w, r)
		} else {
			m.OptionalMFA(next).ServeHTTP(w, r)
		}
	})
}

// MFAByRole creates middleware that requires MFA for users with specific roles
func (m *MFAMiddleware) MFAByRole(roles []string, next http.Handler) http.Handler {
	return m.ConditionalMFA(func(r *http.Request) bool {
		// Get session from cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			return false
		}

		// Validate session
		session, err := m.authManager.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			return false
		}

		// Get user
		user, err := m.authManager.GetUserByID(r.Context(), session.UserID)
		if err != nil {
			return false
		}

		// Check if user has any of the specified roles
		for _, role := range roles {
			hasRole, err := m.authManager.HasRole(r.Context(), user.ID, role)
			if err == nil && hasRole {
				return true
			}
		}

		return false
	}, next)
}

// MFAByPath creates middleware that requires MFA for specific paths
func (m *MFAMiddleware) MFAByPath(paths []string, next http.Handler) http.Handler {
	return m.ConditionalMFA(func(r *http.Request) bool {
		// Check if the request path matches any of the specified paths
		for _, path := range paths {
			if r.URL.Path == path {
				return true
			}
		}
		return false
	}, next)
}
