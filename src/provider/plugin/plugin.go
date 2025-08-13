// Package plugin provides functionality for dynamically loading provider plugins.
package plugin

import (
	"fmt"
	"plugin"
	"sync"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/provider/factory"
)

// ProviderPlugin represents a provider plugin
type ProviderPlugin struct {
	// Name is the name of the plugin
	Name string
	// Path is the path to the plugin
	Path string
	// ProviderType is the type of provider
	ProviderType core.ProviderType
	// Constructor is the provider constructor (legacy)
	Constructor factory.ProviderConstructor
	// Plugin is the plugin instance
	Plugin *plugin.Plugin
	// Metadata is the plugin metadata
	Metadata *PluginMetadata
	// PluginInterface is the plugin interface
	PluginInterface PluginInterface
	// IsLegacy indicates if this is a legacy plugin
	IsLegacy bool
}

// PluginManager is responsible for managing provider plugins
type PluginManager struct {
	// plugins is a map of plugin names to plugins
	plugins map[string]*ProviderPlugin
	// providerFactory is the provider factory
	providerFactory *factory.ProviderFactory
	// pluginDirs is a list of directories to search for plugins
	pluginDirs []string
	// mutex is a mutex for concurrent access to plugins
	mutex sync.RWMutex
	// discovery is the plugin discovery
	discovery *PluginDiscovery
	// validator is the plugin validator
	validator PluginValidator
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(providerFactory *factory.ProviderFactory, pluginDirs []string) *PluginManager {
	// Create plugin discovery
	discovery := NewPluginDiscovery(pluginDirs)
	
	// Create plugin validator
	validator := NewDefaultPluginValidator()
	
	// Add validator to discovery
	discovery.AddValidator(validator)
	
	return &PluginManager{
		plugins:        make(map[string]*ProviderPlugin),
		providerFactory: providerFactory,
		pluginDirs:     pluginDirs,
		discovery:      discovery,
		validator:      validator,
	}
}

// LoadPlugin loads a plugin from a file
func (m *PluginManager) LoadPlugin(pluginPath string) (*ProviderPlugin, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if plugin is already loaded
	pluginName := filepath.Base(pluginPath)
	if _, ok := m.plugins[pluginName]; ok {
		return nil, fmt.Errorf("plugin %s already loaded", pluginName)
	}

	// Load plugin using discovery
	providerPlugin, err := m.discovery.loadPlugin(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	// Register provider constructor
	if providerPlugin.IsLegacy {
		// Legacy plugin
		m.providerFactory.RegisterProvider(providerPlugin.ProviderType, providerPlugin.Constructor)
	} else {
		// Modern plugin
		m.providerFactory.RegisterProvider(providerPlugin.ProviderType, func(config *core.ProviderConfig) (core.Provider, error) {
			return providerPlugin.PluginInterface.CreateProvider(config)
		})
	}

	// Store plugin
	m.plugins[pluginName] = providerPlugin

	return providerPlugin, nil
}

// LoadPluginsFromDirs loads plugins from directories
func (m *PluginManager) LoadPluginsFromDirs() ([]string, []error) {
	// Discover plugins
	plugins, errors := m.discovery.DiscoverPlugins()
	
	var loadedPlugins []string
	
	// Register discovered plugins
	for _, plugin := range plugins {
		// Check if plugin is already loaded
		if _, ok := m.plugins[plugin.Name]; ok {
			continue
		}
		
		// Register provider constructor
		if plugin.IsLegacy {
			// Legacy plugin
			m.providerFactory.RegisterProvider(plugin.ProviderType, plugin.Constructor)
		} else {
			// Modern plugin
			m.providerFactory.RegisterProvider(plugin.ProviderType, func(config *core.ProviderConfig) (core.Provider, error) {
				return plugin.PluginInterface.CreateProvider(config)
			})
		}
		
		// Store plugin
		m.mutex.Lock()
		m.plugins[plugin.Name] = plugin
		m.mutex.Unlock()
		
		loadedPlugins = append(loadedPlugins, plugin.Name)
	}
	
	return loadedPlugins, errors
}

// GetPlugin returns a plugin by name
func (m *PluginManager) GetPlugin(name string) (*ProviderPlugin, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plugin, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// GetPluginByProviderType returns a plugin by provider type
func (m *PluginManager) GetPluginByProviderType(providerType core.ProviderType) (*ProviderPlugin, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, plugin := range m.plugins {
		if plugin.ProviderType == providerType {
			return plugin, nil
		}
	}

	return nil, fmt.Errorf("plugin for provider type %s not found", providerType)
}

// GetAllPlugins returns all plugins
func (m *PluginManager) GetAllPlugins() []*ProviderPlugin {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plugins := make([]*ProviderPlugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// UnloadPlugin unloads a plugin
func (m *PluginManager) UnloadPlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, ok := m.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Remove plugin
	delete(m.plugins, name)

	return nil
}

// AddPluginDir adds a plugin directory
func (m *PluginManager) AddPluginDir(dir string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.pluginDirs = append(m.pluginDirs, dir)
	m.discovery.AddPluginDir(dir)
}

// GetPluginDirs returns the plugin directories
func (m *PluginManager) GetPluginDirs() []string {
	return m.discovery.GetPluginDirs()
}

// ValidatePlugin validates a plugin
func (m *PluginManager) ValidatePlugin(plugin *ProviderPlugin) error {
	return m.validator.ValidatePlugin(plugin)
}

// ValidatePluginConfig validates a plugin configuration
func (m *PluginManager) ValidatePluginConfig(plugin *ProviderPlugin, config *core.ProviderConfig) error {
	if plugin == nil {
		return fmt.Errorf("plugin is nil")
	}
	
	return plugin.PluginInterface.ValidateConfig(config)
}
