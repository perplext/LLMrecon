// Package cli provides command-line interfaces for bundle operations
package cli

import (
	"fmt"

	"github.com/perplext/LLMrecon/src/security/access/audit/trail"
)

// RunOfflineBundleCLI runs the offline bundle CLI
func RunOfflineBundleCLI() {
	// Create audit trail manager
	auditConfig := &trail.AuditConfig{
		Enabled:        true,
		LoggingBackend: "file",
		LogDirectory:   "logs/audit",
		RetentionDays:  90,
		SigningEnabled: true,
	}
	
	auditTrailManager, err := trail.NewAuditTrailManager(auditConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create audit trail manager: %v\n", err)
		fmt.Fprintf(os.Stderr, "Continuing without audit logging...\n")
		auditTrailManager = nil
	}

	// Create CLI
	cli := NewOfflineBundleCLI(os.Stdout, auditTrailManager)

	// Execute CLI
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
