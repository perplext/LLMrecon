package main

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/repository"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create repository manager
	manager := repository.NewManager()

	// Example 1: Local repository
	fmt.Println("=== Example 1: Local Repository ===")
	localConfig := repository.NewConfig(repository.LocalFS, "local-repo", "./examples")
	localRepo, err := repository.Create(localConfig)
	if err != nil {
		fmt.Printf("Failed to create local repository: %v\n", err)
		return
	}

	// Add to manager
	if err := manager.AddRepository(localRepo); err != nil {
		fmt.Printf("Failed to add local repository to manager: %v\n", err)
		return
	}

	// Connect to repository
	if err := localRepo.Connect(ctx); err != nil {
		fmt.Printf("Failed to connect to local repository: %v\n", err)
		return
	}
	defer localRepo.Disconnect()

	// List files
	fmt.Println("Files in local repository:")
	files, err := localRepo.ListFiles(ctx, "*.go")
	if err != nil {
		fmt.Printf("Failed to list files: %v\n", err)
	} else {
		for _, file := range files {
			fmt.Printf("- %s (Size: %d bytes, Last Modified: %s)\n", 
				file.Path, file.Size, file.LastModified.Format(time.RFC3339))
		}
	}
	fmt.Println()

	// Example 2: HTTP repository (read-only)
	fmt.Println("=== Example 2: HTTP Repository ===")
	// Using a public HTTP server for demonstration
	httpConfig := repository.NewConfig(repository.HTTP, "http-repo", "https://raw.githubusercontent.com/LLMrecon/LLMrecon/main/")
	httpRepo, err := repository.Create(httpConfig)
	if err != nil {
		fmt.Printf("Failed to create HTTP repository: %v\n", err)
		return
	}

	// Add to manager
	if err := manager.AddRepository(httpRepo); err != nil {
		fmt.Printf("Failed to add HTTP repository to manager: %v\n", err)
		return
	}

	// Connect to repository
	if err := httpRepo.Connect(ctx); err != nil {
		fmt.Printf("Failed to connect to HTTP repository: %v\n", err)
	} else {
		defer httpRepo.Disconnect()

		// Check if README.md exists
		exists, err := httpRepo.FileExists(ctx, "README.md")
		if err != nil {
			fmt.Printf("Failed to check if README.md exists: %v\n", err)
		} else {
			fmt.Printf("README.md exists: %v\n", exists)
			if exists {
				// Get file
				file, err := httpRepo.GetFile(ctx, "README.md")
				if err != nil {
					fmt.Printf("Failed to get README.md: %v\n", err)
				} else {
					defer file.Close()
					
					// Read first 100 bytes
					content := make([]byte, 100)
					n, err := file.Read(content)
					if err != nil && err != io.EOF {
						fmt.Printf("Failed to read README.md: %v\n", err)
					} else {
						fmt.Printf("First %d bytes of README.md: %s...\n", n, content[:n])
					}
				}
			}
		}
	}
	fmt.Println()

	// Example 3: Using the repository manager
	fmt.Println("=== Example 3: Repository Manager ===")
	
	// Connect to all repositories
	if err := manager.ConnectAll(ctx); err != nil {
		fmt.Printf("Failed to connect to all repositories: %v\n", err)
	}

	// List all repositories
	fmt.Println("All repositories:")
	for _, repo := range manager.ListRepositories() {
		fmt.Printf("- %s (%s): %s\n", repo.GetName(), repo.GetType(), repo.GetURL())
	}
	fmt.Println()

	// Find files matching a pattern across all repositories
	fmt.Println("Finding *.md files across all repositories:")
	filesByRepo, err := manager.FindFiles(ctx, "*.md")
	if err != nil {
		fmt.Printf("Failed to find files: %v\n", err)
	} else {
		for repo, files := range filesByRepo {
			fmt.Printf("Repository: %s\n", repo.GetName())
			for _, file := range files {
				fmt.Printf("  - %s\n", file.Path)
			}
		}
	}
	fmt.Println()

	// Example 4: GitHub repository (if credentials are available)
	// Note: This requires a GitHub token to be set in the environment
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken != "" {
		fmt.Println("=== Example 4: GitHub Repository ===")
		githubConfig := repository.NewConfig(repository.GitHub, "github-repo", "https://github.com/perplext/LLMrecon")
		githubConfig.Password = githubToken // Set token for authentication
		
		githubRepo, err := repository.Create(githubConfig)
		if err != nil {
			fmt.Printf("Failed to create GitHub repository: %v\n", err)
		} else {
			// Add to manager
			if err := manager.AddRepository(githubRepo); err != nil {
				fmt.Printf("Failed to add GitHub repository to manager: %v\n", err)
			} else {
				// Connect to repository
				if err := githubRepo.Connect(ctx); err != nil {
					fmt.Printf("Failed to connect to GitHub repository: %v\n", err)
				} else {
					defer githubRepo.Disconnect()
					
					// List files
					fmt.Println("Files in GitHub repository:")
					files, err := githubRepo.ListFiles(ctx, "*.go")
					if err != nil {
						fmt.Printf("Failed to list files: %v\n", err)
					} else {
						for _, file := range files {
							fmt.Printf("- %s\n", file.Path)
						}
					}
				}
			}
		}
	} else {
		fmt.Println("=== Example 4: GitHub Repository ===")
		fmt.Println("Skipping GitHub example because GITHUB_TOKEN is not set")
	}

	// Disconnect from all repositories
	if err := manager.DisconnectAll(); err != nil {
		fmt.Printf("Failed to disconnect from all repositories: %v\n", err)
	}
}
