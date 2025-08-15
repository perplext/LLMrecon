// Package notification provides functionality for notifying users about updates
package notification

import (
	"encoding/json"
	"fmt"
	"sync"
)

// SeverityLevel represents the severity level of a notification
type SeverityLevel string

const (
	// Info represents an informational notification
	Info SeverityLevel = "info"
	// Warning represents a warning notification
	Warning SeverityLevel = "warning"
	// Critical represents a critical notification
	Critical SeverityLevel = "critical"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// UpdateAvailable represents a notification about an available update
	UpdateAvailable NotificationType = "update-available"
	// UpdateRequired represents a notification about a required update
	UpdateRequired NotificationType = "update-required"
	// SecurityUpdate represents a notification about a security update
	SecurityUpdate NotificationType = "security-update"
	// FeatureUpdate represents a notification about a feature update
	FeatureUpdate NotificationType = "feature-update"
	// MaintenanceUpdate represents a notification about a maintenance update
	MaintenanceUpdate NotificationType = "maintenance-update"
)

// DeliveryStatus represents the delivery status of a notification
type DeliveryStatus string

const (
	// Pending represents a notification that is pending delivery
	Pending DeliveryStatus = "pending"
	// Delivered represents a notification that has been delivered
	Delivered DeliveryStatus = "delivered"
	// Failed represents a notification that failed to be delivered
	Failed DeliveryStatus = "failed"
	// Acknowledged represents a notification that has been acknowledged by the user
	Acknowledged DeliveryStatus = "acknowledged"
	// Dismissed represents a notification that has been dismissed by the user
	Dismissed DeliveryStatus = "dismissed"
)

// Notification represents a notification to be delivered to the user
type Notification struct {
	ID              string          `json:"id"`
	Type            NotificationType `json:"type"`
	Title           string          `json:"title"`
	Message         string          `json:"message"`
	Severity        SeverityLevel   `json:"severity"`
	CreatedAt       time.Time       `json:"createdAt"`
	ExpiresAt       time.Time       `json:"expiresAt,omitempty"`
	ScheduledFor    time.Time       `json:"scheduledFor,omitempty"`
	DeliveredAt     time.Time       `json:"deliveredAt,omitempty"`
	AcknowledgedAt  time.Time       `json:"acknowledgedAt,omitempty"`
	Status          DeliveryStatus  `json:"status"`
	TargetChannels  []string        `json:"targetChannels,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	RequiresAction  bool            `json:"requiresAction"`
	ActionURL       string          `json:"actionUrl,omitempty"`
	ActionLabel     string          `json:"actionLabel,omitempty"`

// NotificationChannel represents a channel for delivering notifications
type NotificationChannel interface {
	// ID returns the unique identifier for the channel
	ID() string
	
	// Name returns the human-readable name of the channel
	Name() string
	
	// Deliver delivers a notification through the channel
	Deliver(notification *Notification) error
	
	// CanDeliver checks if the channel can deliver the notification
	CanDeliver(notification *Notification) bool

// NotificationManager manages notifications and their delivery
type NotificationManager struct {
	channels        map[string]NotificationChannel
	notifications   map[string]*Notification
	history         []*Notification
	historyLimit    int
	storageDir      string
	storageFile     string
	mutex           sync.RWMutex
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(storageDir string) (*NotificationManager, error) {
	if storageDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		storageDir = filepath.Join(homeDir, ".LLMrecon")
	}
	
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	
	manager := &NotificationManager{
		channels:      make(map[string]NotificationChannel),
		notifications: make(map[string]*Notification),
		history:       make([]*Notification, 0),
		historyLimit:  100,
		storageDir:    storageDir,
		storageFile:   filepath.Join(storageDir, "notifications.json"),
	}
	
	// Load existing notifications from storage
	if err := manager.loadFromStorage(); err != nil {
		// Just log the error and continue
		fmt.Printf("Warning: Failed to load notifications from storage: %v\n", err)
	}
	
	return manager, nil

// RegisterChannel registers a notification channel
func (m *NotificationManager) RegisterChannel(channel NotificationChannel) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.channels[channel.ID()] = channel

// UnregisterChannel unregisters a notification channel
func (m *NotificationManager) UnregisterChannel(channelID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	delete(m.channels, channelID)

// CreateNotification creates a new notification
func (m *NotificationManager) CreateNotification(
	notificationType NotificationType,
	title string,
	message string,
	severity SeverityLevel,
	requiresAction bool,
	metadata map[string]string,
) (*Notification, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Generate a unique ID for the notification
	id := fmt.Sprintf("%s-%d", notificationType, time.Now().UnixNano())
	
	notification := &Notification{
		ID:             id,
		Type:           notificationType,
		Title:          title,
		Message:        message,
		Severity:       severity,
		CreatedAt:      time.Now(),
		Status:         Pending,
		Metadata:       metadata,
		RequiresAction: requiresAction,
	}
	
	// Store the notification
	m.notifications[id] = notification
	
	// Add to history
	m.addToHistory(notification)
	
	// Save to storage
	if err := m.saveToStorage(); err != nil {
		return notification, fmt.Errorf("failed to save notification to storage: %w", err)
	}
	
	return notification, nil

// ScheduleNotification schedules a notification for delivery at a specific time
func (m *NotificationManager) ScheduleNotification(notification *Notification, scheduledTime time.Time) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}
	
	// Update scheduled time
	notification.ScheduledFor = scheduledTime
	
	// Save to storage
	if err := m.saveToStorage(); err != nil {
		return fmt.Errorf("failed to save scheduled notification to storage: %w", err)
	}
	
	return nil

// DeliverNotification delivers a notification through all registered channels
func (m *NotificationManager) DeliverNotification(notificationID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	notification, exists := m.notifications[notificationID]
	if !exists {
		return fmt.Errorf("notification not found: %s", notificationID)
	}
	
	// Check if notification is scheduled for the future
	if !notification.ScheduledFor.IsZero() && notification.ScheduledFor.After(time.Now()) {
		return fmt.Errorf("notification is scheduled for future delivery: %s", notificationID)
	}
	
	// Deliver through all registered channels or specific target channels
	var deliveryError error
	deliveredToAny := false
	
	for channelID, channel := range m.channels {
		// Skip channels that are not in the target channels list if specified
		if len(notification.TargetChannels) > 0 {
			isTargetChannel := false
			for _, targetChannel := range notification.TargetChannels {
				if targetChannel == channelID {
					isTargetChannel = true
					break
				}
			}
			if !isTargetChannel {
				continue
			}
		}
		
		// Check if the channel can deliver this notification
		if !channel.CanDeliver(notification) {
			continue
		}
		
		// Deliver the notification
		if err := channel.Deliver(notification); err != nil {
			if deliveryError == nil {
				deliveryError = err
			} else {
				deliveryError = fmt.Errorf("%v; %v", deliveryError, err)
			}
		} else {
			deliveredToAny = true
		}
	}
	
	// Update notification status
	if deliveredToAny {
		notification.Status = Delivered
		notification.DeliveredAt = time.Now()
	} else {
		notification.Status = Failed
	}
	
	// Add to history
	m.addToHistory(notification)
	
	// Save to storage
	if err := m.saveToStorage(); err != nil {
		return fmt.Errorf("failed to save notification status to storage: %w", err)
	}
	
	if !deliveredToAny {
		return fmt.Errorf("failed to deliver notification through any channel: %v", deliveryError)
	}
	
	if deliveryError != nil {
		return fmt.Errorf("partial delivery failure: %v", deliveryError)
	}
	
	return nil

// AcknowledgeNotification marks a notification as acknowledged by the user
func (m *NotificationManager) AcknowledgeNotification(notificationID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	notification, exists := m.notifications[notificationID]
	if !exists {
		return fmt.Errorf("notification not found: %s", notificationID)
	}
	
	// Update notification status
	notification.Status = Acknowledged
	notification.AcknowledgedAt = time.Now()
	
	// Save to storage
	if err := m.saveToStorage(); err != nil {
		return fmt.Errorf("failed to save notification acknowledgment to storage: %w", err)
	}
	
	return nil

// DismissNotification marks a notification as dismissed by the user
func (m *NotificationManager) DismissNotification(notificationID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	notification, exists := m.notifications[notificationID]
	if !exists {
		return fmt.Errorf("notification not found: %s", notificationID)
	}
	
	// Update notification status
	notification.Status = Dismissed
	
	// Save to storage
	if err := m.saveToStorage(); err != nil {
		return fmt.Errorf("failed to save notification dismissal to storage: %w", err)
	}
	
	return nil

// GetNotification returns a notification by ID
func (m *NotificationManager) GetNotification(notificationID string) (*Notification, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	notification, exists := m.notifications[notificationID]
	if !exists {
		return nil, fmt.Errorf("notification not found: %s", notificationID)
	}
	
	return notification, nil

// GetPendingNotifications returns all pending notifications
func (m *NotificationManager) GetPendingNotifications() []*Notification {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	pending := make([]*Notification, 0)
	for _, notification := range m.notifications {
		if notification.Status == Pending {
			pending = append(pending, notification)
		}
	}
	
	return pending

// GetUnacknowledgedNotifications returns all unacknowledged notifications
func (m *NotificationManager) GetUnacknowledgedNotifications() []*Notification {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	unacknowledged := make([]*Notification, 0)
	for _, notification := range m.notifications {
		if notification.Status != Acknowledged && notification.Status != Dismissed {
			unacknowledged = append(unacknowledged, notification)
		}
	}
	
	return unacknowledged

// GetNotificationHistory returns the notification history
func (m *NotificationManager) GetNotificationHistory() []*Notification {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Return a copy of the history to prevent modification
	history := make([]*Notification, len(m.history))
	copy(history, m.history)
	
	return history

// ClearNotificationHistory clears the notification history
func (m *NotificationManager) ClearNotificationHistory() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.history = make([]*Notification, 0)
	
	// Save to storage
	if err := m.saveToStorage(); err != nil {
		return fmt.Errorf("failed to save notification history to storage: %w", err)
	}
	
	return nil

// addToHistory adds a notification to the history
func (m *NotificationManager) addToHistory(notification *Notification) {
	// Add to the beginning of the history
	m.history = append([]*Notification{notification}, m.history...)
	
	// Trim history if it exceeds the limit
	if len(m.history) > m.historyLimit {
		m.history = m.history[:m.historyLimit]
	}
	

// loadFromStorage loads notifications from storage
func (m *NotificationManager) loadFromStorage() error {
	// Check if the storage file exists
	if _, err := os.Stat(m.storageFile); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to load
	}
	
	// Read the storage file
	data, err := os.ReadFile(filepath.Clean(m.storageFile))
	if err != nil {
		return fmt.Errorf("failed to read storage file: %w", err)
	}
	
	// Parse the storage file
	var storage struct {
		Notifications map[string]*Notification `json:"notifications"`
		History       []*Notification          `json:"history"`
	}
	
	if err := json.Unmarshal(data, &storage); err != nil {
		return fmt.Errorf("failed to parse storage file: %w", err)
	}
	
	// Update the manager
	m.notifications = storage.Notifications
	m.history = storage.History
	
	return nil

// saveToStorage saves notifications to storage
func (m *NotificationManager) saveToStorage() error {
	// Create the storage structure
	storage := struct {
		Notifications map[string]*Notification `json:"notifications"`
		History       []*Notification          `json:"history"`
	}{
		Notifications: m.notifications,
		History:       m.history,
	}
	
	// Marshal the storage structure
	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal storage: %w", err)
	}
	
	// Write to the storage file
	if err := os.WriteFile(filepath.Clean(m.storageFile, data, 0600)); err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}
	
	return nil

// ProcessScheduledNotifications processes all scheduled notifications
func (m *NotificationManager) ProcessScheduledNotifications() error {
	m.mutex.Lock()
	
	// Find scheduled notifications that are due
	var dueNotifications []string
	now := time.Now()
	
	for id, notification := range m.notifications {
		if notification.Status == Pending && 
		   !notification.ScheduledFor.IsZero() && 
		   notification.ScheduledFor.Before(now) {
			dueNotifications = append(dueNotifications, id)
		}
	}
	
	m.mutex.Unlock()
	
	// Deliver due notifications
	for _, id := range dueNotifications {
		if err := m.DeliverNotification(id); err != nil {
			return fmt.Errorf("failed to deliver scheduled notification %s: %w", id, err)
		}
	}
	
	return nil

// PurgeExpiredNotifications removes expired notifications
func (m *NotificationManager) PurgeExpiredNotifications() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Find expired notifications
	var expiredIDs []string
	now := time.Now()
	
	for id, notification := range m.notifications {
		if !notification.ExpiresAt.IsZero() && notification.ExpiresAt.Before(now) {
			expiredIDs = append(expiredIDs, id)
		}
	}
	
	// Remove expired notifications
	for _, id := range expiredIDs {
		delete(m.notifications, id)
	}
	
	// Save to storage if any notifications were removed
	if len(expiredIDs) > 0 {
		if err := m.saveToStorage(); err != nil {
			return fmt.Errorf("failed to save after purging expired notifications: %w", err)
		}
	}
	
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
