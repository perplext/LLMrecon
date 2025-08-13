// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/security/access/adapters"
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/types"
)

// SecurityManagerFactory creates security manager instances
type SecurityManagerFactory struct {
	config *types.SecurityConfig
}

// NewSecurityManagerFactory creates a new security manager factory
func NewSecurityManagerFactory(config *types.SecurityConfig) *SecurityManagerFactory {
	if config == nil {
		config = &types.SecurityConfig{}
	}

	return &SecurityManagerFactory{
		config: config,
	}
}

// CreateSecurityManager creates a security manager based on the specified implementation type
func (f *SecurityManagerFactory) CreateSecurityManager(
	ctx context.Context,
	implementationType string,
	incidentStore interfaces.IncidentStore,
	vulnerabilityStore interfaces.VulnerabilityStore,
	auditLogger interfaces.AuditLogger,
) (interfaces.SecurityManager, error) {
	switch implementationType {
	case "legacy":
		return f.createLegacySecurityManager(ctx, incidentStore, vulnerabilityStore, auditLogger)
	case "new":
		return f.createNewSecurityManager(ctx, incidentStore, vulnerabilityStore, auditLogger)
	default:
		// Default to the new implementation
		return f.createNewSecurityManager(ctx, incidentStore, vulnerabilityStore, auditLogger)
	}
}

// createLegacySecurityManager creates a legacy security manager with an adapter
func (f *SecurityManagerFactory) createLegacySecurityManager(
	ctx context.Context,
	incidentStore interfaces.IncidentStore,
	vulnerabilityStore interfaces.VulnerabilityStore,
	auditLogger interfaces.AuditLogger,
) (interfaces.SecurityManager, error) {
	// Create a legacy security manager
	// Note: This is a simplified version, in a real implementation we would need to
	// create adapters for the stores and logger to convert between the new and legacy types
	legacyManager := NewSecurityManager(
		&AccessControlConfig{
			// Map the security config to access control config
			// This is a simplified mapping, in a real implementation we would need to
			// properly map all fields between the two config types
			SecurityConfig: &models.SecurityConfig{
				IncidentEscalationThreshold: f.config.IncidentEscalationThreshold,
			},
		},
		NewInMemoryIncidentStore(),
		NewInMemoryVulnerabilityStore(),
		NewInMemoryAuditLogger(),
	)

	// Create an adapter for the legacy security manager
	adapter := adapters.NewSecurityManagerAdapter(legacyManager)

	// Initialize the adapter
	if err := adapter.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize security manager adapter: %w", err)
	}

	return adapter, nil
}

// createNewSecurityManager creates a new security manager
func (f *SecurityManagerFactory) createNewSecurityManager(
	ctx context.Context,
	incidentStore interfaces.IncidentStore,
	vulnerabilityStore interfaces.VulnerabilityStore,
	auditLogger interfaces.AuditLogger,
) (interfaces.SecurityManager, error) {
	// Create a new security manager
	manager, err := NewSecurityManagerImpl(
		incidentStore,
		vulnerabilityStore,
		auditLogger,
		f.config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create security manager: %w", err)
	}

	// Initialize the manager
	if err := manager.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize security manager: %w", err)
	}

	return manager, nil
}

// CreateInMemorySecurityManager creates a security manager with in-memory stores
func (f *SecurityManagerFactory) CreateInMemorySecurityManager(
	ctx context.Context,
	implementationType string,
) (interfaces.SecurityManager, error) {
	// Create in-memory stores
	incidentStore := NewInMemoryIncidentStoreAdapter()
	vulnerabilityStore := NewInMemoryVulnerabilityStoreAdapter()
	auditLogger := NewInMemoryAuditLoggerAdapter()

	// Create the security manager
	return f.CreateSecurityManager(
		ctx,
		implementationType,
		incidentStore,
		vulnerabilityStore,
		auditLogger,
	)
}

// NewFactoryInMemoryIncidentStoreAdapter creates a new in-memory incident store adapter from factory
func NewFactoryInMemoryIncidentStoreAdapter() interfaces.IncidentStore {
	return adapters.NewInMemoryIncidentStoreAdapter()
}

// NewFactoryInMemoryVulnerabilityStoreAdapter creates a new in-memory vulnerability store adapter from factory
func NewFactoryInMemoryVulnerabilityStoreAdapter() interfaces.VulnerabilityStore {
	return adapters.NewInMemoryVulnerabilityStoreAdapter()
}

// NewFactoryInMemoryAuditLoggerAdapter creates a new in-memory audit logger adapter from factory
func NewFactoryInMemoryAuditLoggerAdapter() interfaces.AuditLogger {
	return adapters.NewInMemoryAuditLoggerAdapter()
}
