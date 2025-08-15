package version

import (
	"context"
	"fmt"
	"strings"
	"time"
	
	"github.com/perplext/LLMrecon/src/repository/interfaces"
	"github.com/perplext/LLMrecon/src/template/format"
)

// AnalyzerOptions represents options for the version analyzer
type AnalyzerOptions struct {
	// DiffOptions are the options for file diffing
	DiffOptions *DiffOptions
	
	// IncludeDependencies determines if dependencies should be analyzed
	IncludeDependencies bool
	
	// MaxConcurrentOperations is the maximum number of concurrent operations
	MaxConcurrentOperations int
	
	// Timeout is the timeout for operations
	Timeout time.Duration
	
	// TemplatePatterns are the patterns for template files
	TemplatePatterns []string
	
	// ModulePatterns are the patterns for module files
	ModulePatterns []string

// DefaultAnalyzerOptions returns the default analyzer options
func DefaultAnalyzerOptions() *AnalyzerOptions {
	return &AnalyzerOptions{
		DiffOptions:             DefaultDiffOptions(),
		IncludeDependencies:     true,
		MaxConcurrentOperations: 5,
		Timeout:                 5 * time.Minute,
		TemplatePatterns:        []string{"*.yaml", "*.yml"},
		ModulePatterns:          []string{"*.yaml", "*.yml"},
	}
}

// Analyzer analyzes template and module versions
type Analyzer struct {
	// Options are the analyzer options
	Options *AnalyzerOptions
	
	// LocalRepo is the local repository
	LocalRepo interfaces.Repository
	
	// RemoteRepo is the remote repository
	RemoteRepo interfaces.Repository
	
	// DependencyGraph is the dependency graph
	DependencyGraph *DependencyGraph

// NewAnalyzer creates a new version analyzer
func NewAnalyzer(localRepo, remoteRepo interfaces.Repository, options *AnalyzerOptions) *Analyzer {
	if options == nil {
		options = DefaultAnalyzerOptions()
	}
	
	return &Analyzer{
		Options:         options,
		LocalRepo:       localRepo,
		RemoteRepo:      remoteRepo,
		DependencyGraph: NewDependencyGraph(),
	}
}

// AnalysisResult represents the result of a version analysis
type AnalysisResult struct {
	// LocalVersion is the local version
	LocalVersion *VersionInfo
	
	// RemoteVersion is the remote version
	RemoteVersion *VersionInfo
	
	// Diff is the difference between local and remote versions
	Diff *DiffResult
	
	// DependencyGraph is the dependency graph
	DependencyGraph *DependencyGraph
	
	// UpdateRequired indicates if an update is required
	UpdateRequired bool
	
	// ImpactedItems are the items impacted by the update
	ImpactedItems []string
	
	// AnalysisTime is the time the analysis was performed
	AnalysisTime time.Time

// AnalyzeTemplate analyzes a template
func (a *Analyzer) AnalyzeTemplate(ctx context.Context, templatePath string) (*AnalysisResult, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.Options.Timeout)
	defer cancel()
	
	// Check if template exists in local repository
	localExists, err := a.LocalRepo.FileExists(ctx, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if template exists in local repository: %w", err)
	}
	
	if !localExists {
		return nil, fmt.Errorf("template %s does not exist in local repository", templatePath)
	}
	
	// Check if template exists in remote repository
	remoteExists, err := a.RemoteRepo.FileExists(ctx, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if template exists in remote repository: %w", err)
	}
	
	if !remoteExists {
		return nil, fmt.Errorf("template %s does not exist in remote repository", templatePath)
	}
	
	// Get local template
	localReader, err := a.LocalRepo.GetFile(ctx, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get local template: %w", err)
	}
	defer func() { 
		if err := localReader.Close(); err != nil { 
			fmt.Printf("Failed to close: %v\n", err) 
		} 
	}()
		
	localContent, err := ReadFileContent(localReader, a.Options.DiffOptions.MaxContentSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read local template content: %w", err)
	}
	
	// Get remote template
	remoteReader, err := a.RemoteRepo.GetFile(ctx, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote template: %w", err)
	}
	defer func() { 
		if err := remoteReader.Close(); err != nil { 
			fmt.Printf("Failed to close: %v\n", err) 
		} 
	}()
	
	remoteContent, err := ReadFileContent(remoteReader, a.Options.DiffOptions.MaxContentSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote template content: %w", err)
	}
	
	// Create version info
	localVersion := &VersionInfo{
		Version: "local",
		Content: localContent,
	}
	
	remoteVersion := &VersionInfo{
		Version: "remote",
		Content: remoteContent,
	}
	
	// Compare versions
	diff := CompareVersions(localVersion, remoteVersion, a.Options.DiffOptions)
	
	// Build result
	result := &AnalysisResult{
		LocalVersion:    localVersion,
		RemoteVersion:   remoteVersion,
		Diff:           diff,
		DependencyGraph: a.DependencyGraph,
		UpdateRequired:  diff.HasChanges,
		ImpactedItems:   []string{templatePath},
		AnalysisTime:   time.Now(),
	}
	
	return result, nil

// AnalyzeModule analyzes a module
func (a *Analyzer) AnalyzeModule(ctx context.Context, modulePath string) (*AnalysisResult, error) {
	// Similar to AnalyzeTemplate
	return a.AnalyzeTemplate(ctx, modulePath)

// AnalyzeAll analyzes all templates and modules
func (a *Analyzer) AnalyzeAll(ctx context.Context) ([]*AnalysisResult, error) {
	var results []*AnalysisResult
	
	// Find all templates
	templates, err := a.findFiles(ctx, a.Options.TemplatePatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to find templates: %w", err)
	}
	
	// Analyze each template
	for _, template := range templates {
		result, err := a.AnalyzeTemplate(ctx, template)
		if err != nil {
			// Log error but continue
			fmt.Printf("Failed to analyze template %s: %v\n", template, err)
			continue
		}
		results = append(results, result)
	}
	
	// Find all modules
	modules, err := a.findFiles(ctx, a.Options.ModulePatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to find modules: %w", err)
	}
	
	// Analyze each module
	for _, module := range modules {
		result, err := a.AnalyzeModule(ctx, module)
		if err != nil {
			// Log error but continue
			fmt.Printf("Failed to analyze module %s: %v\n", module, err)
			continue
		}
		results = append(results, result)
	}
	
	return results, nil

// findFiles finds files matching patterns
func (a *Analyzer) findFiles(ctx context.Context, patterns []string) ([]string, error) {
	var files []string
	
	for _, pattern := range patterns {
		matches, err := a.LocalRepo.ListFiles(ctx, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to list files for pattern %s: %w", pattern, err)
		}
		files = append(files, matches...)
	}
	
}
}
}
}
