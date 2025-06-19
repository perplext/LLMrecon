package notification

import (
	"fmt"
	"strings"
	"time"
)

// ConsoleChannel represents a notification channel that displays notifications in the console
type ConsoleChannel struct {
	id   string
	name string
}

// NewConsoleChannel creates a new console notification channel
func NewConsoleChannel() *ConsoleChannel {
	return &ConsoleChannel{
		id:   "console",
		name: "Console",
	}
}

// ID returns the unique identifier for the channel
func (c *ConsoleChannel) ID() string {
	return c.id
}

// Name returns the human-readable name of the channel
func (c *ConsoleChannel) Name() string {
	return c.name
}

// Deliver delivers a notification through the console
func (c *ConsoleChannel) Deliver(notification *Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	// Format the notification based on severity
	var prefix string
	switch notification.Severity {
	case Info:
		prefix = "[INFO]"
	case Warning:
		prefix = "[WARNING]"
	case Critical:
		prefix = "[CRITICAL]"
	default:
		prefix = "[NOTIFICATION]"
	}

	// Format the notification message
	var builder strings.Builder
	
	// Add a separator line
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("=", 80))
	builder.WriteString("\n")
	
	// Add the notification header
	builder.WriteString(fmt.Sprintf("%s %s\n", prefix, notification.Title))
	builder.WriteString(strings.Repeat("-", 80))
	builder.WriteString("\n")
	
	// Add the notification message
	builder.WriteString(fmt.Sprintf("%s\n", notification.Message))
	
	// Add metadata if available
	if len(notification.Metadata) > 0 {
		builder.WriteString("\nDetails:\n")
		for key, value := range notification.Metadata {
			builder.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
		}
	}
	
	// Add action information if required
	if notification.RequiresAction {
		builder.WriteString("\nAction Required: ")
		if notification.ActionLabel != "" {
			builder.WriteString(notification.ActionLabel)
		} else {
			builder.WriteString("Please take action")
		}
		
		if notification.ActionURL != "" {
			builder.WriteString(fmt.Sprintf(" (%s)", notification.ActionURL))
		}
		builder.WriteString("\n")
	}
	
	// Add timestamp
	builder.WriteString(fmt.Sprintf("\nTimestamp: %s\n", time.Now().Format(time.RFC3339)))
	
	// Add a separator line
	builder.WriteString(strings.Repeat("=", 80))
	builder.WriteString("\n")
	
	// Print the notification
	fmt.Print(builder.String())
	
	return nil
}

// CanDeliver checks if the channel can deliver the notification
func (c *ConsoleChannel) CanDeliver(notification *Notification) bool {
	// Console channel can deliver all notifications
	return true
}
