package update

import (
	"context"
	"fmt"
)

// ExtendedUpdateNotifier extends the UpdateNotifier interface with additional notification methods
type ExtendedUpdateNotifier interface {
	UpdateNotifier
	NotifyAvailableUpdate(ctx context.Context, versionInfo *UpdateVersionInfo) error
	NotifyRequiredUpdate(ctx context.Context, versionInfo *UpdateVersionInfo) error
	NotifySecurityUpdate(ctx context.Context, versionInfo *UpdateVersionInfo, details string) error
}

// NotificationIntegration provides integration between the update system and notification system
type NotificationIntegration struct {
	notifier ExtendedUpdateNotifier
}

// NewNotificationIntegration creates a new notification integration
func NewNotificationIntegration(notifier ExtendedUpdateNotifier) *NotificationIntegration {
	return &NotificationIntegration{
		notifier: notifier,
	}
}

// HandleUpdateCheck processes the results of a version check and creates notifications
func (n *NotificationIntegration) HandleUpdateCheck(ctx context.Context, versionInfo *UpdateVersionInfo) error {
	if versionInfo == nil {
		return fmt.Errorf("version info cannot be nil")
	}

	// If no update is available, do nothing
	if !versionInfo.UpdateAvailable {
		return nil
	}

	// Create appropriate notifications based on update type
	var err error
	if versionInfo.RequiredUpdate {
		err = n.notifier.NotifyRequiredUpdate(ctx, versionInfo)
	} else if versionInfo.SecurityFixes {
		err = n.notifier.NotifySecurityUpdate(ctx, versionInfo, "Security fixes are available in this update.")
	} else {
		err = n.notifier.NotifyAvailableUpdate(ctx, versionInfo)
	}

	if err != nil {
		return fmt.Errorf("failed to create update notification: %w", err)
	}

	return nil
}

// NotifyUpdateSuccess creates a notification for a successful update
func (n *NotificationIntegration) NotifyUpdateSuccess(ctx context.Context, fromVersion, toVersion string) error {
	return n.notifier.NotifyUpdateSuccess(ctx, fromVersion, toVersion)
}

// NotifyUpdateFailure creates a notification for a failed update
func (n *NotificationIntegration) NotifyUpdateFailure(ctx context.Context, fromVersion, toVersion string, err error) error {
	return n.notifier.NotifyUpdateFailure(ctx, fromVersion, toVersion, err)
}
