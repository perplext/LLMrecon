package cmd

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/notification"
	"github.com/perplext/LLMrecon/src/update"
	"github.com/spf13/cobra"
)

var (
	notificationManager *notification.NotificationManager
	updateNotifier      *notification.UpdateNotifier
	scheduler           *notification.NotificationScheduler
)

// initNotificationSystem initializes the notification system
func initNotificationSystem() error {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create the notification manager
	storageDir := filepath.Join(homeDir, ".LLMrecon")
	manager, err := notification.NewNotificationManager(storageDir)
	if err != nil {
		return fmt.Errorf("failed to create notification manager: %w", err)
	}

	// Register notification channels
	// Console channel for in-app notifications
	consoleChannel := notification.NewConsoleChannel()
	manager.RegisterChannel(consoleChannel)

	// File channel for logging notifications
	logFilePath := filepath.Join(storageDir, "notifications.log")
	fileChannel, err := notification.NewFileChannel(logFilePath, &notification.TextFormatter{})
	if err != nil {
		return fmt.Errorf("failed to create file channel: %w", err)
	}
	manager.RegisterChannel(fileChannel)

	// Create the update notifier
	updateNotifier = notification.NewUpdateNotifier(manager)

	// Create the notification scheduler
	scheduler = notification.NewNotificationScheduler(manager, 5*time.Minute)
	if err := scheduler.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start notification scheduler: %w", err)
	}

	// Store the manager for later use
	notificationManager = manager

	return nil

// notificationCmd represents the notification command
var notificationCmd = &cobra.Command{
	Use:   "notification",
	Short: "Manage notifications",
	Long:  `Manage notifications for the LLMreconing Tool.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize the notification system if not already initialized
		if notificationManager == nil {
			if err := initNotificationSystem(); err != nil {
				return err
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},

// listCmd represents the notification list command
var listNotificationsCmd = &cobra.Command{
	Use:   "list",
	Short: "List notifications",
	Long:  `List all notifications or filter by status.`,
	Run: func(cmd *cobra.Command, args []string) {
		status, _ := cmd.Flags().GetString("status")
		showHistory, _ := cmd.Flags().GetBool("history")

		if showHistory {
			// Get notification history
			history := notificationManager.GetNotificationHistory()
			if len(history) == 0 {
				fmt.Println("No notification history found.")
				return
			}

			fmt.Println("Notification History:")
			fmt.Println("---------------------")
			for i, notification := range history {
				fmt.Printf("%d. [%s] %s - %s\n", i+1, notification.Severity, notification.Title, notification.CreatedAt.Format(time.RFC3339))
				fmt.Printf("   Status: %s\n", notification.Status)
				if !notification.DeliveredAt.IsZero() {
					fmt.Printf("   Delivered: %s\n", notification.DeliveredAt.Format(time.RFC3339))
				}
				if !notification.AcknowledgedAt.IsZero() {
					fmt.Printf("   Acknowledged: %s\n", notification.AcknowledgedAt.Format(time.RFC3339))
				}
				fmt.Println()
			}
			return
		}

		var notifications []*notification.Notification
		switch status {
		case "pending":
			notifications = notificationManager.GetPendingNotifications()
		case "unacknowledged":
			notifications = notificationManager.GetUnacknowledgedNotifications()
		default:
			// Get all notifications
			notifications = notificationManager.GetUnacknowledgedNotifications()
		}

		if len(notifications) == 0 {
			fmt.Println("No notifications found.")
			return
		}

		fmt.Println("Notifications:")
		fmt.Println("--------------")
		for i, notification := range notifications {
			fmt.Printf("%d. [%s] %s - %s\n", i+1, notification.Severity, notification.Title, notification.CreatedAt.Format(time.RFC3339))
			fmt.Printf("   ID: %s\n", notification.ID)
			fmt.Printf("   Status: %s\n", notification.Status)
			fmt.Println()
		}
	},

// showNotificationCmd represents the notification show command
var showNotificationCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show notification details",
	Long:  `Show details of a specific notification.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		notification, err := notificationManager.GetNotification(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Notification: %s\n", notification.Title)
		fmt.Printf("ID: %s\n", notification.ID)
		fmt.Printf("Type: %s\n", notification.Type)
		fmt.Printf("Severity: %s\n", notification.Severity)
		fmt.Printf("Status: %s\n", notification.Status)
		fmt.Printf("Created: %s\n", notification.CreatedAt.Format(time.RFC3339))
		if !notification.DeliveredAt.IsZero() {
			fmt.Printf("Delivered: %s\n", notification.DeliveredAt.Format(time.RFC3339))
		}
		if !notification.AcknowledgedAt.IsZero() {
			fmt.Printf("Acknowledged: %s\n", notification.AcknowledgedAt.Format(time.RFC3339))
		}
		if !notification.ScheduledFor.IsZero() {
			fmt.Printf("Scheduled For: %s\n", notification.ScheduledFor.Format(time.RFC3339))
		}
		fmt.Println()
		fmt.Println("Message:")
		fmt.Println(notification.Message)
		fmt.Println()

		if len(notification.Metadata) > 0 {
			fmt.Println("Metadata:")
			for key, value := range notification.Metadata {
				fmt.Printf("  %s: %s\n", key, value)
			}
			fmt.Println()
		}

		if notification.RequiresAction {
			fmt.Println("Action Required:")
			if notification.ActionLabel != "" {
				fmt.Printf("  %s\n", notification.ActionLabel)
			} else {
				fmt.Println("  Please take action")
			}
			if notification.ActionURL != "" {
				fmt.Printf("  URL: %s\n", notification.ActionURL)
			}
			fmt.Println()
		}
	},

// acknowledgeNotificationCmd represents the notification acknowledge command
var acknowledgeNotificationCmd = &cobra.Command{
	Use:   "acknowledge [id]",
	Short: "Acknowledge a notification",
	Long:  `Mark a notification as acknowledged.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		if err := notificationManager.AcknowledgeNotification(id); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Notification %s acknowledged.\n", id)
	},

// dismissNotificationCmd represents the notification dismiss command
var dismissNotificationCmd = &cobra.Command{
	Use:   "dismiss [id]",
	Short: "Dismiss a notification",
	Long:  `Mark a notification as dismissed.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		if err := notificationManager.DismissNotification(id); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Notification %s dismissed.\n", id)
	},

// clearHistoryCmd represents the notification clear-history command
var clearHistoryCmd = &cobra.Command{
	Use:   "clear-history",
	Short: "Clear notification history",
	Long:  `Clear the notification history.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := notificationManager.ClearNotificationHistory(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("Notification history cleared.")
	},
	

// checkUpdatesCmd represents the notification check-updates command
var checkUpdatesCmd = &cobra.Command{
	Use:   "check-updates",
	Short: "Check for updates and notify",
	Long:  `Check for updates and create notifications for available updates.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Create a version checker
		checker, err := update.NewVersionChecker(ctx)
		if err != nil {
			fmt.Printf("Error creating version checker: %v\n", err)
			return
		}

		// Check for updates
		versionInfo, err := checker.CheckVersion(ctx)
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			return
		}

		// Notify about updates
		if versionInfo.UpdateAvailable {
			if versionInfo.RequiredUpdate {
				if err := updateNotifier.NotifyRequiredUpdate(ctx, versionInfo); err != nil {
					fmt.Printf("Error creating required update notification: %v\n", err)
					return
				}
				fmt.Println("Required update notification created.")
			} else if versionInfo.SecurityFixes {
				if err := updateNotifier.NotifySecurityUpdate(ctx, versionInfo, "Security fixes are available in this update."); err != nil {
					fmt.Printf("Error creating security update notification: %v\n", err)
					return
				}
				fmt.Println("Security update notification created.")
			} else {
				if err := updateNotifier.NotifyAvailableUpdate(ctx, versionInfo); err != nil {
					fmt.Printf("Error creating update notification: %v\n", err)
					return
				}
				fmt.Println("Update notification created.")
			}
		} else {
			fmt.Println("No updates available.")
		}
	},

func init() {
	// Add to root command when it's available
	// RootCmd.AddCommand(notificationCmd)

	// Add subcommands
	notificationCmd.AddCommand(listNotificationsCmd)
	notificationCmd.AddCommand(showNotificationCmd)
	notificationCmd.AddCommand(acknowledgeNotificationCmd)
	notificationCmd.AddCommand(dismissNotificationCmd)
	notificationCmd.AddCommand(clearHistoryCmd)
	notificationCmd.AddCommand(checkUpdatesCmd)

	// Add flags
	listNotificationsCmd.Flags().String("status", "", "Filter notifications by status (pending, unacknowledged)")
	listNotificationsCmd.Flags().Bool("history", false, "Show notification history")
