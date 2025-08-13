package monitoring

// AlertManagerAdapter adapts the existing AlertManager to implement the AlertManagerInterface
type AlertManagerAdapter struct {
	manager interface{
		CheckThreshold(name string, value interface{}, labels map[string]string) error
	}
}

// NewAlertManagerAdapter creates a new adapter for AlertManager
func NewAlertManagerAdapter(manager interface{
	CheckThreshold(name string, value interface{}, labels map[string]string) error
}) *AlertManagerAdapter {
	return &AlertManagerAdapter{
		manager: manager,
	}
}

// CheckThreshold checks if a value exceeds a threshold
func (a *AlertManagerAdapter) CheckThreshold(name string, value interface{}, tags map[string]string) error {
	return a.manager.CheckThreshold(name, value, tags)
}
