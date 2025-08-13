// Package interfaces provides interfaces for template management components
package interfaces

import (
	"context"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
)

// TemplateLoaderExtended defines extended interface for loading templates from various sources
type TemplateLoaderExtended interface {
	TemplateLoader
	// LoadTemplate loads a template from a source
	LoadTemplate(ctx context.Context, source string, sourceType string) (*format.Template, error)
	
	// LoadTemplateWithTimeout loads a template with a timeout
	LoadTemplateWithTimeout(ctx context.Context, source string, sourceType string, timeout time.Duration) (*format.Template, error)
	
	// LoadTemplates loads multiple templates from a source
	LoadTemplates(ctx context.Context, source string, sourceType string) ([]*format.Template, error)
	
	// LoadTemplatesWithTimeout loads multiple templates with a timeout
	LoadTemplatesWithTimeout(ctx context.Context, source string, sourceType string, timeout time.Duration) ([]*format.Template, error)
	
	// ClearCache clears the template cache
	ClearCache()
}

// TemplateSource represents the source type for templates
type TemplateSource string

const (
	// FileSource indicates the template is from a file
	FileSource TemplateSource = "file"
	// DirectorySource indicates the template is from a directory
	DirectorySource TemplateSource = "directory"
	// GitHubSource indicates the template is from GitHub
	GitHubSource TemplateSource = "github"
	// GitLabSource indicates the template is from GitLab
	GitLabSource TemplateSource = "gitlab"
	// HTTPSource indicates the template is from HTTP
	HTTPSource TemplateSource = "http"
	// DatabaseSource indicates the template is from a database
	DatabaseSource TemplateSource = "database"
)
