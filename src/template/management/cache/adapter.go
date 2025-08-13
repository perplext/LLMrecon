// Package cache provides caching functionality for templates.
package cache

import (
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
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

// Get gets a template from the cache
func (a *TemplateCacheAdapter) Get(id string) (*format.Template, bool) {
	return a.cache.Get(id)
}

// Set sets a template in the cache
func (a *TemplateCacheAdapter) Set(id string, template *format.Template) {
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
