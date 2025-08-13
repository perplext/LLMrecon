// Package core provides the core interfaces and types for the Multi-Provider LLM Integration Framework.
package core

import (
	"fmt"
	"sync"
)

// DefaultProviderFactory is the default implementation of the ProviderFactory interface
type DefaultProviderFactory struct {
	// providerCreators is a map of provider types to creator functions
	providerCreators map[ProviderType]ProviderCreator
	// mutex is a mutex for concurrent access to providerCreators
	mutex sync.RWMutex
}

// ProviderCreator is a function that creates a provider with the given configuration
type ProviderCreator func(config *ProviderConfig) (Provider, error)

// NewDefaultProviderFactory creates a new default provider factory
func NewDefaultProviderFactory() *DefaultProviderFactory {
	return &DefaultProviderFactory{
		providerCreators: make(map[ProviderType]ProviderCreator),
	}
}

// RegisterProviderCreator registers a provider creator function for a provider type
func (f *DefaultProviderFactory) RegisterProviderCreator(providerType ProviderType, creator ProviderCreator) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.providerCreators[providerType] = creator
}

// CreateProvider creates a provider with the given configuration
func (f *DefaultProviderFactory) CreateProvider(config *ProviderConfig) (Provider, error) {
	f.mutex.RLock()
	creator, ok := f.providerCreators[config.Type]
	f.mutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}

	return creator(config)
}

// GetSupportedProviderTypes returns the provider types supported by this factory
func (f *DefaultProviderFactory) GetSupportedProviderTypes() []ProviderType {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	types := make([]ProviderType, 0, len(f.providerCreators))
	for providerType := range f.providerCreators {
		types = append(types, providerType)
	}

	return types
}

// DefaultProviderRegistry is the default implementation of the ProviderRegistry interface
type DefaultProviderRegistry struct {
	// providers is a map of provider types to providers
	providers map[ProviderType]Provider
	// providerFactories is a list of provider factories
	providerFactories []ProviderFactory
	// modelProviderMap is a map of model IDs to provider types
	modelProviderMap map[string]ProviderType
	// capabilityProviderMap is a map of capabilities to provider types
	capabilityProviderMap map[ModelCapability][]ProviderType
	// mutex is a mutex for concurrent access to the registry
	mutex sync.RWMutex
}

// NewDefaultProviderRegistry creates a new default provider registry
func NewDefaultProviderRegistry() *DefaultProviderRegistry {
	return &DefaultProviderRegistry{
		providers:            make(map[ProviderType]Provider),
		providerFactories:    make([]ProviderFactory, 0),
		modelProviderMap:     make(map[string]ProviderType),
		capabilityProviderMap: make(map[ModelCapability][]ProviderType),
	}
}

// RegisterProvider registers a provider
func (r *DefaultProviderRegistry) RegisterProvider(provider Provider) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	providerType := provider.GetType()
	if _, exists := r.providers[providerType]; exists {
		return fmt.Errorf("provider with type %s already registered", providerType)
	}

	r.providers[providerType] = provider

	// Update model provider map
	models, err := provider.GetModels(nil)
	if err == nil {
		for _, model := range models {
			r.modelProviderMap[model.ID] = providerType

			// Update capability provider map
			for _, capability := range model.Capabilities {
				if _, exists := r.capabilityProviderMap[capability]; !exists {
					r.capabilityProviderMap[capability] = make([]ProviderType, 0)
				}
				r.capabilityProviderMap[capability] = append(r.capabilityProviderMap[capability], providerType)
			}
		}
	}

	return nil
}

// RegisterProviderFactory registers a provider factory
func (r *DefaultProviderRegistry) RegisterProviderFactory(factory ProviderFactory) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.providerFactories = append(r.providerFactories, factory)
	return nil
}

// GetProvider returns a provider by type
func (r *DefaultProviderRegistry) GetProvider(providerType ProviderType) (Provider, error) {
	r.mutex.RLock()
	provider, exists := r.providers[providerType]
	r.mutex.RUnlock()

	if exists {
		return provider, nil
	}

	// Try to create provider using factories
	for _, factory := range r.providerFactories {
		for _, supportedType := range factory.GetSupportedProviderTypes() {
			if supportedType == providerType {
				// Create provider with default configuration
				provider, err := factory.CreateProvider(&ProviderConfig{
					Type: providerType,
				})
				if err != nil {
					return nil, err
				}

				// Register provider
				err = r.RegisterProvider(provider)
				if err != nil {
					return nil, err
				}

				return provider, nil
			}
		}
	}

	return nil, fmt.Errorf("provider with type %s not found", providerType)
}

// GetProviderByModel returns a provider that supports a specific model
func (r *DefaultProviderRegistry) GetProviderByModel(modelID string) (Provider, error) {
	r.mutex.RLock()
	providerType, exists := r.modelProviderMap[modelID]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no provider found for model %s", modelID)
	}

	return r.GetProvider(providerType)
}

// GetProviderByCapability returns a provider that supports a specific capability
func (r *DefaultProviderRegistry) GetProviderByCapability(capability ModelCapability) (Provider, error) {
	r.mutex.RLock()
	providerTypes, exists := r.capabilityProviderMap[capability]
	r.mutex.RUnlock()

	if !exists || len(providerTypes) == 0 {
		return nil, fmt.Errorf("no provider found for capability %s", capability)
	}

	// Return the first provider that supports the capability
	return r.GetProvider(providerTypes[0])
}

// GetAllProviders returns all registered providers
func (r *DefaultProviderRegistry) GetAllProviders() []Provider {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers
}

// GetAllProviderTypes returns all registered provider types
func (r *DefaultProviderRegistry) GetAllProviderTypes() []ProviderType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	types := make([]ProviderType, 0, len(r.providers))
	for providerType := range r.providers {
		types = append(types, providerType)
	}

	return types
}

// DefaultModelRegistry is the default implementation of the ModelRegistry interface
type DefaultModelRegistry struct {
	// models is a map of model IDs to models
	models map[string]*ModelInfo
	// providerModels is a map of provider types to model IDs
	providerModels map[ProviderType][]string
	// typeModels is a map of model types to model IDs
	typeModels map[ModelType][]string
	// capabilityModels is a map of capabilities to model IDs
	capabilityModels map[ModelCapability][]string
	// mutex is a mutex for concurrent access to the registry
	mutex sync.RWMutex
}

// NewDefaultModelRegistry creates a new default model registry
func NewDefaultModelRegistry() *DefaultModelRegistry {
	return &DefaultModelRegistry{
		models:          make(map[string]*ModelInfo),
		providerModels:  make(map[ProviderType][]string),
		typeModels:      make(map[ModelType][]string),
		capabilityModels: make(map[ModelCapability][]string),
	}
}

// RegisterModel registers a model
func (r *DefaultModelRegistry) RegisterModel(model *ModelInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.models[model.ID]; exists {
		return fmt.Errorf("model with ID %s already registered", model.ID)
	}

	r.models[model.ID] = model

	// Update provider models map
	if _, exists := r.providerModels[model.Provider]; !exists {
		r.providerModels[model.Provider] = make([]string, 0)
	}
	r.providerModels[model.Provider] = append(r.providerModels[model.Provider], model.ID)

	// Update type models map
	if _, exists := r.typeModels[model.Type]; !exists {
		r.typeModels[model.Type] = make([]string, 0)
	}
	r.typeModels[model.Type] = append(r.typeModels[model.Type], model.ID)

	// Update capability models map
	for _, capability := range model.Capabilities {
		if _, exists := r.capabilityModels[capability]; !exists {
			r.capabilityModels[capability] = make([]string, 0)
		}
		r.capabilityModels[capability] = append(r.capabilityModels[capability], model.ID)
	}

	return nil
}

// GetModel returns a model by ID
func (r *DefaultModelRegistry) GetModel(modelID string) (*ModelInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	model, exists := r.models[modelID]
	if !exists {
		return nil, fmt.Errorf("model with ID %s not found", modelID)
	}

	return model, nil
}

// GetModelsByProvider returns models by provider
func (r *DefaultModelRegistry) GetModelsByProvider(providerType ProviderType) ([]*ModelInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	modelIDs, exists := r.providerModels[providerType]
	if !exists {
		return nil, fmt.Errorf("no models found for provider %s", providerType)
	}

	models := make([]*ModelInfo, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		models = append(models, r.models[modelID])
	}

	return models, nil
}

// GetModelsByType returns models by type
func (r *DefaultModelRegistry) GetModelsByType(modelType ModelType) ([]*ModelInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	modelIDs, exists := r.typeModels[modelType]
	if !exists {
		return nil, fmt.Errorf("no models found for type %s", modelType)
	}

	models := make([]*ModelInfo, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		models = append(models, r.models[modelID])
	}

	return models, nil
}

// GetModelsByCapability returns models by capability
func (r *DefaultModelRegistry) GetModelsByCapability(capability ModelCapability) ([]*ModelInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	modelIDs, exists := r.capabilityModels[capability]
	if !exists {
		return nil, fmt.Errorf("no models found for capability %s", capability)
	}

	models := make([]*ModelInfo, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		models = append(models, r.models[modelID])
	}

	return models, nil
}

// GetAllModels returns all registered models
func (r *DefaultModelRegistry) GetAllModels() []*ModelInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	models := make([]*ModelInfo, 0, len(r.models))
	for _, model := range r.models {
		models = append(models, model)
	}

	return models
}
