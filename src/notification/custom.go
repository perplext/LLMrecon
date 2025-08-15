package notification

import (
	"fmt"
)

// CustomChannelConfig represents the configuration for a custom notification channel
type CustomChannelConfig struct {
	ID          string
	Name        string
	DeliverFunc func(notification *Notification) error
	FilterFunc  func(notification *Notification) bool

// CustomChannel represents a user-defined notification channel
type CustomChannel struct {
	id          string
	name        string
	deliverFunc func(notification *Notification) error
	filterFunc  func(notification *Notification) bool

// NewCustomChannel creates a new custom notification channel
func NewCustomChannel(config CustomChannelConfig) (*CustomChannel, error) {
	if config.ID == "" {
		return nil, fmt.Errorf("custom channel ID cannot be empty")
	}

	if config.Name == "" {
		return nil, fmt.Errorf("custom channel name cannot be empty")
	}

	if config.DeliverFunc == nil {
		return nil, fmt.Errorf("custom channel deliver function cannot be nil")
	}

	// Use default filter function if none provided
	filterFunc := config.FilterFunc
	if filterFunc == nil {
		filterFunc = func(notification *Notification) bool {
			return true
		}
	}

	return &CustomChannel{
		id:          config.ID,
		name:        config.Name,
		deliverFunc: config.DeliverFunc,
		filterFunc:  filterFunc,
	}, nil

// ID returns the unique identifier for the channel
func (c *CustomChannel) ID() string {
	return c.id

// Name returns the human-readable name of the channel
func (c *CustomChannel) Name() string {
	return c.name

// Deliver delivers a notification through the custom channel
func (c *CustomChannel) Deliver(notification *Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	return c.deliverFunc(notification)

// CanDeliver checks if the channel can deliver the notification
func (c *CustomChannel) CanDeliver(notification *Notification) bool {
	return c.filterFunc(notification)
