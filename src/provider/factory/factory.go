// Package factory provides functionality for creating provider instances.
package factory

import (
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/provider/config"
	"github.com/perplext/LLMrecon/src/provider/core"
)

// ProviderFactory is responsible for creating provider instances
type ProviderFactory struct {
	// configManager is the configuration manager
	configManager *config.ConfigManager
	// providers is a map of provider types to provider constructors
	providers map[core.ProviderType]ProviderConstructor
	// instances is a cache of provider instances
	instances map[core.ProviderType]core.Provider
	// mutex is a mutex for concurrent access to instances
	mutex sync.RWMutex

// ProviderConstructor is a function that creates a provider instance
type ProviderConstructor func(config *core.ProviderConfig) (core.Provider, error)

// NewProviderFactory creates a new provider factory
func NewProviderFactory(configManager *config.ConfigManager) *ProviderFactory {
	return &ProviderFactory{
		configManager: configManager,
		providers:     make(map[core.ProviderType]ProviderConstructor),
		instances:     make(map[core.ProviderType]core.Provider),
	}

// RegisterProvider registers a provider constructor
func (f *ProviderFactory) RegisterProvider(providerType core.ProviderType, constructor ProviderConstructor) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.providers[providerType] = constructor

// CreateProvider creates a provider instance
func (f *ProviderFactory) CreateProvider(providerType core.ProviderType) (core.Provider, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Check if provider is already instantiated
	if provider, ok := f.instances[providerType]; ok {
		return provider, nil
	}

	// Check if provider constructor is registered
	constructor, ok := f.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("provider constructor for type %s not registered", providerType)
	}

	// Get provider configuration
	config, err := f.configManager.GetConfig(providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider configuration: %w", err)
	}

	// Create provider instance
	provider, err := constructor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider instance: %w", err)
	}

	// Cache provider instance
	f.instances[providerType] = provider

	return provider, nil

// GetProvider gets a provider instance
func (f *ProviderFactory) GetProvider(providerType core.ProviderType) (core.Provider, error) {
	f.mutex.RLock()
	provider, ok := f.instances[providerType]
	f.mutex.RUnlock()

	if ok {
		return provider, nil
	}

	return f.CreateProvider(providerType)

// CloseProvider closes a provider instance
func (f *ProviderFactory) CloseProvider(providerType core.ProviderType) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	provider, ok := f.instances[providerType]
	if !ok {
		return fmt.Errorf("provider instance for type %s not found", providerType)
	}

	if err := provider.Close(); err != nil {
		return fmt.Errorf("failed to close provider instance: %w", err)
	}

	delete(f.instances, providerType)

	return nil

// CloseAllProviders closes all provider instances
func (f *ProviderFactory) CloseAllProviders() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	var errs []error

	for providerType, provider := range f.instances {
		if err := provider.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close provider instance %s: %w", providerType, err))
		}
	}

	f.instances = make(map[core.ProviderType]core.Provider)

	if len(errs) > 0 {
		return fmt.Errorf("failed to close all provider instances: %v", errs)
	}

	return nil

// GetSupportedProviderTypes returns all supported provider types
func (f *ProviderFactory) GetSupportedProviderTypes() []core.ProviderType {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	providerTypes := make([]core.ProviderType, 0, len(f.providers))
	for providerType := range f.providers {
		providerTypes = append(providerTypes, providerType)
	}

	return providerTypes

// IsProviderSupported returns whether a provider type is supported
func (f *ProviderFactory) IsProviderSupported(providerType core.ProviderType) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	_, ok := f.providers[providerType]
	return ok

// RefreshProvider refreshes a provider instance
func (f *ProviderFactory) RefreshProvider(providerType core.ProviderType) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	// Close existing provider instance
	if provider, ok := f.instances[providerType]; ok {
		if err := provider.Close(); err != nil {
			return fmt.Errorf("failed to close provider instance: %w", err)
		}
		delete(f.instances, providerType)
	}

	// Check if provider constructor is registered
	constructor, ok := f.providers[providerType]
	if !ok {
		return fmt.Errorf("provider constructor for type %s not registered", providerType)
	}

	// Get provider configuration
	config, err := f.configManager.GetConfig(providerType)
	if err != nil {
		return fmt.Errorf("failed to get provider configuration: %w", err)
	}

	// Create provider instance
	provider, err := constructor(config)
	if err != nil {
		return fmt.Errorf("failed to create provider instance: %w", err)
	}

	// Cache provider instance
	f.instances[providerType] = provider

	return nil

// UpdateProviderConfig updates a provider configuration and refreshes the provider instance
func (f *ProviderFactory) UpdateProviderConfig(providerType core.ProviderType, updates *core.ProviderConfig) error {
	// Update provider configuration
	if err := f.configManager.UpdateConfig(providerType, updates); err != nil {
		return fmt.Errorf("failed to update provider configuration: %w", err)
	}

	// Refresh provider instance
	if err := f.RefreshProvider(providerType); err != nil {
		return fmt.Errorf("failed to refresh provider instance: %w", err)
	}

	return nil

// SaveConfigurations saves all configurations
func (f *ProviderFactory) SaveConfigurations() error {
	if err := f.configManager.Save(); err != nil {
		return fmt.Errorf("failed to save configurations: %w", err)
	}

