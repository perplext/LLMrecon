package update


// Checker interface for checking updates
type Checker interface {
	CheckForUpdates() (*VersionInfo, error)
	GetCurrentVersion() string
	GetLatestVersion() (string, error)

// VersionInfo contains version information
type VersionInfo struct {
	Current        string    `json:"current"`
	Latest         string    `json:"latest,omitempty"`
	UpdateAvailable bool     `json:"update_available"`
	ReleaseDate    time.Time `json:"release_date,omitempty"`
	Changelog      string    `json:"changelog_url,omitempty"`
}

// UpdateRequest represents an update operation request
type UpdateRequest struct {
	Components []string `json:"components,omitempty"` // ["core", "templates", "modules"]
	Force      bool     `json:"force,omitempty"`
	DryRun     bool     `json:"dry_run,omitempty"`
}

// UpdateResponse contains update operation results
type UpdateResponse struct {
	Status   string          `json:"status"`
	Updates  []UpdateResult  `json:"updates"`
	Messages []string        `json:"messages,omitempty"`
}

// UpdateResult represents the result of updating a component
type UpdateResult struct {
	Component   string `json:"component"`
	OldVersion  string `json:"old_version"`
	NewVersion  string `json:"new_version"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
}

// Manager interface for managing updates
type Manager interface {
	Checker
	PerformUpdate(request UpdateRequest) (*UpdateResponse, error)
	GetUpdateHistory() ([]UpdateResult, error)
