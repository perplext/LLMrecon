package bundle


// Manager interface for bundle operations
type Manager interface {
	ListBundles() ([]Info, error)
	GetBundle(id string) (*Info, error)
	CreateBundle(request CreateRequest) (*Info, error)
	DeleteBundle(id string) error
	ExportBundle(request ExportRequest) (*OperationResult, error)
	ImportBundle(request ImportRequest) (*OperationResult, error)
}

// Info represents metadata about a bundle
type Info struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Tags        []string               `json:"tags,omitempty"`
	Templates   []string               `json:"templates"`
	Modules     []string               `json:"modules,omitempty"`
	Size        int64                  `json:"size"`
	Checksum    string                 `json:"checksum"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRequest represents a request to create a bundle
type CreateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Templates   []string `json:"templates,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Modules     []string `json:"modules,omitempty"`
}

// ExportRequest represents a request to export a bundle
type ExportRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Templates   []string `json:"templates,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Modules     []string `json:"modules,omitempty"`
	Format      string   `json:"format,omitempty"` // "zip", "tar.gz"
	Compress    bool     `json:"compress,omitempty"`
}

// ImportRequest represents a request to import a bundle
type ImportRequest struct {
	Source        string `json:"source"` // file path or URL
	ValidateOnly  bool   `json:"validate_only,omitempty"`
	Overwrite     bool   `json:"overwrite,omitempty"`
	SkipConflicts bool   `json:"skip_conflicts,omitempty"`
}

// OperationResult represents the result of a bundle operation
type OperationResult struct {
	BundleID   string            `json:"bundle_id,omitempty"`
	Status     string            `json:"status"`
	Message    string            `json:"message,omitempty"`
	Templates  []string          `json:"templates,omitempty"`
	Modules    []string          `json:"modules,omitempty"`
	Conflicts  []string          `json:"conflicts,omitempty"`
	Errors     []string          `json:"errors,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}