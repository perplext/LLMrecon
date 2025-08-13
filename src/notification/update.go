package notification

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/update"
)

// UpdateNotifier handles notifications related to updates
type UpdateNotifier struct {
	manager *NotificationManager
}

// NewUpdateNotifier creates a new update notifier
func NewUpdateNotifier(manager *NotificationManager) *UpdateNotifier {
	return &UpdateNotifier{
		manager: manager,
	}
}

// NotifyAvailableUpdate creates and delivers a notification for an available update
func (n *UpdateNotifier) NotifyAvailableUpdate(ctx context.Context, versionInfo *update.UpdateVersionInfo) error {
	if versionInfo == nil {
		return fmt.Errorf("version info cannot be nil")
	}

	// Create metadata for the notification
	metadata := map[string]string{
		"currentVersion": versionInfo.CurrentVersion,
		"latestVersion":  versionInfo.LatestVersion,
		"releaseDate":    versionInfo.ReleaseDate,
	}

	if versionInfo.ChangelogURL != "" {
		metadata["changelogURL"] = versionInfo.ChangelogURL
	}

	// Determine severity based on update type
	severity := Info
	if versionInfo.SecurityFixes {
		severity = Warning
	}

	// Create notification
	notification, err := n.manager.CreateNotification(
		UpdateAvailable,
		fmt.Sprintf("Update Available: %s", versionInfo.LatestVersion),
		fmt.Sprintf("A new version of the LLMreconing Tool is available. You are currently using version %s, and version %s is now available.", 
			versionInfo.CurrentVersion, 
			versionInfo.LatestVersion),
		severity,
		false,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create update notification: %w", err)
	}

	// Set action URL if available
	if versionInfo.DownloadURL != "" {
		notification.ActionURL = versionInfo.DownloadURL
		notification.ActionLabel = "Download Update"
		notification.RequiresAction = true
	}

	// Deliver notification
	if err := n.manager.DeliverNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to deliver update notification: %w", err)
	}

	return nil
}

// NotifyRequiredUpdate creates and delivers a notification for a required update
func (n *UpdateNotifier) NotifyRequiredUpdate(ctx context.Context, versionInfo *update.UpdateVersionInfo) error {
	if versionInfo == nil {
		return fmt.Errorf("version info cannot be nil")
	}

	// Create metadata for the notification
	metadata := map[string]string{
		"currentVersion": versionInfo.CurrentVersion,
		"latestVersion":  versionInfo.LatestVersion,
		"releaseDate":    versionInfo.ReleaseDate,
		"required":       "true",
	}

	if versionInfo.ChangelogURL != "" {
		metadata["changelogURL"] = versionInfo.ChangelogURL
	}

	// Determine severity based on update type
	severity := Warning
	if versionInfo.SecurityFixes {
		severity = Critical
	}

	// Create notification
	notification, err := n.manager.CreateNotification(
		UpdateRequired,
		fmt.Sprintf("Required Update: %s", versionInfo.LatestVersion),
		fmt.Sprintf("A required update for the LLMreconing Tool is available. You must update from version %s to version %s to continue using the tool.", 
			versionInfo.CurrentVersion, 
			versionInfo.LatestVersion),
		severity,
		true,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create required update notification: %w", err)
	}

	// Set action URL if available
	if versionInfo.DownloadURL != "" {
		notification.ActionURL = versionInfo.DownloadURL
		notification.ActionLabel = "Download Required Update"
		notification.RequiresAction = true
	}

	// Deliver notification
	if err := n.manager.DeliverNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to deliver required update notification: %w", err)
	}

	return nil
}

// NotifySecurityUpdate creates and delivers a notification for a security update
func (n *UpdateNotifier) NotifySecurityUpdate(ctx context.Context, versionInfo *update.UpdateVersionInfo, details string) error {
	if versionInfo == nil {
		return fmt.Errorf("version info cannot be nil")
	}

	// Create metadata for the notification
	metadata := map[string]string{
		"currentVersion": versionInfo.CurrentVersion,
		"latestVersion":  versionInfo.LatestVersion,
		"releaseDate":    versionInfo.ReleaseDate,
		"securityFixes":  "true",
	}

	if versionInfo.ChangelogURL != "" {
		metadata["changelogURL"] = versionInfo.ChangelogURL
	}

	// Create notification
	notification, err := n.manager.CreateNotification(
		SecurityUpdate,
		fmt.Sprintf("Security Update: %s", versionInfo.LatestVersion),
		fmt.Sprintf("A security update for the LLMreconing Tool is available. It is strongly recommended to update from version %s to version %s.\n\n%s", 
			versionInfo.CurrentVersion, 
			versionInfo.LatestVersion,
			details),
		Critical,
		true,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create security update notification: %w", err)
	}

	// Set action URL if available
	if versionInfo.DownloadURL != "" {
		notification.ActionURL = versionInfo.DownloadURL
		notification.ActionLabel = "Download Security Update"
		notification.RequiresAction = true
	}

	// Deliver notification
	if err := n.manager.DeliverNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to deliver security update notification: %w", err)
	}

	return nil
}

// NotifyUpdateSuccess creates and delivers a notification for a successful update
func (n *UpdateNotifier) NotifyUpdateSuccess(ctx context.Context, fromVersion, toVersion string) error {
	// Create metadata for the notification
	metadata := map[string]string{
		"fromVersion": fromVersion,
		"toVersion":   toVersion,
		"updateTime":  time.Now().Format(time.RFC3339),
	}

	// Create notification
	notification, err := n.manager.CreateNotification(
		UpdateAvailable,
		"Update Successful",
		fmt.Sprintf("The LLMreconing Tool has been successfully updated from version %s to version %s.", 
			fromVersion, 
			toVersion),
		Info,
		false,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create update success notification: %w", err)
	}

	// Deliver notification
	if err := n.manager.DeliverNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to deliver update success notification: %w", err)
	}

	return nil
}

// NotifyUpdateFailure creates and delivers a notification for a failed update
func (n *UpdateNotifier) NotifyUpdateFailure(ctx context.Context, fromVersion, toVersion string, err error) error {
	// Create metadata for the notification
	metadata := map[string]string{
		"fromVersion": fromVersion,
		"toVersion":   toVersion,
		"updateTime":  time.Now().Format(time.RFC3339),
	}

	if err != nil {
		metadata["error"] = err.Error()
	}

	// Create notification
	notification, err := n.manager.CreateNotification(
		UpdateAvailable,
		"Update Failed",
		fmt.Sprintf("The update from version %s to version %s failed. Please try again or contact support if the issue persists.", 
			fromVersion, 
			toVersion),
		Warning,
		true,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create update failure notification: %w", err)
	}

	// Deliver notification
	if err := n.manager.DeliverNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to deliver update failure notification: %w", err)
	}

	return nil
}

// ScheduleUpdateReminder schedules a reminder notification for a pending update
func (n *UpdateNotifier) ScheduleUpdateReminder(ctx context.Context, versionInfo *update.UpdateVersionInfo, reminderTime time.Time) error {
	if versionInfo == nil {
		return fmt.Errorf("version info cannot be nil")
	}

	// Create metadata for the notification
	metadata := map[string]string{
		"currentVersion": versionInfo.CurrentVersion,
		"latestVersion":  versionInfo.LatestVersion,
		"releaseDate":    versionInfo.ReleaseDate,
		"reminderTime":   reminderTime.Format(time.RFC3339),
	}

	if versionInfo.ChangelogURL != "" {
		metadata["changelogURL"] = versionInfo.ChangelogURL
	}

	// Determine severity based on update type
	severity := Info
	if versionInfo.SecurityFixes {
		severity = Warning
	}

	// Create notification
	notification, err := n.manager.CreateNotification(
		UpdateAvailable,
		fmt.Sprintf("Update Reminder: %s", versionInfo.LatestVersion),
		fmt.Sprintf("This is a reminder that an update for the LLMreconing Tool is available. You are currently using version %s, and version %s is now available.", 
			versionInfo.CurrentVersion, 
			versionInfo.LatestVersion),
		severity,
		false,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create update reminder notification: %w", err)
	}

	// Set action URL if available
	if versionInfo.DownloadURL != "" {
		notification.ActionURL = versionInfo.DownloadURL
		notification.ActionLabel = "Download Update"
		notification.RequiresAction = true
	}

	// Schedule notification
	if err := n.manager.ScheduleNotification(notification, reminderTime); err != nil {
		return fmt.Errorf("failed to schedule update reminder notification: %w", err)
	}

	return nil
}
