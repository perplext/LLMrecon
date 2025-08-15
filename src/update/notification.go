// Package update provides functionality for checking and applying updates
package update

import (
	"encoding/json"
	"fmt"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// UpdateStarted indicates an update has started
	UpdateStarted NotificationType = "update_started"
	// UpdateCompleted indicates an update has completed successfully
	UpdateCompleted NotificationType = "update_completed"
	// UpdateFailed indicates an update has failed
	UpdateFailed NotificationType = "update_failed"
	// UpdateRolledBack indicates an update has been rolled back
	UpdateRolledBack NotificationType = "update_rolled_back"
	// ComponentUpdated indicates a component has been updated
	ComponentUpdated NotificationType = "component_updated"
	// VerificationFailed indicates a verification has failed
	VerificationFailed NotificationType = "verification_failed"
)

// Notification represents a notification about an update event
type Notification struct {
	// Type is the type of notification
	Type NotificationType `json:"type"`
	// Timestamp is the time the notification was created
	Timestamp time.Time `json:"timestamp"`
	// Message is the notification message
	Message string `json:"message"`
	// TransactionID is the ID of the transaction associated with the notification
	TransactionID string `json:"transaction_id,omitempty"`
	// PackageID is the ID of the package associated with the notification
	PackageID string `json:"package_id,omitempty"`
	// Component is the component associated with the notification
	Component string `json:"component,omitempty"`
	// Details contains additional details about the notification
	Details map[string]interface{} `json:"details,omitempty"`

// NotificationHandler is an interface for handling notifications
type NotificationHandler interface {
	// HandleNotification handles a notification
	HandleNotification(notification *Notification) error

// ConsoleNotificationHandler handles notifications by writing to a console
type ConsoleNotificationHandler struct {
	// Writer is the writer for notification output
	Writer io.Writer

// NewConsoleNotificationHandler creates a new console notification handler
func NewConsoleNotificationHandler(writer io.Writer) *ConsoleNotificationHandler {
	return &ConsoleNotificationHandler{
		Writer: writer,
	}

// HandleNotification handles a notification by writing to the console
func (h *ConsoleNotificationHandler) HandleNotification(notification *Notification) error {
	// Format notification for console output
	var typePrefix string
	switch notification.Type {
	case UpdateStarted:
		typePrefix = "UPDATE STARTED"
	case UpdateCompleted:
		typePrefix = "UPDATE COMPLETED"
	case UpdateFailed:
		typePrefix = "UPDATE FAILED"
	case UpdateRolledBack:
		typePrefix = "UPDATE ROLLED BACK"
	case ComponentUpdated:
		typePrefix = "COMPONENT UPDATED"
	case VerificationFailed:
		typePrefix = "VERIFICATION FAILED"
	default:
		typePrefix = "NOTIFICATION"
	}

	// Write notification to console
	fmt.Fprintf(h.Writer, "[%s] [%s] %s", 
		notification.Timestamp.Format(time.RFC3339),
		typePrefix,
		notification.Message)
	
	// Include transaction ID if provided
	if notification.TransactionID != "" {
		fmt.Fprintf(h.Writer, " (Transaction: %s)", notification.TransactionID)
	}
	
	// End line
	fmt.Fprintln(h.Writer)

	return nil

// JSONNotificationHandler handles notifications by writing JSON to a writer
type JSONNotificationHandler struct {
	// Writer is the writer for JSON notification output
	Writer io.Writer

// NewJSONNotificationHandler creates a new JSON notification handler
func NewJSONNotificationHandler(writer io.Writer) *JSONNotificationHandler {
	return &JSONNotificationHandler{
		Writer: writer,
	}

// HandleNotification handles a notification by writing JSON to the writer
func (h *JSONNotificationHandler) HandleNotification(notification *Notification) error {
	// Marshal notification to JSON
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification to JSON: %w", err)
	}

	// Write JSON notification
	if _, err := h.Writer.Write(data); err != nil {
		return fmt.Errorf("failed to write JSON notification: %w", err)
	}

	// Write newline
	if _, err := h.Writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write JSON notification: %w", err)
	}

	return nil

// WebhookNotificationHandler handles notifications by sending them to a webhook
type WebhookNotificationHandler struct {
	// URL is the URL of the webhook
	URL string
	// Headers are the HTTP headers to include in the webhook request
	Headers map[string]string

// NewWebhookNotificationHandler creates a new webhook notification handler
func NewWebhookNotificationHandler(url string, headers map[string]string) *WebhookNotificationHandler {
	return &WebhookNotificationHandler{
		URL:     url,
		Headers: headers,
	}

// HandleNotification handles a notification by sending it to a webhook
func (h *WebhookNotificationHandler) HandleNotification(notification *Notification) error {
	// In a real implementation, this would send an HTTP request to the webhook URL
	// For now, we'll just log that webhook notification is not implemented
	fmt.Printf("Webhook notification not implemented (URL: %s)\n", h.URL)
	return nil

// NotificationManager manages notification handlers
type NotificationManager struct {
	// Handlers is the list of notification handlers
	Handlers []NotificationHandler

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		Handlers: make([]NotificationHandler, 0),
	}

// AddHandler adds a notification handler
func (m *NotificationManager) AddHandler(handler NotificationHandler) {
	m.Handlers = append(m.Handlers, handler)

// SendNotification sends a notification to all handlers
func (m *NotificationManager) SendNotification(notification *Notification) error {
	var lastErr error
	for _, handler := range m.Handlers {
		if err := handler.HandleNotification(notification); err != nil {
			lastErr = err
		}
	}
	return lastErr

// NotifyUpdateStarted sends a notification that an update has started
func (m *NotificationManager) NotifyUpdateStarted(transactionID, packageID string, details map[string]interface{}) error {
	notification := &Notification{
		Type:          UpdateStarted,
		Timestamp:     time.Now(),
		Message:       fmt.Sprintf("Update started for package %s", packageID),
		TransactionID: transactionID,
		PackageID:     packageID,
		Details:       details,
	}
	return m.SendNotification(notification)

// NotifyUpdateCompleted sends a notification that an update has completed
func (m *NotificationManager) NotifyUpdateCompleted(transactionID, packageID string, details map[string]interface{}) error {
	notification := &Notification{
		Type:          UpdateCompleted,
		Timestamp:     time.Now(),
		Message:       fmt.Sprintf("Update completed for package %s", packageID),
		TransactionID: transactionID,
		PackageID:     packageID,
		Details:       details,
	}
	return m.SendNotification(notification)

// NotifyUpdateFailed sends a notification that an update has failed
func (m *NotificationManager) NotifyUpdateFailed(transactionID, packageID, reason string, details map[string]interface{}) error {
	notification := &Notification{
		Type:          UpdateFailed,
		Timestamp:     time.Now(),
		Message:       fmt.Sprintf("Update failed for package %s: %s", packageID, reason),
		TransactionID: transactionID,
		PackageID:     packageID,
		Details:       details,
	}
	return m.SendNotification(notification)

// NotifyUpdateRolledBack sends a notification that an update has been rolled back
func (m *NotificationManager) NotifyUpdateRolledBack(transactionID, packageID string, details map[string]interface{}) error {
	notification := &Notification{
		Type:          UpdateRolledBack,
		Timestamp:     time.Now(),
		Message:       fmt.Sprintf("Update rolled back for package %s", packageID),
		TransactionID: transactionID,
		PackageID:     packageID,
		Details:       details,
	}
	return m.SendNotification(notification)

// NotifyComponentUpdated sends a notification that a component has been updated
func (m *NotificationManager) NotifyComponentUpdated(transactionID, packageID, component, componentID string, details map[string]interface{}) error {
	notification := &Notification{
		Type:          ComponentUpdated,
		Timestamp:     time.Now(),
		Message:       fmt.Sprintf("Component %s updated for package %s", component, packageID),
		TransactionID: transactionID,
		PackageID:     packageID,
		Component:     component,
		Details:       details,
	}
	return m.SendNotification(notification)

// NotifyVerificationFailed sends a notification that a verification has failed
func (m *NotificationManager) NotifyVerificationFailed(packageID, reason string, details map[string]interface{}) error {
	notification := &Notification{
		Type:      VerificationFailed,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Verification failed for package %s: %s", packageID, reason),
		PackageID: packageID,
		Details:   details,
	}
