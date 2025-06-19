# Provider Plugin System

The Multi-Provider LLM Integration Framework includes a plugin system that allows for easy extension with new providers. This document explains how to create and use provider plugins.

## Overview

The plugin system uses Go's plugin package to dynamically load provider implementations at runtime. This allows for adding new providers without modifying the core framework code.

Key components of the plugin system:

- `PluginManager`: Manages loading and unloading of plugins
- `ProviderPlugin`: Represents a loaded plugin
- Plugin template: A template for implementing new provider plugins

## Creating a Provider Plugin

To create a new provider plugin:

1. Copy the template from `examples/provider/plugin_template/template.go`
2. Rename the file to match your provider (e.g., `myprovider.go`)
3. Update the `ProviderType` variable to match your provider type
4. Implement the required methods for your provider
5. Build the plugin as a shared library

### Required Exports

Each plugin must export the following:

- `ProviderType`: A variable of type `core.ProviderType` that identifies the provider type
- `NewProvider`: A function that creates a new provider instance

### Building a Plugin

To build a plugin as a shared library:

```bash
go build -buildmode=plugin -o myprovider.so myprovider.go
```

### Example Plugin Implementation

```go
package main

import (
    "github.com/perplext/LLMrecon/src/provider/core"
)

// ProviderType is the type of provider
var ProviderType = core.ProviderType("my-provider")

// NewProvider creates a new provider instance
func NewProvider(config *core.ProviderConfig) (core.Provider, error) {
    // Implementation here
}

// Other methods...
```

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

### Loading a Specific Plugin

To load a specific plugin:

```go
plugin, err := pluginManager.LoadPlugin("/path/to/plugins/myprovider.so")
if err != nil {
    // Handle error
}
```

### Getting a Plugin

To get a loaded plugin:

```go
// By name
plugin, err := pluginManager.GetPlugin("myprovider.so")

// By provider type
plugin, err := pluginManager.GetPluginByProviderType(core.ProviderType("my-provider"))
```

### Unloading a Plugin

To unload a plugin:

```go
err := pluginManager.UnloadPlugin("myprovider.so")
if err != nil {
    // Handle error
}
```

## Plugin Directory Structure

The recommended directory structure for plugins:

```
/path/to/plugins/
├── provider1.so
├── provider2.so
└── ...
```

## Best Practices

1. **Error Handling**: Implement robust error handling in your plugin
2. **Configuration**: Support all configuration options from the core framework
3. **Logging**: Use the logging middleware for consistent logging
4. **Rate Limiting**: Implement appropriate rate limiting for your provider
5. **Testing**: Test your plugin thoroughly before deployment

## Limitations

- Plugins must be compiled with the same Go version as the main application
- Plugins must be compiled for the same operating system and architecture
- Plugins cannot be unloaded once loaded (Go limitation)

## Example: Complete Plugin Implementation

See the `examples/provider/plugin_template/template.go` file for a complete example of a provider plugin implementation.

## Building and Installing Plugins

1. Build your plugin:
   ```bash
   go build -buildmode=plugin -o myprovider.so myprovider.go
   ```

2. Copy the plugin to the plugins directory:
   ```bash
   cp myprovider.so /path/to/plugins/
   ```

3. Configure the application to use the plugins directory:
   ```go
   pluginManager := plugin.NewPluginManager(providerFactory, []string{"/path/to/plugins"})
   ```

4. Load the plugins:
   ```go
   loadedPlugins, errors := pluginManager.LoadPluginsFromDirs()
   ```

## Troubleshooting

Common issues and solutions:

- **Plugin not loading**: Ensure the plugin is compiled with the same Go version and for the same OS/architecture
- **Symbol not found**: Ensure the plugin exports the required symbols (`ProviderType` and `NewProvider`)
- **Version mismatch**: Ensure the plugin is compiled against the same version of the framework

## Contributing New Providers

If you've developed a provider plugin that might be useful to others, consider contributing it to the main repository:

1. Fork the repository
2. Add your provider implementation
3. Submit a pull request

Please include documentation and tests for your provider.
