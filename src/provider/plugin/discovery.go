// Package plugin provides functionality for dynamically loading provider plugins.
package plugin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"plugin"
	"sync"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// PluginDiscovery is responsible for discovering provider plugins
type PluginDiscovery struct {
	// pluginDirs is a list of directories to search for plugins
	pluginDirs []string
	// validators is a list of plugin validators
	validators []PluginValidator
	// mutex is a mutex for concurrent access
	mutex sync.RWMutex

// NewPluginDiscovery creates a new plugin discovery
func NewPluginDiscovery(pluginDirs []string) *PluginDiscovery {
	return &PluginDiscovery{
		pluginDirs: pluginDirs,
		validators: make([]PluginValidator, 0),
	}

// AddValidator adds a plugin validator
func (d *PluginDiscovery) AddValidator(validator PluginValidator) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	d.validators = append(d.validators, validator)

// AddPluginDir adds a plugin directory
func (d *PluginDiscovery) AddPluginDir(dir string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	// Check if directory already exists in the list
	for _, existingDir := range d.pluginDirs {
		if existingDir == dir {
			return
		}
	}
	
	d.pluginDirs = append(d.pluginDirs, dir)

// GetPluginDirs returns the plugin directories
func (d *PluginDiscovery) GetPluginDirs() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	dirs := make([]string, len(d.pluginDirs))
	copy(dirs, d.pluginDirs)
	return dirs

// DiscoverPlugins discovers plugins in the plugin directories
func (d *PluginDiscovery) DiscoverPlugins() ([]*ProviderPlugin, []error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	var plugins []*ProviderPlugin
	var errors []error
	
	// Process each plugin directory
	for _, dir := range d.pluginDirs {
		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("plugin directory does not exist: %s", dir))
			continue
		}
		
		// Find plugin files
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to read plugin directory: %s: %w", dir, err))
			continue
		}
		
		// Process each file
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			
			// Check if file is a plugin
			if !d.isPluginFile(file.Name()) {
				continue
			}
			
			// Load plugin
			pluginPath := filepath.Join(dir, file.Name())
			plugin, err := d.loadPlugin(pluginPath)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to load plugin: %s: %w", pluginPath, err))
				continue
			}
			
			// Validate plugin
			for _, validator := range d.validators {
				if err := validator.ValidatePlugin(plugin); err != nil {
					errors = append(errors, fmt.Errorf("plugin validation failed: %s: %w", pluginPath, err))
					continue
				}
			}
			
			plugins = append(plugins, plugin)
		}
	}
	
	return plugins, errors

// isPluginFile checks if a file is a plugin file
func (d *PluginDiscovery) isPluginFile(fileName string) bool {
	// Check file extension
	ext := filepath.Ext(fileName)
	return ext == ".so" || ext == ".dll" || ext == ".dylib"

// loadPlugin loads a plugin from a file
func (d *PluginDiscovery) loadPlugin(pluginPath string) (*ProviderPlugin, error) {
	// Check if plugin has a metadata file
	metadataPath := pluginPath + ".metadata.json"
	metadata, err := d.loadPluginMetadata(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin metadata: %w", err)
	}
	
	// Load plugin
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}
	
	// Get plugin interface
	pluginInterfaceSymbol, err := plug.Lookup("PluginInterface")
	if err != nil {
		// Try legacy plugin format
		return d.loadLegacyPlugin(pluginPath, plug)
	}
	
	pluginInterface, ok := pluginInterfaceSymbol.(PluginInterface)
	if !ok {
		return nil, fmt.Errorf("PluginInterface is not of type plugin.PluginInterface")
	}
	
	// Get plugin metadata
	if metadata == nil {
		metadata = pluginInterface.GetMetadata()
	}
	
	// Validate plugin compatibility
	for _, validator := range d.validators {
		if err := validator.ValidateCompatibility(metadata); err != nil {
			return nil, fmt.Errorf("plugin compatibility validation failed: %w", err)
		}
	}
	
	// Create provider plugin
	providerPlugin := &ProviderPlugin{
		Name:           metadata.Name,
		Path:           pluginPath,
		ProviderType:   metadata.ProviderType,
		Metadata:       metadata,
		PluginInterface: pluginInterface,
		Plugin:         plug,
	}
	
	return providerPlugin, nil
// loadLegacyPlugin loads a legacy plugin
func (d *PluginDiscovery) loadLegacyPlugin(pluginPath string, plug *plugin.Plugin) (*ProviderPlugin, error) {
	// Get provider type
	providerTypeSymbol, err := plug.Lookup("ProviderType")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup ProviderType: %w", err)
	}
	
	providerType, ok := providerTypeSymbol.(*core.ProviderType)
	if !ok {
		return nil, fmt.Errorf("ProviderType is not of type *core.ProviderType")
	}
	
	// Get provider constructor
	constructorSymbol, err := plug.Lookup("NewProvider")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup NewProvider: %w", err)
	}
	
	constructor, ok := constructorSymbol.(func(*core.ProviderConfig) (core.Provider, error))
	if !ok {
		return nil, fmt.Errorf("NewProvider is not of type func(*core.ProviderConfig) (core.Provider, error)")
	}
	
	// Create legacy plugin adapter
	legacyAdapter := &LegacyPluginAdapter{
		providerType: *providerType,
		constructor:  constructor,
	}
	
	// Create metadata
	metadata := &PluginMetadata{
		Name:         filepath.Base(pluginPath),
		ProviderType: *providerType,
		Version:      "1.0.0", // Default version for legacy plugins
	}
	
	// Create provider plugin
	providerPlugin := &ProviderPlugin{
		Name:           metadata.Name,
		Path:           pluginPath,
		ProviderType:   *providerType,
		Metadata:       metadata,
		PluginInterface: legacyAdapter,
		Plugin:         plug,
		IsLegacy:       true,
	}
	
	return providerPlugin, nil

// loadPluginMetadata loads plugin metadata from a file
func (d *PluginDiscovery) loadPluginMetadata(metadataPath string) (*PluginMetadata, error) {
	// Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, nil
	}
	
	// Read metadata file
	data, err := ioutil.ReadFile(filepath.Clean(metadataPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}
	
	// Parse metadata
	var metadata PluginMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	
	return &metadata, nil

// LegacyPluginAdapter adapts legacy plugins to the new plugin interface
type LegacyPluginAdapter struct {
	providerType core.ProviderType
	constructor  func(*core.ProviderConfig) (core.Provider, error)

// GetMetadata returns metadata about the plugin
func (a *LegacyPluginAdapter) GetMetadata() *PluginMetadata {
	return &PluginMetadata{
		Name:         string(a.providerType),
		ProviderType: a.providerType,
		Version:      "1.0.0", // Default version for legacy plugins
	}

// CreateProvider creates a new provider instance
func (a *LegacyPluginAdapter) CreateProvider(config *core.ProviderConfig) (core.Provider, error) {
	return a.constructor(config)

// ValidateConfig validates the provider configuration
func (a *LegacyPluginAdapter) ValidateConfig(config *core.ProviderConfig) error {
	// Legacy plugins don't have explicit validation
