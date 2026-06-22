package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"sync"
	"time"
)

// RotationPolicy defines the rotation policy for keys
type RotationPolicy struct {
	MaxAge          time.Duration `json:"max_age"`
	RotationPeriod  time.Duration `json:"rotation_period"`
	AutoRotate      bool          `json:"auto_rotate"`
	NotifyBefore    time.Duration `json:"notify_before"`
	ArchiveOldKeys  bool          `json:"archive_old_keys"`
	MaxArchivedKeys int           `json:"max_archived_keys"`
}

// RotationStatus represents the rotation status of a key
type RotationStatus string

const (
	StatusActive   RotationStatus = "active"
	StatusExpiring RotationStatus = "expiring"
	StatusExpired  RotationStatus = "expired"
	StatusRotated  RotationStatus = "rotated"
	StatusArchived RotationStatus = "archived"
)

// KeyRotationInfo contains information about key rotation
type KeyRotationInfo struct {
	KeyID          string         `json:"key_id"`
	Status         RotationStatus `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	LastRotated    *time.Time     `json:"last_rotated"`
	NextRotation   *time.Time     `json:"next_rotation"`
	RotationCount  int            `json:"rotation_count"`
	Policy         RotationPolicy `json:"policy"`
	PreviousKeyIDs []string       `json:"previous_key_ids"`
}

// RotationEvent represents a key rotation event
type RotationEvent struct {
	KeyID        string    `json:"key_id"`
	OldKeyID     string    `json:"old_key_id"`
	NewKeyID     string    `json:"new_key_id"`
	RotationType string    `json:"rotation_type"` // "manual", "scheduled", "emergency"
	Timestamp    time.Time `json:"timestamp"`
	Reason       string    `json:"reason"`
	RotatedBy    string    `json:"rotated_by"`
}

// KeyRotator handles key rotation operations
type KeyRotator struct {
	keystore        *Keystore
	rotationInfos   map[string]*KeyRotationInfo
	policies        map[string]*RotationPolicy
	rotationHistory []RotationEvent
	mu              sync.RWMutex
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewKeyRotator creates a new key rotator
func NewKeyRotator(ks *Keystore) *KeyRotator {
	// stopChan is initialized in a *closed* state to mean "not running".
	// StartAutoRotation distinguishes running from stopped by reading this
	// channel: a closed channel selects immediately (proceed to start), an
	// open channel falls through to default (already running). If it were
	// created open, the very first StartAutoRotation would wrongly report
	// "already running" and the worker would never launch.
	stopChan := make(chan struct{})
	close(stopChan)
	return &KeyRotator{
		keystore:        ks,
		rotationInfos:   make(map[string]*KeyRotationInfo),
		policies:        make(map[string]*RotationPolicy),
		rotationHistory: make([]RotationEvent, 0),
		stopChan:        stopChan,
	}
}

// SetRotationPolicy sets the rotation policy for a key
func (kr *KeyRotator) SetRotationPolicy(keyID string, policy RotationPolicy) error {
	if keyID == "" {
		return errors.New("key ID cannot be empty")
	}

	kr.mu.Lock()
	defer kr.mu.Unlock()

	// Check if key exists
	if _, exists := kr.keystore.keys[keyID]; !exists {
		return fmt.Errorf("key with ID %s not found", keyID)
	}

	kr.policies[keyID] = &policy

	// Update rotation info
	if info, exists := kr.rotationInfos[keyID]; exists {
		info.Policy = policy
		kr.updateNextRotation(info)
	} else {
		keyEntry := kr.keystore.keys[keyID]
		info := &KeyRotationInfo{
			KeyID:          keyID,
			Status:         StatusActive,
			CreatedAt:      keyEntry.CreatedAt,
			Policy:         policy,
			PreviousKeyIDs: make([]string, 0),
		}
		kr.updateNextRotation(info)
		kr.rotationInfos[keyID] = info
	}

	return nil
}

// GetRotationPolicy returns the rotation policy for a key
func (kr *KeyRotator) GetRotationPolicy(keyID string) (*RotationPolicy, error) {
	if keyID == "" {
		return nil, errors.New("key ID cannot be empty")
	}

	kr.mu.RLock()
	defer kr.mu.RUnlock()

	policy, exists := kr.policies[keyID]
	if !exists {
		return nil, fmt.Errorf("no rotation policy set for key %s", keyID)
	}

	// Return a copy
	policyCopy := *policy
	return &policyCopy, nil
}

// GetRotationInfo returns rotation information for a key
func (kr *KeyRotator) GetRotationInfo(keyID string) (*KeyRotationInfo, error) {
	if keyID == "" {
		return nil, errors.New("key ID cannot be empty")
	}

	kr.mu.RLock()
	defer kr.mu.RUnlock()

	info, exists := kr.rotationInfos[keyID]
	if !exists {
		return nil, fmt.Errorf("no rotation info found for key %s", keyID)
	}

	// Return a copy
	infoCopy := *info
	infoCopy.PreviousKeyIDs = make([]string, len(info.PreviousKeyIDs))
	copy(infoCopy.PreviousKeyIDs, info.PreviousKeyIDs)

	return &infoCopy, nil
}

// RotateKey manually rotates a key
func (kr *KeyRotator) RotateKey(keyID string, reason string, rotatedBy string) error {
	if keyID == "" {
		return errors.New("key ID cannot be empty")
	}

	kr.mu.Lock()
	defer kr.mu.Unlock()

	return kr.rotateKeyInternal(keyID, "manual", reason, rotatedBy)
}

// RotateAllKeys rotates all keys that have rotation policies
func (kr *KeyRotator) RotateAllKeys(reason string, rotatedBy string) error {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	var rotationErrors []string

	for keyID := range kr.policies {
		if err := kr.rotateKeyInternal(keyID, "bulk", reason, rotatedBy); err != nil {
			rotationErrors = append(rotationErrors, fmt.Sprintf("failed to rotate key %s: %v", keyID, err))
		}
	}

	if len(rotationErrors) > 0 {
		return fmt.Errorf("rotation errors: %v", rotationErrors)
	}

	return nil
}

// CheckRotationStatus checks the rotation status of all keys
func (kr *KeyRotator) CheckRotationStatus() map[string]RotationStatus {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	statusMap := make(map[string]RotationStatus)

	for keyID, info := range kr.rotationInfos {
		status := kr.calculateRotationStatus(info)
		statusMap[keyID] = status
	}

	return statusMap
}

// GetExpiringKeys returns keys that are approaching expiration
func (kr *KeyRotator) GetExpiringKeys() []*KeyRotationInfo {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	var expiringKeys []*KeyRotationInfo

	for _, info := range kr.rotationInfos {
		if kr.calculateRotationStatus(info) == StatusExpiring {
			infoCopy := *info
			infoCopy.PreviousKeyIDs = make([]string, len(info.PreviousKeyIDs))
			copy(infoCopy.PreviousKeyIDs, info.PreviousKeyIDs)
			expiringKeys = append(expiringKeys, &infoCopy)
		}
	}

	return expiringKeys
}

// GetExpiredKeys returns keys that have expired
func (kr *KeyRotator) GetExpiredKeys() []*KeyRotationInfo {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	var expiredKeys []*KeyRotationInfo

	for _, info := range kr.rotationInfos {
		if kr.calculateRotationStatus(info) == StatusExpired {
			infoCopy := *info
			infoCopy.PreviousKeyIDs = make([]string, len(info.PreviousKeyIDs))
			copy(infoCopy.PreviousKeyIDs, info.PreviousKeyIDs)
			expiredKeys = append(expiredKeys, &infoCopy)
		}
	}

	return expiredKeys
}

// StartAutoRotation starts the automatic rotation process
func (kr *KeyRotator) StartAutoRotation() error {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	// Check if already running
	select {
	case <-kr.stopChan:
		// Channel is closed, recreate it
		kr.stopChan = make(chan struct{})
	default:
		// Already running
		return errors.New("auto rotation is already running")
	}

	kr.wg.Add(1)
	go kr.autoRotationWorker()

	return nil
}

// StopAutoRotation stops the automatic rotation process
func (kr *KeyRotator) StopAutoRotation() error {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	select {
	case <-kr.stopChan:
		return errors.New("auto rotation is not running")
	default:
		close(kr.stopChan)
	}

	kr.wg.Wait()
	return nil
}

// GetRotationHistory returns the rotation history
func (kr *KeyRotator) GetRotationHistory() []RotationEvent {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	history := make([]RotationEvent, len(kr.rotationHistory))
	copy(history, kr.rotationHistory)

	return history
}

// rotateKeyInternal performs the actual key rotation
func (kr *KeyRotator) rotateKeyInternal(keyID string, rotationType string, reason string, rotatedBy string) error {
	// Get the original key
	originalKey, exists := kr.keystore.keys[keyID]
	if !exists {
		return fmt.Errorf("key with ID %s not found", keyID)
	}

	// Generate new key of the same type
	newKeyID := fmt.Sprintf("%s-rotated-%d", keyID, time.Now().Unix())
	var newKey interface{}

	switch originalKey.Type {
	case KeyTypeRSA:
		// Generate new RSA key
		rsaKey, rsaErr := rsa.GenerateKey(rand.Reader, 2048)
		if rsaErr != nil {
			return fmt.Errorf("failed to generate new RSA key: %w", rsaErr)
		}
		newKey = rsaKey
	case KeyTypeAES:
		// Generate new AES key
		aesKey := make([]byte, 32) // 256-bit key
		if _, randErr := rand.Read(aesKey); randErr != nil {
			return fmt.Errorf("failed to generate new AES key: %w", randErr)
		}
		newKey = aesKey
	default:
		return fmt.Errorf("unsupported key type for rotation: %s", originalKey.Type)
	}

	// Create new key entry
	newKeyEntry := &KeyEntry{
		ID:        newKeyID,
		Type:      originalKey.Type,
		Algorithm: originalKey.Algorithm,
		Key:       newKey,
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Copy metadata
	for k, v := range originalKey.Metadata {
		newKeyEntry.Metadata[k] = v
	}

	// Add rotation metadata
	newKeyEntry.Metadata["rotated_from"] = keyID
	newKeyEntry.Metadata["rotation_reason"] = reason
	newKeyEntry.Metadata["rotated_by"] = rotatedBy

	// Store the new key
	kr.keystore.keys[newKeyID] = newKeyEntry

	// Update rotation info
	info, exists := kr.rotationInfos[keyID]
	if !exists {
		info = &KeyRotationInfo{
			KeyID:          keyID,
			Status:         StatusActive,
			CreatedAt:      originalKey.CreatedAt,
			PreviousKeyIDs: make([]string, 0),
		}
		kr.rotationInfos[keyID] = info
	}

	// Update rotation info
	now := time.Now()
	info.LastRotated = &now
	info.RotationCount++
	info.Status = StatusRotated
	info.PreviousKeyIDs = append(info.PreviousKeyIDs, keyID)

	// Create rotation info for new key
	newInfo := &KeyRotationInfo{
		KeyID:          newKeyID,
		Status:         StatusActive,
		CreatedAt:      now,
		LastRotated:    &now,
		RotationCount:  0,
		PreviousKeyIDs: make([]string, 0),
	}

	// Copy policy if exists
	if policy, policyExists := kr.policies[keyID]; policyExists {
		newInfo.Policy = *policy
		kr.policies[newKeyID] = policy
		kr.updateNextRotation(newInfo)
	}

	kr.rotationInfos[newKeyID] = newInfo

	// Archive old key if policy requires it
	if policy, policyExists := kr.policies[keyID]; policyExists && policy.ArchiveOldKeys {
		originalKey.Metadata["archived_at"] = time.Now().Format(time.RFC3339)
		originalKey.Metadata["status"] = "archived"
		info.Status = StatusArchived

		// Cleanup old archived keys if limit is set
		if policy.MaxArchivedKeys > 0 {
			kr.cleanupArchivedKeys(keyID, policy.MaxArchivedKeys)
		}
	} else {
		// Remove old key if not archiving
		delete(kr.keystore.keys, keyID)
		delete(kr.rotationInfos, keyID)
		delete(kr.policies, keyID)
	}

	// Record rotation event
	event := RotationEvent{
		KeyID:        newKeyID,
		OldKeyID:     keyID,
		NewKeyID:     newKeyID,
		RotationType: rotationType,
		Timestamp:    now,
		Reason:       reason,
		RotatedBy:    rotatedBy,
	}

	kr.rotationHistory = append(kr.rotationHistory, event)

	return nil
}

// updateNextRotation calculates and updates the next rotation time
func (kr *KeyRotator) updateNextRotation(info *KeyRotationInfo) {
	if info.Policy.RotationPeriod > 0 {
		var baseTime time.Time
		if info.LastRotated != nil {
			baseTime = *info.LastRotated
		} else {
			baseTime = info.CreatedAt
		}

		nextRotation := baseTime.Add(info.Policy.RotationPeriod)
		info.NextRotation = &nextRotation
	}
}

// calculateRotationStatus calculates the current rotation status
func (kr *KeyRotator) calculateRotationStatus(info *KeyRotationInfo) RotationStatus {
	now := time.Now()

	// Check if already archived
	if info.Status == StatusArchived {
		return StatusArchived
	}

	// Check if already rotated
	if info.Status == StatusRotated {
		return StatusRotated
	}

	// Check age-based expiration
	if info.Policy.MaxAge > 0 {
		expirationTime := info.CreatedAt.Add(info.Policy.MaxAge)
		if now.After(expirationTime) {
			return StatusExpired
		}

		// Check if expiring soon
		if info.Policy.NotifyBefore > 0 {
			notifyTime := expirationTime.Add(-info.Policy.NotifyBefore)
			if now.After(notifyTime) {
				return StatusExpiring
			}
		}
	}

	// Check rotation period-based status
	if info.NextRotation != nil {
		if now.After(*info.NextRotation) {
			return StatusExpired
		}

		// Check if rotation is due soon
		if info.Policy.NotifyBefore > 0 {
			notifyTime := info.NextRotation.Add(-info.Policy.NotifyBefore)
			if now.After(notifyTime) {
				return StatusExpiring
			}
		}
	}

	return StatusActive
}

// cleanupArchivedKeys removes old archived keys beyond the limit
func (kr *KeyRotator) cleanupArchivedKeys(baseKeyID string, maxArchived int) {
	// Find all archived keys for this base key
	var archivedKeys []string

	for keyID, info := range kr.rotationInfos {
		if info.Status == StatusArchived {
			// Check if this is an archived version of the base key
			for _, prevID := range info.PreviousKeyIDs {
				if prevID == baseKeyID {
					archivedKeys = append(archivedKeys, keyID)
					break
				}
			}
		}
	}

	// Remove excess archived keys (keep the most recent ones)
	if len(archivedKeys) > maxArchived {
		// Sort by creation time (oldest first)
		// In a real implementation, you'd sort properly
		excessCount := len(archivedKeys) - maxArchived
		for i := 0; i < excessCount; i++ {
			keyToRemove := archivedKeys[i]
			delete(kr.keystore.keys, keyToRemove)
			delete(kr.rotationInfos, keyToRemove)
			delete(kr.policies, keyToRemove)
		}
	}
}

// autoRotationWorker runs the automatic rotation process
func (kr *KeyRotator) autoRotationWorker() {
	defer kr.wg.Done()

	ticker := time.NewTicker(time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-kr.stopChan:
			return
		case <-ticker.C:
			kr.performAutoRotations()
		}
	}
}

// performAutoRotations checks and performs automatic rotations
func (kr *KeyRotator) performAutoRotations() {
	kr.mu.Lock()
	defer kr.mu.Unlock()

	for keyID, policy := range kr.policies {
		if !policy.AutoRotate {
			continue
		}

		info, exists := kr.rotationInfos[keyID]
		if !exists {
			continue
		}

		status := kr.calculateRotationStatus(info)
		if status == StatusExpired {
			err := kr.rotateKeyInternal(keyID, "scheduled", "automatic rotation due to policy", "system")
			if err != nil {
				// Log error in a real implementation
				fmt.Printf("Failed to auto-rotate key %s: %v\n", keyID, err)
			}
		}
	}
}
