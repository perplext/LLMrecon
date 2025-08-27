package keystore

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// KeyFilter represents filters for key retrieval
type KeyFilter struct {
	Types         []KeyType         `json:"types,omitempty"`
	Algorithms    []string          `json:"algorithms,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAfter  *time.Time        `json:"created_after,omitempty"`
	CreatedBefore *time.Time        `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time        `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time        `json:"updated_before,omitempty"`
	Active        *bool             `json:"active,omitempty"`
}

// SortOptions represents sorting options for key retrieval
type SortOptions struct {
	Field     string `json:"field"`     // "id", "type", "algorithm", "created_at", "updated_at"
	Direction string `json:"direction"` // "asc", "desc"
}

// PaginationOptions represents pagination options
type PaginationOptions struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// KeyListResult represents the result of a key list operation
type KeyListResult struct {
	Keys    []*KeyEntry `json:"keys"`
	Total   int         `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
	HasMore bool        `json:"has_more"`
}

// KeySearchOptions contains options for key searching
type KeySearchOptions struct {
	Query      string             `json:"query"`
	Filter     *KeyFilter         `json:"filter,omitempty"`
	Sort       *SortOptions       `json:"sort,omitempty"`
	Pagination *PaginationOptions `json:"pagination,omitempty"`
}

// KeyRetriever handles key retrieval operations
type KeyRetriever struct {
	keystore *Keystore
}

// NewKeyRetriever creates a new key retriever
func NewKeyRetriever(ks *Keystore) *KeyRetriever {
	return &KeyRetriever{keystore: ks}
}

// GetKey retrieves a key by its ID
func (kr *KeyRetriever) GetKey(keyID string) (*KeyEntry, error) {
	if keyID == "" {
		return nil, errors.New("key ID cannot be empty")
	}

	kr.keystore.mu.RLock()
	defer kr.keystore.mu.RUnlock()

	keyEntry, exists := kr.keystore.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key with ID %s not found", keyID)
	}

	// Return a copy to prevent modification
	return kr.copyKeyEntry(keyEntry), nil
}

// GetKeys retrieves multiple keys by their IDs
func (kr *KeyRetriever) GetKeys(keyIDs []string) (map[string]*KeyEntry, error) {
	if len(keyIDs) == 0 {
		return make(map[string]*KeyEntry), nil
	}

	kr.keystore.mu.RLock()
	defer kr.keystore.mu.RUnlock()

	result := make(map[string]*KeyEntry)
	for _, keyID := range keyIDs {
		if keyEntry, exists := kr.keystore.keys[keyID]; exists {
			result[keyID] = kr.copyKeyEntry(keyEntry)
		}
	}

	return result, nil
}

// ListKeys retrieves all keys with optional filtering, sorting, and pagination
func (kr *KeyRetriever) ListKeys(options *KeySearchOptions) (*KeyListResult, error) {
	kr.keystore.mu.RLock()
	defer kr.keystore.mu.RUnlock()

	// Start with all keys
	var allKeys []*KeyEntry
	for _, keyEntry := range kr.keystore.keys {
		allKeys = append(allKeys, kr.copyKeyEntry(keyEntry))
	}

	// Apply filters
	if options != nil && options.Filter != nil {
		allKeys = kr.applyFilters(allKeys, options.Filter)
	}

	// Apply search query
	if options != nil && options.Query != "" {
		allKeys = kr.applySearch(allKeys, options.Query)
	}

	// Sort results
	if options != nil && options.Sort != nil {
		kr.sortKeys(allKeys, options.Sort)
	}

	total := len(allKeys)

	// Apply pagination
	var keys []*KeyEntry
	var hasMore bool
	if options != nil && options.Pagination != nil {
		keys, hasMore = kr.applyPagination(allKeys, options.Pagination)
	} else {
		keys = allKeys
		hasMore = false
	}

	limit := 0
	offset := 0
	if options != nil && options.Pagination != nil {
		limit = options.Pagination.Limit
		offset = options.Pagination.Offset
	}

	return &KeyListResult{
		Keys:    keys,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
}

// FindKeysByType retrieves keys by their type
func (kr *KeyRetriever) FindKeysByType(keyType KeyType) ([]*KeyEntry, error) {
	filter := &KeyFilter{
		Types: []KeyType{keyType},
	}

	options := &KeySearchOptions{
		Filter: filter,
	}

	result, err := kr.ListKeys(options)
	if err != nil {
		return nil, err
	}

	return result.Keys, nil
}

// FindKeysByAlgorithm retrieves keys by their algorithm
func (kr *KeyRetriever) FindKeysByAlgorithm(algorithm string) ([]*KeyEntry, error) {
	if algorithm == "" {
		return nil, errors.New("algorithm cannot be empty")
	}

	filter := &KeyFilter{
		Algorithms: []string{algorithm},
	}

	options := &KeySearchOptions{
		Filter: filter,
	}

	result, err := kr.ListKeys(options)
	if err != nil {
		return nil, err
	}

	return result.Keys, nil
}

// FindKeysByMetadata retrieves keys that match the specified metadata
func (kr *KeyRetriever) FindKeysByMetadata(metadata map[string]string) ([]*KeyEntry, error) {
	if len(metadata) == 0 {
		return nil, errors.New("metadata cannot be empty")
	}

	filter := &KeyFilter{
		Metadata: metadata,
	}

	options := &KeySearchOptions{
		Filter: filter,
	}

	result, err := kr.ListKeys(options)
	if err != nil {
		return nil, err
	}

	return result.Keys, nil
}

// FindKeysCreatedAfter retrieves keys created after the specified time
func (kr *KeyRetriever) FindKeysCreatedAfter(after time.Time) ([]*KeyEntry, error) {
	filter := &KeyFilter{
		CreatedAfter: &after,
	}

	options := &KeySearchOptions{
		Filter: filter,
	}

	result, err := kr.ListKeys(options)
	if err != nil {
		return nil, err
	}

	return result.Keys, nil
}

// SearchKeys performs a text search across key IDs and metadata
func (kr *KeyRetriever) SearchKeys(query string) ([]*KeyEntry, error) {
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	options := &KeySearchOptions{
		Query: query,
	}

	result, err := kr.ListKeys(options)
	if err != nil {
		return nil, err
	}

	return result.Keys, nil
}

// GetKeyCount returns the total number of keys in the keystore
func (kr *KeyRetriever) GetKeyCount() int {
	kr.keystore.mu.RLock()
	defer kr.keystore.mu.RUnlock()

	return len(kr.keystore.keys)
}

// GetKeyStatistics returns statistics about the keys in the keystore
func (kr *KeyRetriever) GetKeyStatistics() map[string]interface{} {
	kr.keystore.mu.RLock()
	defer kr.keystore.mu.RUnlock()

	stats := make(map[string]interface{})
	typeCount := make(map[string]int)
	algorithmCount := make(map[string]int)

	for _, keyEntry := range kr.keystore.keys {
		typeCount[string(keyEntry.Type)]++
		algorithmCount[keyEntry.Algorithm]++
	}

	stats["total_keys"] = len(kr.keystore.keys)
	stats["types"] = typeCount
	stats["algorithms"] = algorithmCount

	return stats
}

// copyKeyEntry creates a deep copy of a key entry
func (kr *KeyRetriever) copyKeyEntry(original *KeyEntry) *KeyEntry {
	copy := &KeyEntry{
		ID:        original.ID,
		Type:      original.Type,
		Algorithm: original.Algorithm,
		Key:       original.Key,
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
	}

	// Copy metadata
	if original.Metadata != nil {
		copy.Metadata = make(map[string]string)
		for k, v := range original.Metadata {
			copy.Metadata[k] = v
		}
	}

	return copy
}

// applyFilters applies the specified filters to the key list
func (kr *KeyRetriever) applyFilters(keys []*KeyEntry, filter *KeyFilter) []*KeyEntry {
	var filtered []*KeyEntry

	for _, key := range keys {
		if kr.matchesFilter(key, filter) {
			filtered = append(filtered, key)
		}
	}

	return filtered
}

// matchesFilter checks if a key matches the specified filter
func (kr *KeyRetriever) matchesFilter(key *KeyEntry, filter *KeyFilter) bool {
	// Check type filter
	if len(filter.Types) > 0 {
		typeMatches := false
		for _, filterType := range filter.Types {
			if key.Type == filterType {
				typeMatches = true
				break
			}
		}
		if !typeMatches {
			return false
		}
	}

	// Check algorithm filter
	if len(filter.Algorithms) > 0 {
		algorithmMatches := false
		for _, filterAlgorithm := range filter.Algorithms {
			if key.Algorithm == filterAlgorithm {
				algorithmMatches = true
				break
			}
		}
		if !algorithmMatches {
			return false
		}
	}

	// Check metadata filter
	if len(filter.Metadata) > 0 {
		for filterKey, filterValue := range filter.Metadata {
			if value, exists := key.Metadata[filterKey]; !exists || value != filterValue {
				return false
			}
		}
	}

	// Check created after filter
	if filter.CreatedAfter != nil && key.CreatedAt.Before(*filter.CreatedAfter) {
		return false
	}

	// Check created before filter
	if filter.CreatedBefore != nil && key.CreatedAt.After(*filter.CreatedBefore) {
		return false
	}

	// Check updated after filter
	if filter.UpdatedAfter != nil && key.UpdatedAt.Before(*filter.UpdatedAfter) {
		return false
	}

	// Check updated before filter
	if filter.UpdatedBefore != nil && key.UpdatedAt.After(*filter.UpdatedBefore) {
		return false
	}

	return true
}

// applySearch applies text search to the key list
func (kr *KeyRetriever) applySearch(keys []*KeyEntry, query string) []*KeyEntry {
	var results []*KeyEntry
	lowerQuery := strings.ToLower(query)

	for _, key := range keys {
		if kr.matchesQuery(key, lowerQuery) {
			results = append(results, key)
		}
	}

	return results
}

// matchesQuery checks if a key matches the search query
func (kr *KeyRetriever) matchesQuery(key *KeyEntry, query string) bool {
	// Search in key ID
	if strings.Contains(strings.ToLower(key.ID), query) {
		return true
	}

	// Search in algorithm
	if strings.Contains(strings.ToLower(key.Algorithm), query) {
		return true
	}

	// Search in metadata values
	for _, value := range key.Metadata {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}

	return false
}

// sortKeys sorts the key list based on the specified sort options
func (kr *KeyRetriever) sortKeys(keys []*KeyEntry, sortOpts *SortOptions) {
	if sortOpts == nil {
		return
	}

	sort.Slice(keys, func(i, j int) bool {
		var result bool

		switch sortOpts.Field {
		case "id":
			result = keys[i].ID < keys[j].ID
		case "type":
			result = string(keys[i].Type) < string(keys[j].Type)
		case "algorithm":
			result = keys[i].Algorithm < keys[j].Algorithm
		case "created_at":
			result = keys[i].CreatedAt.Before(keys[j].CreatedAt)
		case "updated_at":
			result = keys[i].UpdatedAt.Before(keys[j].UpdatedAt)
		default:
			// Default sort by ID
			result = keys[i].ID < keys[j].ID
		}

		if sortOpts.Direction == "desc" {
			result = !result
		}

		return result
	})
}

// applyPagination applies pagination to the key list
func (kr *KeyRetriever) applyPagination(keys []*KeyEntry, pagination *PaginationOptions) ([]*KeyEntry, bool) {
	if pagination == nil || pagination.Limit <= 0 {
		return keys, false
	}

	total := len(keys)
	start := pagination.Offset

	if start >= total {
		return []*KeyEntry{}, false
	}

	end := start + pagination.Limit
	if end > total {
		end = total
	}

	hasMore := end < total
	return keys[start:end], hasMore
}

// KeyExists checks if a key with the specified ID exists
func (kr *KeyRetriever) KeyExists(keyID string) bool {
	if keyID == "" {
		return false
	}

	kr.keystore.mu.RLock()
	defer kr.keystore.mu.RUnlock()

	_, exists := kr.keystore.keys[keyID]
	return exists
}

// GetKeyIDs returns all key IDs in the keystore
func (kr *KeyRetriever) GetKeyIDs() []string {
	kr.keystore.mu.RLock()
	defer kr.keystore.mu.RUnlock()

	var ids []string
	for id := range kr.keystore.keys {
		ids = append(ids, id)
	}

	sort.Strings(ids)
	return ids
}
