// Package cache provides caching functionality for templates.
package cache

import (
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// TemplateCacheAdapter adapts the TemplateCache to implement the interfaces.TemplateCache interface
type TemplateCacheAdapter struct {
	cache *TemplateCache
}

// NewTemplateCacheAdapter creates a new template cache adapter
func NewTemplateCacheAdapter(cache *TemplateCache) *TemplateCacheAdapter {
	return &TemplateCacheAdapter{
		cache: cache,
	}
}

// Get gets a template from the cache (interface compatibility)
func (a *TemplateCacheAdapter) Get(id string) (interfaces.Template, bool) {
	template, found := a.cache.Get(id)
	if !found {
		return nil, false
	}
	return template, true
}

// Set sets a template in the cache (interface compatibility)
func (a *TemplateCacheAdapter) Set(key string, template interfaces.Template, ttl time.Duration) {
	// Convert interfaces.Template to *format.Template
	formatTemplate, ok := template.(*format.Template)
	if !ok {
		return // Skip if not the right type
	}

	if ttl > 0 {
		a.cache.SetWithTTL(key, formatTemplate, ttl)
	} else {
		a.cache.Set(key, formatTemplate)
	}
}

// SetTemplate sets a template in the cache (backward compatibility)
func (a *TemplateCacheAdapter) SetTemplate(id string, template *format.Template) {
	a.cache.Set(id, template)
}

// SetWithTTL sets a template in the cache with a specific TTL
func (a *TemplateCacheAdapter) SetWithTTL(id string, template *format.Template, ttl time.Duration) {
	a.cache.SetWithTTL(id, template, ttl)
}

// Delete deletes a template from the cache
func (a *TemplateCacheAdapter) Delete(id string) {
	a.cache.Delete(id)
}

// Clear clears the cache
func (a *TemplateCacheAdapter) Clear() {
	a.cache.Clear()
}

// Size returns the number of templates in the cache
func (a *TemplateCacheAdapter) Size() int {
	return a.cache.Size()
}

// Prune removes entries from the cache that are older than the specified duration
func (a *TemplateCacheAdapter) Prune(maxAge time.Duration) int {
	return a.cache.Prune(maxAge)
}

// GetStats gets cache statistics
func (a *TemplateCacheAdapter) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["size"] = a.cache.Size()
	stats["maxSize"] = a.cache.maxSize
	stats["evictionPolicy"] = string(a.cache.evictionPolicy)
	stats["defaultTTL"] = a.cache.defaultTTL.String()
	return stats
}
