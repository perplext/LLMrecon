package notification

import (
	"context"
	"fmt"
	"sync"
)

// NotificationScheduler handles scheduling and processing of notifications
type NotificationScheduler struct {
	manager       *NotificationManager
	ticker        *time.Ticker
	done          chan struct{}
	wg            sync.WaitGroup
	checkInterval time.Duration
	running       bool
	mutex         sync.Mutex
}

// NewNotificationScheduler creates a new notification scheduler
func NewNotificationScheduler(manager *NotificationManager, checkInterval time.Duration) *NotificationScheduler {
	if checkInterval <= 0 {
		checkInterval = 5 * time.Minute // Default check interval
	}

	return &NotificationScheduler{
		manager:       manager,
		checkInterval: checkInterval,
		done:          make(chan struct{}),
	}
}

// Start starts the notification scheduler
func (s *NotificationScheduler) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.ticker = time.NewTicker(s.checkInterval)
	s.running = true

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(ctx)
	}()

	return nil
}

// Stop stops the notification scheduler
func (s *NotificationScheduler) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	s.ticker.Stop()
	close(s.done)
	s.wg.Wait()
	s.running = false

	return nil
}

// run runs the notification scheduler
func (s *NotificationScheduler) run(ctx context.Context) {
	// Process scheduled notifications immediately on start
	if err := s.processNotifications(ctx); err != nil {
		fmt.Printf("Error processing scheduled notifications: %v\n", err)
	}

	for {
		select {
		case <-s.ticker.C:
			if err := s.processNotifications(ctx); err != nil {
				fmt.Printf("Error processing scheduled notifications: %v\n", err)
			}
		case <-s.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// processNotifications processes scheduled notifications and purges expired ones
func (s *NotificationScheduler) processNotifications(ctx context.Context) error {
	// Process scheduled notifications
	if err := s.manager.ProcessScheduledNotifications(); err != nil {
		return fmt.Errorf("failed to process scheduled notifications: %w", err)
	}

	// Purge expired notifications
	if err := s.manager.PurgeExpiredNotifications(); err != nil {
		return fmt.Errorf("failed to purge expired notifications: %w", err)
	}

	return nil
}

// ScheduleRecurringNotification schedules a notification to recur at a specified interval
func (s *NotificationScheduler) ScheduleRecurringNotification(
	ctx context.Context,
	notificationType NotificationType,
	title string,
	message string,
	severity SeverityLevel,
	requiresAction bool,
	metadata map[string]string,
	startTime time.Time,
	interval time.Duration,
	count int,
) error {
	// Validate parameters
	if interval <= 0 {
		return fmt.Errorf("interval must be greater than zero")
	}

	if count < 0 {
		return fmt.Errorf("count must be greater than or equal to zero")
	}

	// Schedule the first notification
	notification, err := s.manager.CreateNotification(
		notificationType,
		title,
		message,
		severity,
		requiresAction,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Schedule the first occurrence
	if err := s.manager.ScheduleNotification(notification, startTime); err != nil {
		return fmt.Errorf("failed to schedule notification: %w", err)
	}

	// If count is 0, schedule indefinitely (until manually stopped)
	// If count is 1, we've already scheduled the only occurrence
	// If count > 1, schedule the remaining occurrences
	if count != 1 {
		// Start a goroutine to schedule future occurrences
		go func() {
			remaining := count - 1
			nextTime := startTime.Add(interval)

			for count == 0 || remaining > 0 {
				select {
				case <-time.After(time.Until(nextTime)):
					// Create a new notification for this occurrence
					notification, err := s.manager.CreateNotification(
						notificationType,
						title,
						message,
						severity,
						requiresAction,
						metadata,
					)
					if err != nil {
						fmt.Printf("Failed to create recurring notification: %v\n", err)
						return
					}

					// Deliver the notification immediately
					if err := s.manager.DeliverNotification(notification.ID); err != nil {
						fmt.Printf("Failed to deliver recurring notification: %v\n", err)
					}

					// Update for next iteration
					nextTime = nextTime.Add(interval)
					if count > 0 {
						remaining--
					}
				case <-s.done:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return nil
}
