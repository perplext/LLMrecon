package version

import (
	"context"
	"fmt"
	"strings"
	
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
}

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
}

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
}

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
	defer localReader.Close()
	
	localContent, err := ReadFileContent(localReader, a.Options.DiffOptions.MaxContentSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read local template content: %w", err)
	}
	
	// Get remote template
	remoteReader, err := a.RemoteRepo.GetFile(ctx, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote template: %w", err)
	}
	defer remoteReader.Close()
	
	remoteContent, err := ReadFileContent(remoteReader, a.Options.DiffOptions.MaxContentSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote template content: %w", err)
	}
	
	// Parse template versions
	localTemplate, err := format.ParseTemplate(localContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse local template: %w", err)
	}
	
	remoteTemplate, err := format.ParseTemplate(remoteContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote template: %w", err)
	}
	
	// Extract versions
	localVersion, err := Parse(localTemplate.Info.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse local template version: %w", err)
	}
	
	remoteVersion, err := Parse(remoteTemplate.Info.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote template version: %w", err)
	}
	
	// Create file info for diffing
	localLastModified, _ := a.LocalRepo.GetLastModified(ctx, templatePath)
	remoteLastModified, _ := a.RemoteRepo.GetLastModified(ctx, templatePath)
	
	localFileInfo := &FileInfo{
		Path:    templatePath,
		Hash:    ComputeHash(localContent),
		Size:    int64(len(localContent)),
		ModTime: localLastModified,
		Content: localContent,
	}
	
	remoteFileInfo := &FileInfo{
		Path:    templatePath,
		Hash:    ComputeHash(remoteContent),
		Size:    int64(len(remoteContent)),
		ModTime: remoteLastModified,
		Content: remoteContent,
	}
	
	// Create analysis result
	localVersionInfo := &VersionInfo{
		Version: localVersion,
	}
	remoteVersionInfo := &VersionInfo{
		Version: remoteVersion,
	}
	
	// Perform diff
	diff := DiffFiles([]*FileInfo{localFileInfo}, []*FileInfo{remoteFileInfo}, a.Options.DiffOptions)
	diff.LocalVersion = localVersionInfo
	diff.RemoteVersion = remoteVersionInfo
	
	result := &AnalysisResult{
		LocalVersion:  localVersionInfo,
		RemoteVersion: remoteVersionInfo,
		Diff:          diff,
		AnalysisTime:  time.Now(),
	}
	
	// Check if update is required
	result.UpdateRequired = remoteVersion.GreaterThan(localVersion)
	
	// Build dependency graph if requested
	if a.Options.IncludeDependencies {
		if err := a.buildDependencyGraph(ctx, templatePath); err != nil {
			return nil, fmt.Errorf("failed to build dependency graph: %w", err)
		}
		
		result.DependencyGraph = a.DependencyGraph
		
		// Get impacted items - stub implementation
		if result.UpdateRequired {
			// TODO: Implement dependency graph node retrieval
			result.ImpactedItems = []string{} // Empty for now
		}
	}
	
	return result, nil
}

// AnalyzeModule analyzes a module
func (a *Analyzer) AnalyzeModule(ctx context.Context, modulePath string) (*AnalysisResult, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.Options.Timeout)
	defer cancel()
	
	// Check if module exists in local repository
	localExists, err := a.LocalRepo.FileExists(ctx, modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if module exists in local repository: %w", err)
	}
	
	if !localExists {
		return nil, fmt.Errorf("module %s does not exist in local repository", modulePath)
	}
	
	// Check if module exists in remote repository
	remoteExists, err := a.RemoteRepo.FileExists(ctx, modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if module exists in remote repository: %w", err)
	}
	
	if !remoteExists {
		return nil, fmt.Errorf("module %s does not exist in remote repository", modulePath)
	}
	
	// Get local module
	localReader, err := a.LocalRepo.GetFile(ctx, modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get local module: %w", err)
	}
	defer localReader.Close()
	
	localContent, err := ReadFileContent(localReader, a.Options.DiffOptions.MaxContentSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read local module content: %w", err)
	}
	
	// Get remote module
	remoteReader, err := a.RemoteRepo.GetFile(ctx, modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote module: %w", err)
	}
	defer remoteReader.Close()
	
	remoteContent, err := ReadFileContent(remoteReader, a.Options.DiffOptions.MaxContentSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote module content: %w", err)
	}
	
	// Parse module versions
	localModule, err := format.ParseTemplate(localContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse local module: %w", err)
	}
	
	remoteModule, err := format.ParseTemplate(remoteContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote module: %w", err)
	}
	
	// Extract versions
	localVersion, err := Parse(localModule.Info.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse local module version: %w", err)
	}
	
	remoteVersion, err := Parse(remoteModule.Info.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote module version: %w", err)
	}
	
	// Create file info for diffing
	localLastModified, _ := a.LocalRepo.GetLastModified(ctx, modulePath)
	remoteLastModified, _ := a.RemoteRepo.GetLastModified(ctx, modulePath)
	
	localFileInfo := &FileInfo{
		Path:    modulePath,
		Hash:    ComputeHash(localContent),
		Size:    int64(len(localContent)),
		ModTime: localLastModified,
		Content: localContent,
	}
	
	remoteFileInfo := &FileInfo{
		Path:    modulePath,
		Hash:    ComputeHash(remoteContent),
		Size:    int64(len(remoteContent)),
		ModTime: remoteLastModified,
		Content: remoteContent,
	}
	
	// Create analysis result
	localVersionInfo := &VersionInfo{
		Version: localVersion,
	}
	remoteVersionInfo := &VersionInfo{
		Version: remoteVersion,
	}
	
	// Perform diff
	diff := DiffFiles([]*FileInfo{localFileInfo}, []*FileInfo{remoteFileInfo}, a.Options.DiffOptions)
	diff.LocalVersion = localVersionInfo
	diff.RemoteVersion = remoteVersionInfo
	
	result := &AnalysisResult{
		LocalVersion:  localVersionInfo,
		RemoteVersion: remoteVersionInfo,
		Diff:          diff,
		AnalysisTime:  time.Now(),
	}
	
	// Check if update is required
	result.UpdateRequired = remoteVersion.GreaterThan(localVersion)
	
	// Build dependency graph if requested
	if a.Options.IncludeDependencies {
		if err := a.buildDependencyGraph(ctx, modulePath); err != nil {
			return nil, fmt.Errorf("failed to build dependency graph: %w", err)
		}
		
		result.DependencyGraph = a.DependencyGraph
		
		// Get impacted items - stub implementation
		if result.UpdateRequired {
			// TODO: Implement dependency graph node retrieval for modules
			result.ImpactedItems = []string{} // Empty for now
		}
	}
	
	return result, nil
}

// AnalyzeAll analyzes all templates and modules
func (a *Analyzer) AnalyzeAll(ctx context.Context) (map[string]*AnalysisResult, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.Options.Timeout)
	defer cancel()
	
	// Get all templates and modules from local repository
	var allFiles []interfaces.FileInfo
	
	// Get templates
	for _, pattern := range a.Options.TemplatePatterns {
		files, err := a.LocalRepo.ListFiles(ctx, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to list template files: %w", err)
		}
		allFiles = append(allFiles, files...)
	}
	
	// Get modules
	for _, pattern := range a.Options.ModulePatterns {
		files, err := a.LocalRepo.ListFiles(ctx, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to list module files: %w", err)
		}
		allFiles = append(allFiles, files...)
	}
	
	// Analyze each file
	results := make(map[string]*AnalysisResult)
	
	for _, file := range allFiles {
		// Skip directories
		if file.IsDir {
			continue
		}
		
		// Skip non-YAML files
		ext := strings.ToLower(filepath.Ext(file.Path))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		
		// Determine if it's a template or module
		isTemplate := false
		for _, pattern := range a.Options.TemplatePatterns {
			if matchPattern(file.Path, pattern) {
				isTemplate = true
				break
			}
		}
		
		// Analyze file
		var result *AnalysisResult
		var err error
		
		if isTemplate {
			result, err = a.AnalyzeTemplate(ctx, file.Path)
		} else {
			result, err = a.AnalyzeModule(ctx, file.Path)
		}
		
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("Error analyzing %s: %v\n", file.Path, err)
			continue
		}
		
		results[file.Path] = result
	}
	
	return results, nil
}

// buildDependencyGraph builds a dependency graph for templates and modules
func (a *Analyzer) buildDependencyGraph(ctx context.Context, rootPath string) error {
	// TODO: Implement dependency graph building
	// For now, just return nil as a stub
	return nil
}

// findDependencyPath finds the path of a dependency
func findDependencyPath(ctx context.Context, repo interfaces.Repository, depID string, patterns []string) string {
	// TODO: Implement dependency path finding
	// For now, just return empty string as stub
	return ""
}
