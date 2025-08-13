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
	LocalVersion *Version
	
	// RemoteVersion is the remote version
	RemoteVersion *Version
	
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
	localVersion, err := Parse(localTemplate.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse local template version: %w", err)
	}
	
	remoteVersion, err := Parse(remoteTemplate.Version)
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
	
	// Perform diff
	diff := DiffFiles([]*FileInfo{localFileInfo}, []*FileInfo{remoteFileInfo}, a.Options.DiffOptions)
	diff.OldVersion = localVersion
	diff.NewVersion = remoteVersion
	
	// Create analysis result
	result := &AnalysisResult{
		LocalVersion:  localVersion,
		RemoteVersion: remoteVersion,
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
		
		// Get impacted items
		if result.UpdateRequired {
			node := a.DependencyGraph.GetNode(templatePath)
			if node != nil {
				impacted, err := a.DependencyGraph.GetImpactedNodes(node.ID)
				if err == nil {
					result.ImpactedItems = make([]string, len(impacted))
					for i, item := range impacted {
						result.ImpactedItems[i] = item.ID
					}
				}
			}
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
	localModule, err := format.ParseModule(localContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse local module: %w", err)
	}
	
	remoteModule, err := format.ParseModule(remoteContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote module: %w", err)
	}
	
	// Extract versions
	localVersion, err := Parse(localModule.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse local module version: %w", err)
	}
	
	remoteVersion, err := Parse(remoteModule.Version)
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
	
	// Perform diff
	diff := DiffFiles([]*FileInfo{localFileInfo}, []*FileInfo{remoteFileInfo}, a.Options.DiffOptions)
	diff.OldVersion = localVersion
	diff.NewVersion = remoteVersion
	
	// Create analysis result
	result := &AnalysisResult{
		LocalVersion:  localVersion,
		RemoteVersion: remoteVersion,
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
		
		// Get impacted items
		if result.UpdateRequired {
			node := a.DependencyGraph.GetNode(modulePath)
			if node != nil {
				impacted, err := a.DependencyGraph.GetImpactedNodes(node.ID)
				if err == nil {
					result.ImpactedItems = make([]string, len(impacted))
					for i, item := range impacted {
						result.ImpactedItems[i] = item.ID
					}
				}
			}
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
	var allFiles []repository.FileInfo
	
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
		if file.IsDirectory {
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
	// Check if the node already exists
	if a.DependencyGraph.GetNode(rootPath) != nil {
		return nil
	}
	
	// Determine if it's a template or module
	isTemplate := false
	for _, pattern := range a.Options.TemplatePatterns {
		if matchPattern(rootPath, pattern) {
			isTemplate = true
			break
		}
	}
	
	// Get file content
	reader, err := a.LocalRepo.GetFile(ctx, rootPath)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	defer reader.Close()
	
	content, err := ReadFileContent(reader, a.Options.DiffOptions.MaxContentSize)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}
	
	// Parse file
	var dependencies []string
	var version *Version
	var nodeType string
	var name string
	
	if isTemplate {
		template, err := format.ParseTemplate(content)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}
		
		version, err = Parse(template.Version)
		if err != nil {
			return fmt.Errorf("failed to parse template version: %w", err)
		}
		
		nodeType = "template"
		name = template.Name
		
		// Extract dependencies
		for _, dep := range template.Dependencies {
			dependencies = append(dependencies, dep.ModuleID)
		}
	} else {
		module, err := format.ParseModule(content)
		if err != nil {
			return fmt.Errorf("failed to parse module: %w", err)
		}
		
		version, err = Parse(module.Version)
		if err != nil {
			return fmt.Errorf("failed to parse module version: %w", err)
		}
		
		nodeType = "module"
		name = module.Name
		
		// Extract dependencies
		for _, dep := range module.Dependencies {
			dependencies = append(dependencies, dep.ModuleID)
		}
	}
	
	// Add node to graph
	metadata := map[string]interface{}{
		"path": rootPath,
		"type": nodeType,
	}
	
	node := a.DependencyGraph.AddNode(rootPath, name, nodeType, version, metadata)
	
	// Add dependencies
	for _, dep := range dependencies {
		// Check if dependency exists
		depPath := findDependencyPath(ctx, a.LocalRepo, dep, a.Options.ModulePatterns)
		if depPath == "" {
			// Dependency not found, skip
			continue
		}
		
		// Recursively build dependency graph
		if err := a.buildDependencyGraph(ctx, depPath); err != nil {
			// Log error but continue with other dependencies
			fmt.Printf("Error building dependency graph for %s: %v\n", depPath, err)
			continue
		}
		
		// Add dependency to graph
		if err := a.DependencyGraph.AddDependency(rootPath, depPath, "", false); err != nil {
			// Log error but continue with other dependencies
			fmt.Printf("Error adding dependency from %s to %s: %v\n", rootPath, depPath, err)
			continue
		}
	}
	
	return nil
}

// findDependencyPath finds the path of a dependency
func findDependencyPath(ctx context.Context, repo interfaces.Repository, depID string, patterns []string) string {
	// Try each pattern
	for _, pattern := range patterns {
		files, err := repo.ListFiles(ctx, pattern)
		if err != nil {
			continue
		}
		
		for _, file := range files {
			// Skip directories
			if file.IsDir {
				continue
			}
			
			// Get file content
			reader, err := repo.GetFile(ctx, file.Path)
			if err != nil {
				continue
			}
			
			content, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				continue
			}
			
			// Parse module
			module, err := format.ParseModule(content)
			if err != nil {
				continue
			}
			
			// Check if this is the dependency we're looking for
			if module.ID == depID {
				return file.Path
			}
		}
	}
	
	return ""
}
