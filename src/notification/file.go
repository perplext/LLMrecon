package notification

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileChannel represents a notification channel that writes notifications to a file
type FileChannel struct {
	id        string
	name      string
	filePath  string
	formatter FileFormatter
}

// FileFormatter defines the interface for formatting notifications for file output
type FileFormatter interface {
	Format(notification *Notification) ([]byte, error)
}

// JSONFormatter formats notifications as JSON
type JSONFormatter struct {
	Pretty bool
}

// Format formats a notification as JSON
func (f *JSONFormatter) Format(notification *Notification) ([]byte, error) {
	if f.Pretty {
		return json.MarshalIndent(notification, "", "  ")
	}
	return json.Marshal(notification)
}

// TextFormatter formats notifications as plain text
type TextFormatter struct{}

// Format formats a notification as plain text
func (f *TextFormatter) Format(notification *Notification) ([]byte, error) {
	var severityStr string
	switch notification.Severity {
	case Info:
		severityStr = "INFO"
	case Warning:
		severityStr = "WARNING"
	case Critical:
		severityStr = "CRITICAL"
	default:
		severityStr = "UNKNOWN"
	}

	text := fmt.Sprintf(
		"[%s] [%s] [%s] %s\n%s\n",
		time.Now().Format(time.RFC3339),
		severityStr,
		notification.Type,
		notification.Title,
		notification.Message,
	)

	if len(notification.Metadata) > 0 {
		text += "Metadata:\n"
		for key, value := range notification.Metadata {
			text += fmt.Sprintf("  %s: %s\n", key, value)
		}
	}

	if notification.RequiresAction {
		text += "Action Required: "
		if notification.ActionLabel != "" {
			text += notification.ActionLabel
		} else {
			text += "Please take action"
		}
		if notification.ActionURL != "" {
			text += fmt.Sprintf(" (%s)", notification.ActionURL)
		}
		text += "\n"
	}

	text += fmt.Sprintf("Status: %s\n", notification.Status)
	text += "----------------------------------------\n"

	return []byte(text), nil
}

// NewFileChannel creates a new file notification channel
func NewFileChannel(filePath string, formatter FileFormatter) (*FileChannel, error) {
	if filePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		filePath = filepath.Join(homeDir, ".LLMrecon", "notifications.log")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for notification log: %w", err)
	}

	// Use default formatter if none provided
	if formatter == nil {
		formatter = &TextFormatter{}
	}

	return &FileChannel{
		id:        "file",
		name:      "File",
		filePath:  filePath,
		formatter: formatter,
	}, nil
}

// ID returns the unique identifier for the channel
func (f *FileChannel) ID() string {
	return f.id
}

// Name returns the human-readable name of the channel
func (f *FileChannel) Name() string {
	return f.name
}

// Deliver delivers a notification by writing it to a file
func (f *FileChannel) Deliver(notification *Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	// Format the notification
	data, err := f.formatter.Format(notification)
	if err != nil {
		return fmt.Errorf("failed to format notification: %w", err)
	}

	// Open the file in append mode
	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open notification log file: %w", err)
	}
	defer file.Close()

	// Write the notification to the file
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write notification to log file: %w", err)
	}

	return nil
}

// CanDeliver checks if the channel can deliver the notification
func (f *FileChannel) CanDeliver(notification *Notification) bool {
	// File channel can deliver all notifications
	return true
}
