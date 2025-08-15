// Package registry provides functionality for registering and retrieving providers.
package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// ProviderRegistry is responsible for registering and retrieving providers
type ProviderRegistry struct {
	// providers is a map of provider types to providers
	providers map[core.ProviderType]core.Provider
	// providerFactories is a map of provider types to provider factories
	providerFactories map[core.ProviderType]core.ProviderFactory
	// mutex is a mutex for concurrent access to providers
	mutex sync.RWMutex

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers:        make(map[core.ProviderType]core.Provider),
		providerFactories: make(map[core.ProviderType]core.ProviderFactory),
	}

// RegisterProvider registers a provider
func (r *ProviderRegistry) RegisterProvider(provider core.Provider) error {
	if provider == nil {
		return fmt.Errorf("provider is nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.providers[provider.GetType()] = provider

	return nil

// RegisterProviderFactory registers a provider factory
func (r *ProviderRegistry) RegisterProviderFactory(factory core.ProviderFactory) error {
	if factory == nil {
		return fmt.Errorf("factory is nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, providerType := range factory.GetSupportedProviderTypes() {
		r.providerFactories[providerType] = factory
	}

	return nil

// GetProvider returns a provider by type
func (r *ProviderRegistry) GetProvider(providerType core.ProviderType) (core.Provider, error) {
	r.mutex.RLock()
	provider, ok := r.providers[providerType]
	r.mutex.RUnlock()

	if ok {
		return provider, nil
	}

	// Try to create provider using factory
	r.mutex.Lock()
	defer r.mutex.Unlock()

	factory, ok := r.providerFactories[providerType]
	if !ok {
		return nil, fmt.Errorf("provider factory for type %s not found", providerType)
	}

	// Create provider
	config := &core.ProviderConfig{
		Type: providerType,
	}
	provider, err := factory.CreateProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Cache provider
	r.providers[providerType] = provider

	return provider, nil

// GetProviderByModel returns a provider that supports a specific model
func (r *ProviderRegistry) GetProviderByModel(modelID string) (core.Provider, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, provider := range r.providers {
		if provider.SupportsModel(context.Background(), modelID) {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("provider for model %s not found", modelID)

// GetProviderByCapability returns a provider that supports a specific capability
func (r *ProviderRegistry) GetProviderByCapability(capability core.ModelCapability) (core.Provider, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, provider := range r.providers {
		if provider.SupportsCapability(context.Background(), capability) {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("provider for capability %s not found", capability)

// GetAllProviders returns all registered providers
func (r *ProviderRegistry) GetAllProviders() []core.Provider {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	providers := make([]core.Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers

// GetAllProviderTypes returns all registered provider types
func (r *ProviderRegistry) GetAllProviderTypes() []core.ProviderType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	providerTypes := make([]core.ProviderType, 0, len(r.providers))
	for providerType := range r.providers {
		providerTypes = append(providerTypes, providerType)
	}

	return providerTypes

// CloseProvider closes a provider
func (r *ProviderRegistry) CloseProvider(providerType core.ProviderType) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	provider, ok := r.providers[providerType]
	if !ok {
		return fmt.Errorf("provider for type %s not found", providerType)
	}

	if err := provider.Close(); err != nil {
		return fmt.Errorf("failed to close provider: %w", err)
	}

	delete(r.providers, providerType)

	return nil

// CloseAllProviders closes all providers
func (r *ProviderRegistry) CloseAllProviders() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var errs []error

	for providerType, provider := range r.providers {
		if err := provider.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close provider %s: %w", providerType, err))
		}
	}

	r.providers = make(map[core.ProviderType]core.Provider)

	if len(errs) > 0 {
		return fmt.Errorf("failed to close all providers: %v", errs)
	}

	return nil

// ModelRegistry is responsible for registering and retrieving models
type ModelRegistry struct {
	// models is a map of model IDs to models
	models map[string]*core.ModelInfo
	// mutex is a mutex for concurrent access to models
	mutex sync.RWMutex

// NewModelRegistry creates a new model registry
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models: make(map[string]*core.ModelInfo),
	}

// RegisterModel registers a model
func (r *ModelRegistry) RegisterModel(model *core.ModelInfo) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.models[model.ID] = model

	return nil

// GetModel returns a model by ID
func (r *ModelRegistry) GetModel(modelID string) (*core.ModelInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	model, ok := r.models[modelID]
	if !ok {
		return nil, fmt.Errorf("model with ID %s not found", modelID)
	}

	// Return a copy of the model to prevent modification
	modelCopy := *model
	return &modelCopy, nil

// GetModelsByProvider returns models by provider
func (r *ModelRegistry) GetModelsByProvider(providerType core.ProviderType) ([]*core.ModelInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	models := make([]*core.ModelInfo, 0)
	for _, model := range r.models {
		if model.Provider == providerType {
			// Return a copy of the model to prevent modification
			modelCopy := *model
			models = append(models, &modelCopy)
		}
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no models found for provider %s", providerType)
	}

	return models, nil

// GetModelsByType returns models by type
func (r *ModelRegistry) GetModelsByType(modelType core.ModelType) ([]*core.ModelInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	models := make([]*core.ModelInfo, 0)
	for _, model := range r.models {
		if model.Type == modelType {
			// Return a copy of the model to prevent modification
			modelCopy := *model
			models = append(models, &modelCopy)
		}
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no models found for type %s", modelType)
	}

	return models, nil

// GetModelsByCapability returns models by capability
func (r *ModelRegistry) GetModelsByCapability(capability core.ModelCapability) ([]*core.ModelInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	models := make([]*core.ModelInfo, 0)
	for _, model := range r.models {
		for _, modelCapability := range model.Capabilities {
			if modelCapability == capability {
				// Return a copy of the model to prevent modification
				modelCopy := *model
				models = append(models, &modelCopy)
				break
			}
		}
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no models found for capability %s", capability)
	}

	return models, nil

// GetAllModels returns all registered models
func (r *ModelRegistry) GetAllModels() []*core.ModelInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	models := make([]*core.ModelInfo, 0, len(r.models))
	for _, model := range r.models {
		// Return a copy of the model to prevent modification
		modelCopy := *model
		models = append(models, &modelCopy)
	}

	return models

// DeleteModel deletes a model
func (r *ModelRegistry) DeleteModel(modelID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.models[modelID]; !ok {
		return fmt.Errorf("model with ID %s not found", modelID)
	}

	delete(r.models, modelID)

	return nil

// UpdateModel updates a model
func (r *ModelRegistry) UpdateModel(model *core.ModelInfo) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.models[model.ID]; !ok {
		return fmt.Errorf("model with ID %s not found", model.ID)
	}

	r.models[model.ID] = model

	return nil

// SyncModelsFromProviders syncs models from providers
func (r *ModelRegistry) SyncModelsFromProviders(providers []core.Provider) error {
	if len(providers) == 0 {
		return fmt.Errorf("no providers specified")
	}

	// Get models from providers
	allModels := make([]*core.ModelInfo, 0)
	for _, provider := range providers {
		models, err := provider.GetModels(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get models from provider %s: %w", provider.GetType(), err)
		}

		for i := range models {
			allModels = append(allModels, &models[i])
		}
	}

	// Update registry
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Clear existing models
	r.models = make(map[string]*core.ModelInfo)

	// Add new models
	for _, model := range allModels {
		r.models[model.ID] = model
	}

