// Package impl provides implementations of the security access interfaces
package impl

import (
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
)

// Factory creates adapter instances
type Factory struct {
	converter *DefaultConverter
}

// NewFactory creates a new factory
func NewFactory() *Factory {
	return &Factory{
		converter: NewDefaultConverter(),
	}

// CreateUserStoreAdapter creates a new user store adapter
func (f *Factory) CreateUserStoreAdapter(legacyStore interface{}) interfaces.UserStore {
	return NewUserStoreAdapter(legacyStore, f.converter)

// CreateSessionStoreAdapter creates a new session store adapter
func (f *Factory) CreateSessionStoreAdapter(legacyStore interface{}) interfaces.SessionStore {
	return NewSessionStoreAdapter(legacyStore, f.converter)

// CreateSecurityManagerAdapter creates a new security manager adapter
func (f *Factory) CreateSecurityManagerAdapter(legacyManager interface{}) interfaces.SecurityManager {
	return NewSecurityManagerAdapter(legacyManager, f.converter)

// GetConverter returns the default converter
func (f *Factory) GetConverter() *DefaultConverter {
	return f.converter
}
}
}
}
