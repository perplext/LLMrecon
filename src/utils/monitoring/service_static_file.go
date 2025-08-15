package monitoring

// AddStaticFileMonitor adds a static file handler to the monitoring service
func (s *MonitoringService) AddStaticFileMonitor(fileHandler FileHandlerInterface) *StaticFileMonitor {
	if fileHandler == nil {
		return nil
	}

	// Create a static file monitor
	monitor := NewStaticFileMonitor(fileHandler, s.metricsManager, s.alertManager)
	
	// Set the sample interval from the service configuration
	if s.config != nil && s.config.CollectionInterval > 0 {
		monitor.SetSampleInterval(s.config.CollectionInterval)
	}
	
	// Start the monitor
	monitor.Start()
	
	// Add to the list of monitors
	s.mu.Lock()
	s.staticFileMonitors = append(s.staticFileMonitors, monitor)
	s.mu.Unlock()
	
	// Log the addition
	s.logger.Printf("[INFO] Added static file handler to monitoring service")
	
	return monitor

// GetStaticFileMonitors returns all static file monitors
func (s *MonitoringService) GetStaticFileMonitors() []*StaticFileMonitor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Create a copy of the slice to avoid race conditions
	monitors := make([]*StaticFileMonitor, len(s.staticFileMonitors))
	copy(monitors, s.staticFileMonitors)
	
	return monitors

// GetStaticFileMetrics returns metrics for all static file handlers
func (s *MonitoringService) GetStaticFileMetrics() []*StaticFileMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	metrics := make([]*StaticFileMetrics, 0, len(s.staticFileMonitors))
	
	for _, monitor := range s.staticFileMonitors {
		if monitor != nil {
			metrics = append(metrics, monitor.GetMetrics())
		}
	}
	
	return metrics

// EnableStaticFileMonitoring enables monitoring for all static file handlers
func (s *MonitoringService) EnableStaticFileMonitoring() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, monitor := range s.staticFileMonitors {
		if monitor != nil {
			monitor.Enable()
		}
	}
	
	s.logger.Printf("[INFO] Enabled static file monitoring")

// DisableStaticFileMonitoring disables monitoring for all static file handlers
func (s *MonitoringService) DisableStaticFileMonitoring() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, monitor := range s.staticFileMonitors {
		if monitor != nil {
			monitor.Disable()
		}
	}
	
