// Package db provides stub implementations for testing
package db

import (
	"database/sql"
	"fmt"

	"github.com/perplext/LLMrecon/src/security/access/db/adapter"
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
)

// NewSQLUserStore creates a stub user store for compilation
func NewSQLUserStore(db *sql.DB) (interfaces.UserStore, error) {
	return nil, fmt.Errorf("NewSQLUserStore not implemented")
}

// NewSQLSessionStore creates a stub session store for compilation
func NewSQLSessionStore(db *sql.DB) (adapter.SessionStore, error) {
	return nil, fmt.Errorf("NewSQLSessionStore not implemented")
}

// NewSQLAuditStore creates a stub audit store for compilation
func NewSQLAuditStore(db *sql.DB) (adapter.ModelsAuditStore, error) {
	return nil, fmt.Errorf("NewSQLAuditStore not implemented")
}

// NewSQLIncidentStore creates a stub incident store for compilation
func NewSQLIncidentStore(db *sql.DB) (adapter.IncidentStore, error) {
	return nil, fmt.Errorf("NewSQLIncidentStore not implemented")
}

// NewSQLVulnerabilityStore creates a stub vulnerability store for compilation
func NewSQLVulnerabilityStore(db *sql.DB) (interfaces.VulnerabilityStore, error) {
	return nil, fmt.Errorf("NewSQLVulnerabilityStore not implemented")
}