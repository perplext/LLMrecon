package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// InMemoryKnowledgeBase implements the KnowledgeBase interface
// Provides fast in-memory storage with optional persistence
type InMemoryKnowledgeBase struct {
	knowledge    map[string]*Knowledge
	typeIndex    map[KnowledgeType][]string
	tagIndex     map[string][]string
	contentIndex map[string][]string // For full-text search
	mu           sync.RWMutex
	config       *KnowledgeConfig
	persistence  PersistenceLayer
}

// KnowledgeConfig configures the knowledge base
type KnowledgeConfig struct {
	MaxItems         int
	RetentionPeriod  time.Duration
	AutoPersist      bool
	PersistInterval  time.Duration
	IndexingEnabled  bool
	FullTextSearch   bool
	CompressionLevel int
}

// PersistenceLayer handles knowledge persistence
type PersistenceLayer interface {
	Save(ctx context.Context, knowledge map[string]*Knowledge) error
	Load(ctx context.Context) (map[string]*Knowledge, error)
	Backup(ctx context.Context, filename string) error
	Restore(ctx context.Context, filename string) error
}

// FilePersistence implements file-based persistence
type FilePersistence struct {
	dataPath    string
	backupPath  string
	compression bool
}

// NewInMemoryKnowledgeBase creates a new in-memory knowledge base
func NewInMemoryKnowledgeBase(config *KnowledgeConfig) *InMemoryKnowledgeBase {
	kb := &InMemoryKnowledgeBase{
		knowledge:    make(map[string]*Knowledge),
		typeIndex:    make(map[KnowledgeType][]string),
		tagIndex:     make(map[string][]string),
		contentIndex: make(map[string][]string),
		config:       config,
	}

	// Initialize persistence if configured
	if config.AutoPersist {
		kb.persistence = &FilePersistence{
			dataPath:    "./data/knowledge.json",
			backupPath:  "./data/backups/",
			compression: config.CompressionLevel > 0,
		}
		
		// Start auto-persistence routine
		go kb.autoPersistRoutine()
	}

	// Load existing knowledge if available
	if kb.persistence != nil {
		kb.loadFromPersistence()
	}

	return kb
}

// Store saves knowledge to the knowledge base
func (kb *InMemoryKnowledgeBase) Store(ctx context.Context, knowledge *Knowledge) error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Check capacity limits
	if kb.config.MaxItems > 0 && len(kb.knowledge) >= kb.config.MaxItems {
		if err := kb.evictOldestKnowledge(); err != nil {
			return fmt.Errorf("failed to evict old knowledge: %w", err)
		}
	}

	// Store the knowledge
	kb.knowledge[knowledge.ID] = knowledge

	// Update indexes
	kb.updateIndexes(knowledge)

	return nil
}

// Retrieve finds relevant knowledge based on query
func (kb *InMemoryKnowledgeBase) Retrieve(ctx context.Context, query *KnowledgeQuery) ([]*Knowledge, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	var candidates []*Knowledge

	// Apply filters
	if query.Type != "" {
		candidates = kb.getKnowledgeByType(query.Type)
	} else {
		candidates = kb.getAllKnowledge()
	}

	// Filter by tags
	if len(query.Tags) > 0 {
		candidates = kb.filterByTags(candidates, query.Tags)
	}

	// Filter by content if specified
	if query.Content != "" {
		candidates = kb.filterByContent(candidates, query.Content)
	}

	// Filter by confidence
	if query.MinConfidence > 0 {
		candidates = kb.filterByConfidence(candidates, query.MinConfidence)
	}

	// Sort results
	kb.sortKnowledge(candidates, query.SortBy)

	// Apply result limit
	if query.MaxResults > 0 && len(candidates) > query.MaxResults {
		candidates = candidates[:query.MaxResults]
	}

	return candidates, nil
}

// Update modifies existing knowledge
func (kb *InMemoryKnowledgeBase) Update(ctx context.Context, knowledge *Knowledge) error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	existing, exists := kb.knowledge[knowledge.ID]
	if !exists {
		return fmt.Errorf("knowledge with ID %s not found", knowledge.ID)
	}

	// Remove old indexes
	kb.removeFromIndexes(existing)

	// Update the knowledge
	kb.knowledge[knowledge.ID] = knowledge

	// Update indexes
	kb.updateIndexes(knowledge)

	return nil
}

// Delete removes knowledge from the base
func (kb *InMemoryKnowledgeBase) Delete(ctx context.Context, id string) error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	knowledge, exists := kb.knowledge[id]
	if !exists {
		return fmt.Errorf("knowledge with ID %s not found", id)
	}

	// Remove from indexes
	kb.removeFromIndexes(knowledge)

	// Delete the knowledge
	delete(kb.knowledge, id)

	return nil
}

// Search performs full-text search across knowledge content
func (kb *InMemoryKnowledgeBase) Search(ctx context.Context, query string) ([]*Knowledge, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	query = strings.ToLower(query)
	var results []*Knowledge

	for _, knowledge := range kb.knowledge {
		if kb.matchesQuery(knowledge, query) {
			results = append(results, knowledge)
		}
	}

	// Sort by relevance
	kb.sortByRelevance(results, query)

	return results, nil
}

// GetStatistics returns knowledge base statistics
func (kb *InMemoryKnowledgeBase) GetStatistics() *KnowledgeStatistics {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	stats := &KnowledgeStatistics{
		TotalItems:     len(kb.knowledge),
		TypeCounts:     make(map[KnowledgeType]int),
		TagCounts:      make(map[string]int),
		ConfidenceAvg:  0.0,
		OldestItem:     time.Now(),
		NewestItem:     time.Time{},
	}

	confidenceSum := 0.0
	for _, knowledge := range kb.knowledge {
		// Count by type
		stats.TypeCounts[knowledge.Type]++

		// Count by tags
		for _, tag := range knowledge.Tags {
			stats.TagCounts[tag]++
		}

		// Calculate confidence average
		confidenceSum += knowledge.Confidence

		// Track oldest and newest
		if knowledge.Timestamp.Before(stats.OldestItem) {
			stats.OldestItem = knowledge.Timestamp
		}
		if knowledge.Timestamp.After(stats.NewestItem) {
			stats.NewestItem = knowledge.Timestamp
		}
	}

	if len(kb.knowledge) > 0 {
		stats.ConfidenceAvg = confidenceSum / float64(len(kb.knowledge))
	}

	return stats
}

// Cleanup removes expired or low-confidence knowledge
func (kb *InMemoryKnowledgeBase) Cleanup(ctx context.Context) error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	cutoffTime := time.Now().Add(-kb.config.RetentionPeriod)
	var toDelete []string

	for id, knowledge := range kb.knowledge {
		// Mark for deletion if expired
		if knowledge.Timestamp.Before(cutoffTime) {
			toDelete = append(toDelete, id)
			continue
		}

		// Mark for deletion if confidence is too low
		if knowledge.Confidence < 0.1 {
			toDelete = append(toDelete, id)
		}
	}

	// Delete marked items
	for _, id := range toDelete {
		knowledge := kb.knowledge[id]
		kb.removeFromIndexes(knowledge)
		delete(kb.knowledge, id)
	}

	return nil
}

// Export exports knowledge to a structured format
func (kb *InMemoryKnowledgeBase) Export(ctx context.Context, format string) ([]byte, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	switch format {
	case "json":
		return json.MarshalIndent(kb.knowledge, "", "  ")
	case "csv":
		return kb.exportToCSV()
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// Import imports knowledge from external sources
func (kb *InMemoryKnowledgeBase) Import(ctx context.Context, data []byte, format string) error {
	var importedKnowledge map[string]*Knowledge

	switch format {
	case "json":
		if err := json.Unmarshal(data, &importedKnowledge); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported import format: %s", format)
	}

	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Merge imported knowledge
	for id, knowledge := range importedKnowledge {
		// Check if already exists
		if existing, exists := kb.knowledge[id]; exists {
			// Update if newer or higher confidence
			if knowledge.Timestamp.After(existing.Timestamp) || knowledge.Confidence > existing.Confidence {
				kb.removeFromIndexes(existing)
				kb.knowledge[id] = knowledge
				kb.updateIndexes(knowledge)
			}
		} else {
			kb.knowledge[id] = knowledge
			kb.updateIndexes(knowledge)
		}
	}

	return nil
}

// Helper methods

func (kb *InMemoryKnowledgeBase) updateIndexes(knowledge *Knowledge) {
	// Update type index
	if kb.typeIndex[knowledge.Type] == nil {
		kb.typeIndex[knowledge.Type] = make([]string, 0)
	}
	kb.typeIndex[knowledge.Type] = append(kb.typeIndex[knowledge.Type], knowledge.ID)

	// Update tag index
	for _, tag := range knowledge.Tags {
		if kb.tagIndex[tag] == nil {
			kb.tagIndex[tag] = make([]string, 0)
		}
		kb.tagIndex[tag] = append(kb.tagIndex[tag], knowledge.ID)
	}

	// Update content index for full-text search
	if kb.config.FullTextSearch {
		words := kb.extractWords(knowledge.Content)
		for _, word := range words {
			if kb.contentIndex[word] == nil {
				kb.contentIndex[word] = make([]string, 0)
			}
			kb.contentIndex[word] = append(kb.contentIndex[word], knowledge.ID)
		}
	}
}

func (kb *InMemoryKnowledgeBase) removeFromIndexes(knowledge *Knowledge) {
	// Remove from type index
	kb.typeIndex[knowledge.Type] = kb.removeFromSlice(kb.typeIndex[knowledge.Type], knowledge.ID)

	// Remove from tag index
	for _, tag := range knowledge.Tags {
		kb.tagIndex[tag] = kb.removeFromSlice(kb.tagIndex[tag], knowledge.ID)
	}

	// Remove from content index
	if kb.config.FullTextSearch {
		words := kb.extractWords(knowledge.Content)
		for _, word := range words {
			kb.contentIndex[word] = kb.removeFromSlice(kb.contentIndex[word], knowledge.ID)
		}
	}
}

func (kb *InMemoryKnowledgeBase) removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func (kb *InMemoryKnowledgeBase) getKnowledgeByType(knowledgeType KnowledgeType) []*Knowledge {
	var result []*Knowledge
	for _, id := range kb.typeIndex[knowledgeType] {
		if knowledge, exists := kb.knowledge[id]; exists {
			result = append(result, knowledge)
		}
	}
	return result
}

func (kb *InMemoryKnowledgeBase) getAllKnowledge() []*Knowledge {
	var result []*Knowledge
	for _, knowledge := range kb.knowledge {
		result = append(result, knowledge)
	}
	return result
}

func (kb *InMemoryKnowledgeBase) filterByTags(candidates []*Knowledge, tags []string) []*Knowledge {
	var filtered []*Knowledge
	for _, candidate := range candidates {
		if kb.hasAnyTag(candidate, tags) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func (kb *InMemoryKnowledgeBase) hasAnyTag(knowledge *Knowledge, tags []string) bool {
	for _, requiredTag := range tags {
		for _, knowledgeTag := range knowledge.Tags {
			if knowledgeTag == requiredTag {
				return true
			}
		}
	}
	return false
}

func (kb *InMemoryKnowledgeBase) filterByContent(candidates []*Knowledge, content string) []*Knowledge {
	var filtered []*Knowledge
	content = strings.ToLower(content)
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate.Content), content) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func (kb *InMemoryKnowledgeBase) filterByConfidence(candidates []*Knowledge, minConfidence float64) []*Knowledge {
	var filtered []*Knowledge
	for _, candidate := range candidates {
		if candidate.Confidence >= minConfidence {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func (kb *InMemoryKnowledgeBase) sortKnowledge(knowledge []*Knowledge, sortBy string) {
	switch sortBy {
	case "timestamp":
		sort.Slice(knowledge, func(i, j int) bool {
			return knowledge[i].Timestamp.After(knowledge[j].Timestamp)
		})
	case "confidence":
		sort.Slice(knowledge, func(i, j int) bool {
			return knowledge[i].Confidence > knowledge[j].Confidence
		})
	case "type":
		sort.Slice(knowledge, func(i, j int) bool {
			return knowledge[i].Type < knowledge[j].Type
		})
	default:
		// Default sort by timestamp
		sort.Slice(knowledge, func(i, j int) bool {
			return knowledge[i].Timestamp.After(knowledge[j].Timestamp)
		})
	}
}

func (kb *InMemoryKnowledgeBase) matchesQuery(knowledge *Knowledge, query string) bool {
	// Check content
	if strings.Contains(strings.ToLower(knowledge.Content), query) {
		return true
	}

	// Check tags
	for _, tag := range knowledge.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	// Check source
	if strings.Contains(strings.ToLower(knowledge.Source), query) {
		return true
	}

	return false
}

func (kb *InMemoryKnowledgeBase) sortByRelevance(results []*Knowledge, query string) {
	// Simple relevance scoring
	sort.Slice(results, func(i, j int) bool {
		scoreI := kb.calculateRelevanceScore(results[i], query)
		scoreJ := kb.calculateRelevanceScore(results[j], query)
		return scoreI > scoreJ
	})
}

func (kb *InMemoryKnowledgeBase) calculateRelevanceScore(knowledge *Knowledge, query string) float64 {
	score := 0.0
	query = strings.ToLower(query)
	content := strings.ToLower(knowledge.Content)

	// Exact matches get higher scores
	if strings.Contains(content, query) {
		score += 1.0
	}

	// Word matches
	queryWords := strings.Fields(query)
	contentWords := strings.Fields(content)
	
	for _, qWord := range queryWords {
		for _, cWord := range contentWords {
			if qWord == cWord {
				score += 0.5
			}
		}
	}

	// Confidence factor
	score *= knowledge.Confidence

	return score
}

func (kb *InMemoryKnowledgeBase) extractWords(text string) []string {
	// Simple word extraction
	text = strings.ToLower(text)
	words := strings.Fields(text)
	
	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
	}
	
	var filtered []string
	for _, word := range words {
		if !stopWords[word] && len(word) > 2 {
			filtered = append(filtered, word)
		}
	}
	
	return filtered
}

func (kb *InMemoryKnowledgeBase) evictOldestKnowledge() error {
	if len(kb.knowledge) == 0 {
		return nil
	}

	// Find oldest knowledge
	var oldestID string
	var oldestTime time.Time = time.Now()
	
	for id, knowledge := range kb.knowledge {
		if knowledge.Timestamp.Before(oldestTime) {
			oldestTime = knowledge.Timestamp
			oldestID = id
		}
	}

	// Remove oldest knowledge
	if oldestID != "" {
		knowledge := kb.knowledge[oldestID]
		kb.removeFromIndexes(knowledge)
		delete(kb.knowledge, oldestID)
	}

	return nil
}

func (kb *InMemoryKnowledgeBase) autoPersistRoutine() {
	ticker := time.NewTicker(kb.config.PersistInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := kb.persistence.Save(context.Background(), kb.knowledge); err != nil {
			// Log error but continue
			fmt.Printf("Auto-persistence failed: %v\n", err)
		}
	}
}

func (kb *InMemoryKnowledgeBase) loadFromPersistence() {
	if kb.persistence == nil {
		return
	}

	knowledge, err := kb.persistence.Load(context.Background())
	if err != nil {
		// Log error but continue with empty knowledge base
		fmt.Printf("Failed to load knowledge from persistence: %v\n", err)
		return
	}

	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Load knowledge and rebuild indexes
	for id, k := range knowledge {
		kb.knowledge[id] = k
		kb.updateIndexes(k)
	}
}

func (kb *InMemoryKnowledgeBase) exportToCSV() ([]byte, error) {
	var lines []string
	lines = append(lines, "ID,Type,Content,Source,Timestamp,Confidence,Tags")

	for _, knowledge := range kb.knowledge {
		line := fmt.Sprintf("%s,%s,%s,%s,%s,%.2f,%s",
			knowledge.ID,
			knowledge.Type,
			strings.ReplaceAll(knowledge.Content, ",", ";"),
			knowledge.Source,
			knowledge.Timestamp.Format(time.RFC3339),
			knowledge.Confidence,
			strings.Join(knowledge.Tags, ";"),
		)
		lines = append(lines, line)
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// KnowledgeStatistics provides statistics about the knowledge base
type KnowledgeStatistics struct {
	TotalItems    int
	TypeCounts    map[KnowledgeType]int
	TagCounts     map[string]int
	ConfidenceAvg float64
	OldestItem    time.Time
	NewestItem    time.Time
}

// FilePersistence implementation

func (fp *FilePersistence) Save(ctx context.Context, knowledge map[string]*Knowledge) error {
	data, err := json.MarshalIndent(knowledge, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal knowledge: %w", err)
	}

	// In a real implementation, this would write to file
	// For now, we'll just return nil to indicate success
	return nil
}

func (fp *FilePersistence) Load(ctx context.Context) (map[string]*Knowledge, error) {
	// In a real implementation, this would read from file
	// For now, we'll return empty knowledge
	return make(map[string]*Knowledge), nil
}

func (fp *FilePersistence) Backup(ctx context.Context, filename string) error {
	// Implementation would create a backup file
	return nil
}

func (fp *FilePersistence) Restore(ctx context.Context, filename string) error {
	// Implementation would restore from backup file
	return nil
}

// Enhanced knowledge operations

// AnalyzePatterns identifies patterns in stored knowledge
func (kb *InMemoryKnowledgeBase) AnalyzePatterns(ctx context.Context) (*PatternAnalysis, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	analysis := &PatternAnalysis{
		CommonTags:        make(map[string]int),
		TypeDistribution:  make(map[KnowledgeType]float64),
		ConfidenceTrends:  make([]ConfidenceTrend, 0),
		TemporalPatterns:  make([]TemporalPattern, 0),
	}

	// Analyze tag frequency
	for tag, ids := range kb.tagIndex {
		analysis.CommonTags[tag] = len(ids)
	}

	// Analyze type distribution
	total := float64(len(kb.knowledge))
	for knowledgeType, ids := range kb.typeIndex {
		analysis.TypeDistribution[knowledgeType] = float64(len(ids)) / total
	}

	// Analyze confidence trends over time
	analysis.ConfidenceTrends = kb.analyzeConfidenceTrends()

	// Analyze temporal patterns
	analysis.TemporalPatterns = kb.analyzeTemporalPatterns()

	return analysis, nil
}

func (kb *InMemoryKnowledgeBase) analyzeConfidenceTrends() []ConfidenceTrend {
	// Group knowledge by time periods and calculate average confidence
	trends := make([]ConfidenceTrend, 0)
	
	// For simplicity, analyze by day
	dailyConfidence := make(map[string][]float64)
	
	for _, knowledge := range kb.knowledge {
		day := knowledge.Timestamp.Format("2006-01-02")
		if dailyConfidence[day] == nil {
			dailyConfidence[day] = make([]float64, 0)
		}
		dailyConfidence[day] = append(dailyConfidence[day], knowledge.Confidence)
	}
	
	for day, confidences := range dailyConfidence {
		sum := 0.0
		for _, conf := range confidences {
			sum += conf
		}
		avg := sum / float64(len(confidences))
		
		timestamp, _ := time.Parse("2006-01-02", day)
		trends = append(trends, ConfidenceTrend{
			Timestamp: timestamp,
			Average:   avg,
			Count:     len(confidences),
		})
	}
	
	// Sort by timestamp
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Timestamp.Before(trends[j].Timestamp)
	})
	
	return trends
}

func (kb *InMemoryKnowledgeBase) analyzeTemporalPatterns() []TemporalPattern {
	patterns := make([]TemporalPattern, 0)
	
	// Analyze knowledge creation patterns by hour of day
	hourCounts := make(map[int]int)
	for _, knowledge := range kb.knowledge {
		hour := knowledge.Timestamp.Hour()
		hourCounts[hour]++
	}
	
	// Find peak hours
	var peakHour int
	var peakCount int
	for hour, count := range hourCounts {
		if count > peakCount {
			peakHour = hour
			peakCount = count
		}
	}
	
	patterns = append(patterns, TemporalPattern{
		Type:        "peak_hour",
		Description: fmt.Sprintf("Most knowledge created at hour %d", peakHour),
		Value:       float64(peakHour),
		Frequency:   float64(peakCount) / float64(len(kb.knowledge)),
	})
	
	return patterns
}

// Supporting types for pattern analysis

type PatternAnalysis struct {
	CommonTags       map[string]int
	TypeDistribution map[KnowledgeType]float64
	ConfidenceTrends []ConfidenceTrend
	TemporalPatterns []TemporalPattern
}

type ConfidenceTrend struct {
	Timestamp time.Time
	Average   float64
	Count     int
}

type TemporalPattern struct {
	Type        string
	Description string
	Value       float64
	Frequency   float64
}