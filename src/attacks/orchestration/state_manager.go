package orchestration

import (
	"encoding/json"
	"fmt"
	"sync"
)

// StateManager handles persistence and recovery of conversation states
type StateManager struct {
	states       map[string]*ConversationState
	checkpoints  map[string][]*StateCheckpoint
	stateStore   StateStore
	config       StateConfig
	mu           sync.RWMutex
}

// StateConfig configures the state manager
type StateConfig struct {
	CheckpointInterval   time.Duration
	MaxCheckpoints       int
	PersistenceEnabled   bool
	CompressionEnabled   bool
	EncryptionEnabled    bool
}

// StateStore interface for persisting states
type StateStore interface {
	Save(id string, state []byte) error
	Load(id string) ([]byte, error)
	Delete(id string) error
	List() ([]string, error)
}

// StateCheckpoint represents a saved state
type StateCheckpoint struct {
	ID        string
	StateID   string
	Timestamp time.Time
	TurnCount int
	Data      []byte
	Metadata  map[string]interface{}
}

// AttackState tracks the overall attack progress
type AttackState struct {
	ID              string
	TargetModel     string
	AttackType      string
	StartTime       time.Time
	EndTime         *time.Time
	Status          AttackStatus
	SuccessMetrics  map[string]float64
	Conversations   []*ConversationState
	Vulnerabilities []Vulnerability
	mu              sync.RWMutex
}

// AttackStatus represents the attack state
type AttackStatus string

const (
	AttackStatusActive    AttackStatus = "active"
	AttackStatusSuccess   AttackStatus = "success"
	AttackStatusFailed    AttackStatus = "failed"
	AttackStatusSuspended AttackStatus = "suspended"
	AttackStatusCompleted AttackStatus = "completed"
)

// Vulnerability represents a discovered vulnerability
type Vulnerability struct {
	ID          string
	Type        string
	Severity    string
	Description string
	Evidence    []Evidence
	Discovered  time.Time
}

// Evidence supports a vulnerability finding
type Evidence struct {
	ConversationID string
	TurnNumber     int
	Prompt         string
	Response       string
	Timestamp      time.Time
}

// NewStateManager creates a new state manager
func NewStateManager(config StateConfig, store StateStore) *StateManager {
	sm := &StateManager{
		states:      make(map[string]*ConversationState),
		checkpoints: make(map[string][]*StateCheckpoint),
		stateStore:  store,
		config:      config,
	}

	// Start checkpoint routine if enabled
	if config.PersistenceEnabled && config.CheckpointInterval > 0 {
		go sm.checkpointRoutine()
	}

	return sm
}

// CreateAttackState initializes a new attack state
func (sm *StateManager) CreateAttackState(targetModel, attackType string) *AttackState {
	return &AttackState{
		ID:              generateAttackID(),
		TargetModel:     targetModel,
		AttackType:      attackType,
		StartTime:       time.Now(),
		Status:          AttackStatusActive,
		SuccessMetrics:  make(map[string]float64),
		Conversations:   []*ConversationState{},
		Vulnerabilities: []Vulnerability{},
	}
}

// RegisterConversation adds a conversation to tracking
func (sm *StateManager) RegisterConversation(state *ConversationState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.states[state.ID] = state
	
	// Create initial checkpoint
	if sm.config.PersistenceEnabled {
		return sm.createCheckpoint(state)
	}

	return nil
}

// UpdateState updates conversation state and creates checkpoint if needed
func (sm *StateManager) UpdateState(stateID string, updater func(*ConversationState)) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state, exists := sm.states[stateID]
	if !exists {
		return fmt.Errorf("state not found: %s", stateID)
	}

	// Apply update
	updater(state)

	// Check if checkpoint needed
	if sm.shouldCheckpoint(state) {
		return sm.createCheckpoint(state)
	}

	return nil
}

// RecordVulnerability adds a discovered vulnerability
func (sm *StateManager) RecordVulnerability(attackState *AttackState, vuln Vulnerability) {
	attackState.mu.Lock()
	defer attackState.mu.Unlock()

	vuln.ID = generateVulnID()
	vuln.Discovered = time.Now()
	attackState.Vulnerabilities = append(attackState.Vulnerabilities, vuln)

	// Update success metrics
	sm.updateSuccessMetrics(attackState)
}

// RecoverState recovers a conversation state from checkpoint
func (sm *StateManager) RecoverState(stateID string) (*ConversationState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check in-memory first
	if state, exists := sm.states[stateID]; exists {
		return state, nil
	}

	// Load from store
	if sm.stateStore != nil {
		data, err := sm.stateStore.Load(stateID)
		if err != nil {
			return nil, err
		}

		state := &ConversationState{}
		if err := sm.deserializeState(data, state); err != nil {
			return nil, err
		}

		sm.states[stateID] = state
		return state, nil
	}

	return nil, fmt.Errorf("state not found: %s", stateID)
}

// GetCheckpoints retrieves checkpoints for a state
func (sm *StateManager) GetCheckpoints(stateID string) []*StateCheckpoint {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.checkpoints[stateID]
}

// RollbackToCheckpoint restores state to a checkpoint
func (sm *StateManager) RollbackToCheckpoint(checkpointID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Find checkpoint
	var checkpoint *StateCheckpoint
	for _, checkpoints := range sm.checkpoints {
		for _, cp := range checkpoints {
			if cp.ID == checkpointID {
				checkpoint = cp
				break
			}
		}
	}

	if checkpoint == nil {
		return fmt.Errorf("checkpoint not found: %s", checkpointID)
	}

	// Restore state
	state := &ConversationState{}
	if err := sm.deserializeState(checkpoint.Data, state); err != nil {
		return err
	}

	sm.states[checkpoint.StateID] = state
	return nil
}

// AnalyzeAttackProgress analyzes overall attack progress
func (sm *StateManager) AnalyzeAttackProgress(attackState *AttackState) AttackAnalysis {
	attackState.mu.RLock()
	defer attackState.mu.RUnlock()

	analysis := AttackAnalysis{
		AttackID:        attackState.ID,
		Duration:        time.Since(attackState.StartTime),
		ConversationCount: len(attackState.Conversations),
		VulnerabilityCount: len(attackState.Vulnerabilities),
		SuccessRate:     sm.calculateSuccessRate(attackState),
		Insights:        []string{},
	}

	// Analyze vulnerabilities
	vulnTypes := make(map[string]int)
	for _, vuln := range attackState.Vulnerabilities {
		vulnTypes[vuln.Type]++
	}

	// Generate insights
	if analysis.SuccessRate > 0.7 {
		analysis.Insights = append(analysis.Insights, "High success rate indicates model is vulnerable")
	}

	if len(vulnTypes) > 3 {
		analysis.Insights = append(analysis.Insights, "Multiple vulnerability types discovered")
	}

	for vulnType, count := range vulnTypes {
		if count > 2 {
			analysis.Insights = append(analysis.Insights, 
				fmt.Sprintf("Repeated %s vulnerabilities suggest systematic weakness", vulnType))
		}
	}

	return analysis
}

// AttackAnalysis contains attack analysis results
type AttackAnalysis struct {
	AttackID           string
	Duration           time.Duration
	ConversationCount  int
	VulnerabilityCount int
	SuccessRate        float64
	Insights           []string
	VulnerabilityTypes map[string]int
}

// checkpointRoutine periodically creates checkpoints
func (sm *StateManager) checkpointRoutine() {
	ticker := time.NewTicker(sm.config.CheckpointInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.RLock()
		states := make([]*ConversationState, 0, len(sm.states))
		for _, state := range sm.states {
			states = append(states, state)
		}
		sm.mu.RUnlock()

		for _, state := range states {
			if sm.shouldCheckpoint(state) {
				sm.mu.Lock()
				sm.createCheckpoint(state)
				sm.mu.Unlock()
			}
		}
	}
}

// shouldCheckpoint determines if checkpoint is needed
func (sm *StateManager) shouldCheckpoint(state *ConversationState) bool {
	checkpoints := sm.checkpoints[state.ID]
	if len(checkpoints) == 0 {
		return true
	}

	lastCheckpoint := checkpoints[len(checkpoints)-1]
	
	// Checkpoint if significant progress
	turnDiff := state.TurnCount - lastCheckpoint.TurnCount
	timeDiff := time.Since(lastCheckpoint.Timestamp)

	return turnDiff >= 5 || timeDiff >= sm.config.CheckpointInterval
}

// createCheckpoint creates a new checkpoint
func (sm *StateManager) createCheckpoint(state *ConversationState) error {
	data, err := sm.serializeState(state)
	if err != nil {
		return err
	}

	checkpoint := &StateCheckpoint{
		ID:        generateCheckpointID(),
		StateID:   state.ID,
		Timestamp: time.Now(),
		TurnCount: state.TurnCount,
		Data:      data,
		Metadata: map[string]interface{}{
			"strategy": state.CurrentStrategy,
			"metrics":  state.SuccessMetrics,
		},
	}

	// Add to checkpoints
	sm.checkpoints[state.ID] = append(sm.checkpoints[state.ID], checkpoint)

	// Trim old checkpoints
	if len(sm.checkpoints[state.ID]) > sm.config.MaxCheckpoints {
		sm.checkpoints[state.ID] = sm.checkpoints[state.ID][1:]
	}

	// Persist if enabled
	if sm.stateStore != nil {
		return sm.stateStore.Save(checkpoint.ID, data)
	}

	return nil
}

// serializeState converts state to bytes
func (sm *StateManager) serializeState(state *ConversationState) ([]byte, error) {
	state.mu.RLock()
	defer state.mu.RUnlock()

	data, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}

	// Apply compression if enabled
	if sm.config.CompressionEnabled {
		data = compress(data)
	}

	// Apply encryption if enabled
	if sm.config.EncryptionEnabled {
		data = encrypt(data)
	}

	return data, nil
}

// deserializeState converts bytes to state
func (sm *StateManager) deserializeState(data []byte, state *ConversationState) error {
	// Decrypt if needed
	if sm.config.EncryptionEnabled {
		data = decrypt(data)
	}

	// Decompress if needed
	if sm.config.CompressionEnabled {
		data = decompress(data)
	}

	return json.Unmarshal(data, state)
}

// updateSuccessMetrics updates attack success metrics
func (sm *StateManager) updateSuccessMetrics(attackState *AttackState) {
	totalConversations := float64(len(attackState.Conversations))
	if totalConversations == 0 {
		return
	}

	successfulConversations := 0.0
	totalTurns := 0.0
	totalExtractions := 0.0

	for _, conv := range attackState.Conversations {
		conv.mu.RLock()
		if len(conv.ExtractedInfo) > 0 {
			successfulConversations++
		}
		totalTurns += float64(conv.TurnCount)
		totalExtractions += float64(len(conv.ExtractedInfo))
		conv.mu.RUnlock()
	}

	attackState.SuccessMetrics["conversation_success_rate"] = successfulConversations / totalConversations
	attackState.SuccessMetrics["avg_turns_per_conversation"] = totalTurns / totalConversations
	attackState.SuccessMetrics["avg_extractions_per_conversation"] = totalExtractions / totalConversations
	attackState.SuccessMetrics["vulnerability_discovery_rate"] = float64(len(attackState.Vulnerabilities)) / totalConversations
}

// calculateSuccessRate calculates overall success rate
func (sm *StateManager) calculateSuccessRate(attackState *AttackState) float64 {
	weights := map[string]float64{
		"conversation_success_rate":    0.3,
		"vulnerability_discovery_rate": 0.4,
		"avg_extractions_per_conversation": 0.3,
	}

	totalScore := 0.0
	totalWeight := 0.0

	for metric, weight := range weights {
		if value, exists := attackState.SuccessMetrics[metric]; exists {
			// Normalize extraction rate
			if metric == "avg_extractions_per_conversation" {
				value = value / 10.0 // Normalize to 0-1 assuming 10 is max
				if value > 1.0 {
					value = 1.0
				}
			}
			totalScore += value * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return totalScore / totalWeight
}

// Helper functions for compression and encryption (placeholders)
func compress(data []byte) []byte {
	// Implement compression
	return data
}

func decompress(data []byte) []byte {
	// Implement decompression
	return data
}

func encrypt(data []byte) []byte {
	// Implement encryption
	return data
}

func decrypt(data []byte) []byte {
	// Implement decryption
	return data
}

func generateAttackID() string {
	return fmt.Sprintf("attack_%d", time.Now().UnixNano())
}

func generateVulnID() string {
	return fmt.Sprintf("vuln_%d", time.Now().UnixNano())
}

func generateCheckpointID() string {
	return fmt.Sprintf("checkpoint_%d", time.Now().UnixNano())
}

// InMemoryStateStore provides in-memory state storage
type InMemoryStateStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		data: make(map[string][]byte),
	}
}

func (s *InMemoryStateStore) Save(id string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = data
	return nil
}

func (s *InMemoryStateStore) Load(id string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, exists := s.data[id]
	if !exists {
		return nil, fmt.Errorf("not found: %s", id)
	}
	return data, nil
}

func (s *InMemoryStateStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, id)
	return nil
}

func (s *InMemoryStateStore) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.data))
	for id := range s.data {
		ids = append(ids, id)
	}
	return ids, nil
}