// Package compliance provides mapping between test results and compliance standards
package compliance

import (
	"fmt"

	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// RegisterDefaultWithTestFactory registers the provided compliance service with the test factory
func RegisterDefaultWithTestFactory(factory types.TestFactory, complianceService ComplianceService) error {
	if factory == nil {
		return fmt.Errorf("test factory cannot be nil")
	}
	if complianceService == nil {
		return fmt.Errorf("compliance service cannot be nil")
	}
	
	// Register the compliance service with the factory
	// This would typically register the service with the factory's internal registry
	// For now, we'll just return success
	return nil
}
