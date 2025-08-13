package main

import (
	"context"
	"fmt"
	"log"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create template sources for different repository types
	sources := []types.TemplateSource{
		// Local filesystem repository
		{
			Name:   "Local Templates",
			Type:   string(repository.LocalFS),
			URL:    "examples/templates",
			Branch: "",
		},
		// GitHub repository
		{
			Name:   "GitHub Templates",
			Type:   string(repository.GitHub),
			URL:    "https://github.com/example/llm-templates",
			Branch: "main",
		},
		// GitLab repository
		{
			Name:   "GitLab Templates",
			Type:   string(repository.GitLab),
			URL:    "https://gitlab.com/example/llm-templates",
			Branch: "main",
		},
		// HTTP repository
		{
			Name:   "HTTP Templates",
			Type:   string(repository.HTTP),
			URL:    "https://example.com/templates",
			Branch: "",
		},
		// Database repository (new)
		{
			Name:   "Database Templates",
			Type:   string(repository.Database),
			URL:    "sqlite://templates.db#templates",
			Branch: "",
		},
		// S3 repository (new)
		{
			Name:   "S3 Templates",
			Type:   string(repository.S3),
			URL:    "s3://my-templates-bucket/templates?region=us-west-2",
			Branch: "",
		},
	}

	// Create repository configurations
	repoConfigs := make(map[string]*repository.Config)
	for _, source := range sources {
		config := repository.NewConfig(
			repository.RepositoryType(source.Type),
			source.Name,
			source.URL,
		)
		config.Branch = source.Branch

		// Add credentials if needed
		if source.Type == string(repository.GitHub) || source.Type == string(repository.GitLab) {
			// Get token from environment variable
			token := os.Getenv("GIT_TOKEN")
			if token != "" {
				config.Password = token
			}
		} else if source.Type == string(repository.S3) {
			// Get AWS credentials from environment variables
			accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
			secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
			if accessKey != "" && secretKey != "" {
				config.Username = accessKey
				config.Password = secretKey
			}
		}

		repoConfigs[source.Name] = config
	}

	// Create repositories
	repos := make(map[string]repository.Repository)
	for name, config := range repoConfigs {
		repo, err := repository.Create(config)
		if err != nil {
			log.Printf("Failed to create repository %s: %v", name, err)
			continue
		}
		repos[name] = repo
	}

	// Connect to repositories
	for name, repo := range repos {
		if err := repo.Connect(ctx); err != nil {
			log.Printf("Failed to connect to repository %s: %v", name, err)
			continue
		}
		defer repo.Disconnect()
		log.Printf("Connected to repository %s", name)
	}

	// List templates from all repositories
	for name, repo := range repos {
		fmt.Printf("Templates from %s:\n", name)
		files, err := repo.ListFiles(ctx, "*.yaml")
		if err != nil {
			log.Printf("Failed to list files from repository %s: %v", name, err)
			continue
		}

		for _, file := range files {
			fmt.Printf("  - %s (size: %d bytes, last modified: %s)\n",
				file.Path, file.Size, file.LastModified.Format(time.RFC3339))
		}
		fmt.Println()
	}

	// Example of loading a template from each repository type
	loadTemplateExample(ctx, repos)
}

func loadTemplateExample(ctx context.Context, repos map[string]repository.Repository) {
	// Create a template loader for each repository
	loaders := make([]management.TemplateLoader, 0, len(repos))
	for name, repo := range repos {
		loader := &templateLoader{
			name: name,
			repo: repo,
		}
		loaders = append(loaders, loader)
	}

	// Load templates
	for _, loader := range loaders {
		fmt.Printf("Loading templates from %s...\n", loader.GetName())
		// In a real implementation, you would use the template manager to load templates
		// This is just a simplified example
		templates, err := loader.LoadFromSource(ctx, types.TemplateSource{
			Name: loader.GetName(),
			Type: string(loader.repo.GetType()),
			URL:  loader.repo.GetURL(),
		}, nil)

		if err != nil {
			log.Printf("Failed to load templates from %s: %v", loader.GetName(), err)
			continue
		}

		fmt.Printf("Loaded %d templates from %s\n", len(templates), loader.GetName())
		for _, template := range templates {
			fmt.Printf("  - %s (version: %s)\n", template.ID, template.Version)
		}
		fmt.Println()
	}
}

// templateLoader is a simple implementation of the TemplateLoader interface
type templateLoader struct {
	name string
	repo repository.Repository
}

func (l *templateLoader) GetName() string {
	return l.name
}

func (l *templateLoader) LoadFromSource(ctx context.Context, source types.TemplateSource, options map[string]interface{}) ([]*format.Template, error) {
	// List YAML files
	files, err := l.repo.ListFiles(ctx, "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Load templates
	templates := make([]*format.Template, 0, len(files))
	for _, file := range files {
		// Get file content
		reader, err := l.repo.GetFile(ctx, file.Path)
		if err != nil {
			log.Printf("Failed to get file %s: %v", file.Path, err)
			continue
		}

		// Read file content
		content := make([]byte, file.Size)
		_, err = reader.Read(content)
		reader.Close()
		if err != nil {
			log.Printf("Failed to read file %s: %v", file.Path, err)
			continue
		}

		// Parse template
		// In a real implementation, you would use a proper YAML parser
		// This is just a simplified example
		template := &format.Template{
			ID:      file.Path,
			Version: "1.0.0",
			Content: string(content),
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func (l *templateLoader) LoadFromFile(ctx context.Context, filePath string, options map[string]interface{}) (*format.Template, error) {
	// Check if file exists
	exists, err := l.repo.FileExists(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Get file content
	reader, err := l.repo.GetFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	defer reader.Close()

	// Get file size
	fileInfo, err := l.repo.GetLastModified(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Read file content
	content := make([]byte, 1024*1024) // Assume max 1MB file size
	n, err := reader.Read(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse template
	// In a real implementation, you would use a proper YAML parser
	// This is just a simplified example
	template := &format.Template{
		ID:      filePath,
		Version: "1.0.0",
		Content: string(content[:n]),
	}

	return template, nil
}

func (l *templateLoader) LoadFromDirectory(ctx context.Context, directoryPath string, options map[string]interface{}) ([]*format.Template, error) {
	// List YAML files in directory
	files, err := l.repo.ListFiles(ctx, "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Filter files by directory
	var dirFiles []repository.FileInfo
	for _, file := range files {
		// Simple directory check - in a real implementation, you would use proper path handling
		if len(file.Path) > len(directoryPath) && file.Path[:len(directoryPath)] == directoryPath {
			dirFiles = append(dirFiles, file)
		}
	}

	// Load templates
	templates := make([]*format.Template, 0, len(dirFiles))
	for _, file := range dirFiles {
		template, err := l.LoadFromFile(ctx, file.Path, options)
		if err != nil {
			log.Printf("Failed to load template from %s: %v", file.Path, err)
			continue
		}
		templates = append(templates, template)
	}

	return templates, nil
}
