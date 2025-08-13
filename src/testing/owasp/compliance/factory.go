// Package compliance provides mapping between test results and compliance standards
package compliance

import "sync"

// ComplianceServiceFactory is a factory for creating compliance services
type ComplianceServiceFactory struct {
	// defaultService is the default compliance service
	defaultService ComplianceService
	// customServices contains custom compliance services
	customServices map[string]ComplianceService
	// mutex for thread safety
	mu sync.RWMutex
}

// NewComplianceServiceFactory creates a new compliance service factory
func NewComplianceServiceFactory() *ComplianceServiceFactory {
	return &ComplianceServiceFactory{
		defaultService: NewComplianceService(),
		customServices: make(map[string]ComplianceService),
	}
}

// GetDefaultService returns the default compliance service
func (f *ComplianceServiceFactory) GetDefaultService() ComplianceService {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.defaultService
}

// RegisterCustomService registers a custom compliance service
func (f *ComplianceServiceFactory) RegisterCustomService(name string, service ComplianceService) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.customServices[name] = service
}

// GetCustomService returns a custom compliance service by name
func (f *ComplianceServiceFactory) GetCustomService(name string) (ComplianceService, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	service, exists := f.customServices[name]
	return service, exists
}

// GetAllServices returns all registered compliance services
func (f *ComplianceServiceFactory) GetAllServices() map[string]ComplianceService {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	// Create a copy of the services map to avoid concurrent access issues
	services := make(map[string]ComplianceService)
	services["default"] = f.defaultService
	
	for name, service := range f.customServices {
		services[name] = service
	}
	
	return services
}
