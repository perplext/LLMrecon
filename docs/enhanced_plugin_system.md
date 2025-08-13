# Enhanced Provider Plugin System

This document describes the enhanced plugin system for the LLMrecon framework, which allows for easy extension with new providers.

## Overview

The enhanced plugin system builds on the existing plugin architecture to provide more robust features:

- **Metadata Support**: Plugins can now include metadata about themselves, such as version, author, and supported models
- **Validation**: The system validates plugins for compatibility with the framework
- **Discovery**: Automatic discovery of plugins in configured directories
- **Legacy Support**: Backward compatibility with existing plugins

## Plugin Interface

The new plugin system introduces a formal interface that all plugins should implement:

```go
type PluginInterface interface {
    // GetMetadata returns metadata about the plugin
    GetMetadata() *PluginMetadata
    
    // CreateProvider creates a new provider instance
    CreateProvider(config *core.ProviderConfig) (core.Provider, error)
    
    // ValidateConfig validates the provider configuration
    ValidateConfig(config *core.ProviderConfig) error
}
```

## Plugin Metadata

Plugins can now include metadata:

```go
type PluginMetadata struct {
    // Name is the name of the plugin
    Name string `json:"name"`
    // Version is the version of the plugin
    Version string `json:"version"`
    // Author is the author of the plugin
    Author string `json:"author"`
    // Description is a description of the plugin
    Description string `json:"description"`
    // ProviderType is the type of provider
    ProviderType core.ProviderType `json:"provider_type"`
    // SupportedModels is a list of models supported by the plugin
    SupportedModels []string `json:"supported_models,omitempty"`
    // MinFrameworkVersion is the minimum framework version required by the plugin
    MinFrameworkVersion string `json:"min_framework_version"`
    // MaxFrameworkVersion is the maximum framework version supported by the plugin
    MaxFrameworkVersion string `json:"max_framework_version,omitempty"`
    // Tags is a list of tags for the plugin
    Tags []string `json:"tags,omitempty"`
}
```

## Creating a Modern Provider Plugin

To create a new provider plugin using the modern interface:

1. Copy the template from `examples/provider/plugin_template/modern_template.go`
2. Rename the file to match your provider (e.g., `myprovider.go`)
3. Update the `PluginInterface` variable to implement your provider
4. Implement the required methods for your provider
5. Build the plugin as a shared library

### Required Exports

Each modern plugin must export the following:

- `PluginInterface`: A variable of type `plugin.PluginInterface` that implements the plugin interface

### Optional Metadata File

You can also provide a metadata file alongside your plugin:

```
myprovider.so.metadata.json
```

This file should contain JSON that matches the `PluginMetadata` structure.

### Building a Plugin

To build a plugin as a shared library:

```bash
go build -buildmode=plugin -o myprovider.so myprovider.go
```

### Example Plugin Implementation

See the `examples/provider/plugin_template/modern_template.go` file for a complete example of a modern provider plugin implementation.

## Legacy Plugin Support

The enhanced plugin system maintains backward compatibility with existing plugins. Legacy plugins are automatically adapted to the new interface.

Legacy plugins must export:

- `ProviderType`: A variable of type `core.ProviderType` that identifies the provider type
- `NewProvider`: A function that creates a new provider instance

## Using Provider Plugins

### Loading Plugins

To load plugins at runtime:

```go
import (
    "github.com/perplext/LLMrecon/src/provider/config"
    "github.com/perplext/LLMrecon/src/provider/factory"
    "github.com/perplext/LLMrecon/src/provider/plugin"
)

func main() {
    // Create configuration manager
    configManager, _ := config.NewConfigManager("", nil, "LLM_RED_TEAM")

    // Create provider factory
    providerFactory := factory.NewProviderFactory(configManager)

    // Create plugin manager
    pluginManager := plugin.NewPluginManager(providerFactory, []string{"/path/to/plugins"})

    // Load plugins
    loadedPlugins, errors := pluginManager.LoadPluginsFromDirs()
    if len(errors) > 0 {
        // Handle errors
    }

    // Get provider from plugin
    provider, _ := providerFactory.GetProvider(core.ProviderType("my-provider"))

    // Use provider
    // ...
}
```

### Plugin Discovery

The plugin system automatically discovers plugins in the configured directories:

```go
// Create plugin discovery
discovery := plugin.NewPluginDiscovery([]string{"/path/to/plugins"})

// Discover plugins
plugins, errors := discovery.DiscoverPlugins()
```

### Plugin Validation

The plugin system validates plugins for compatibility with the framework:

```go
// Create plugin validator
validator := plugin.NewDefaultPluginValidator()

// Validate plugin
err := validator.ValidatePlugin(plugin)
```

## Plugin Directory Structure

The recommended directory structure for plugins:

```
/path/to/plugins/
├── provider1.so
├── provider1.so.metadata.json
├── provider2.so
└── ...
```

## Best Practices

1. **Metadata**: Include comprehensive metadata with your plugin
2. **Version Compatibility**: Specify minimum and maximum framework versions
3. **Configuration Validation**: Implement robust validation in `ValidateConfig`
4. **Error Handling**: Implement robust error handling in your plugin
5. **Rate Limiting**: Implement appropriate rate limiting for your provider
6. **Testing**: Test your plugin thoroughly before deployment

## Limitations

- Plugins must be compiled with the same Go version as the main application
- Plugins must be compiled for the same operating system and architecture
- Plugins cannot be unloaded once loaded (Go limitation)

## Troubleshooting

Common issues and solutions:

- **Plugin not loading**: Ensure the plugin is compiled with the same Go version and for the same OS/architecture
- **Symbol not found**: Ensure the plugin exports the required symbols
- **Version mismatch**: Ensure the plugin is compatible with the current framework version
- **Validation failure**: Check the plugin metadata and implementation for compatibility issues

## Contributing New Providers

If you've developed a provider plugin that might be useful to others, consider contributing it to the main repository:

1. Fork the repository
2. Add your provider implementation
3. Submit a pull request

Please include documentation and tests for your provider.
