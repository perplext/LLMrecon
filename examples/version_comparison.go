package main

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/auth"
	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/version"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Set up local and remote repositories
	localRepoPath := "./examples"
	remoteRepoPath := "./examples/remote" // This would typically be a remote URL

	// Create repositories
	localConfig := repository.NewConfig(repository.LocalFS, "local-repo", localRepoPath)
	localRepo, err := repository.Create(localConfig)
	if err != nil {
		fmt.Printf("Failed to create local repository: %v\n", err)
		return
	}

	remoteConfig := repository.NewConfig(repository.LocalFS, "remote-repo", remoteRepoPath)
	remoteRepo, err := repository.Create(remoteConfig)
	if err != nil {
		fmt.Printf("Failed to create remote repository: %v\n", err)
		return
	}

	// Connect to repositories
	if err := localRepo.Connect(ctx); err != nil {
		fmt.Printf("Failed to connect to local repository: %v\n", err)
		return
	}
	defer localRepo.Disconnect()

	if err := remoteRepo.Connect(ctx); err != nil {
		fmt.Printf("Failed to connect to remote repository: %v\n", err)
		return
	}
	defer remoteRepo.Disconnect()

	// Example 1: Semantic Version Comparison
	fmt.Println("=== Example 1: Semantic Version Comparison ===")
	
	v1, err := version.Parse("1.2.3")
	if err != nil {
		fmt.Printf("Failed to parse version: %v\n", err)
	} else {
		fmt.Printf("Version 1: %s\n", v1.String())
	}
	
	v2, err := version.Parse("1.3.0")
	if err != nil {
		fmt.Printf("Failed to parse version: %v\n", err)
	} else {
		fmt.Printf("Version 2: %s\n", v2.String())
	}
	
	if v1.LessThan(v2) {
		fmt.Printf("%s is less than %s\n", v1.String(), v2.String())
	}
	
	if v1.IsCompatible(v2) {
		fmt.Printf("%s is compatible with %s\n", v1.String(), v2.String())
	}
	
	v3 := v1.IncrementMinor()
	fmt.Printf("Incrementing minor version of %s: %s\n", v1.String(), v3.String())
	
	fmt.Println()

	// Example 2: File Differential Analysis
	fmt.Println("=== Example 2: File Differential Analysis ===")
	
	// Create example files for diffing
	oldContent := []byte("This is the old content.\nIt has multiple lines.\nSome content will change.")
	newContent := []byte("This is the new content.\nIt has multiple lines.\nSome content has changed.")
	
	oldFileInfo := &version.FileInfo{
		Path:    "example.txt",
		Hash:    version.ComputeHash(oldContent),
		Size:    int64(len(oldContent)),
		ModTime: time.Now().Add(-24 * time.Hour),
		Content: oldContent,
	}
	
	newFileInfo := &version.FileInfo{
		Path:    "example.txt",
		Hash:    version.ComputeHash(newContent),
		Size:    int64(len(newContent)),
		ModTime: time.Now(),
		Content: newContent,
	}
	
	// Perform diff
	diffOptions := version.DefaultDiffOptions()
	diffOptions.IncludeContent = true
	
	diffResult := version.DiffFiles([]*version.FileInfo{oldFileInfo}, []*version.FileInfo{newFileInfo}, diffOptions)
	
	// Print diff summary
	fmt.Printf("Diff Summary: %s\n", diffResult.GetChangeSummary())
	
	// Print file diffs
	for _, diff := range diffResult.FileDiffs {
		fmt.Printf("File: %s, Type: %s\n", diff.Path, diff.Type)
		fmt.Printf("  Old Hash: %s, Size: %d bytes\n", diff.OldHash, diff.OldSize)
		fmt.Printf("  New Hash: %s, Size: %d bytes\n", diff.NewHash, diff.NewSize)
	}
	
	// Print content diffs
	for _, diff := range diffResult.ContentDiffs {
		fmt.Printf("Content Diff for %s:\n", diff.Path)
		for _, chunk := range diff.Chunks {
			fmt.Printf("  Chunk: Old Lines %d-%d, New Lines %d-%d\n", 
				chunk.OldStart, chunk.OldStart+chunk.OldLines-1,
				chunk.NewStart, chunk.NewStart+chunk.NewLines-1)
			
			if len(chunk.Content) > 0 {
				fmt.Println("  Content:")
				for _, line := range chunk.Content {
					fmt.Printf("    %s\n", line)
				}
			}
		}
	}
	
	fmt.Println()

	// Example 3: Dependency Graph
	fmt.Println("=== Example 3: Dependency Graph ===")
	
	// Create a dependency graph
	graph := version.NewDependencyGraph()
	
	// Add nodes
	graph.AddNode("template-a", "Template A", "template", version.MustParse("1.0.0"), nil)
	graph.AddNode("module-b", "Module B", "module", version.MustParse("1.1.0"), nil)
	graph.AddNode("module-c", "Module C", "module", version.MustParse("1.2.0"), nil)
	
	// Add dependencies
	if err := graph.AddDependency("template-a", "module-b", ">=1.0.0", false); err != nil {
		fmt.Printf("Failed to add dependency: %v\n", err)
	}
	
	if err := graph.AddDependency("template-a", "module-c", ">=1.0.0", false); err != nil {
		fmt.Printf("Failed to add dependency: %v\n", err)
	}
	
	// Get topological order
	order, err := graph.GetTopologicalOrder()
	if err != nil {
		fmt.Printf("Failed to get topological order: %v\n", err)
	} else {
		fmt.Println("Topological Order:")
		for i, node := range order {
			fmt.Printf("  %d. %s (%s %s)\n", i+1, node.Name, node.Type, node.Version.String())
		}
	}
	
	// Get impacted nodes
	impacted, err := graph.GetImpactedNodes("module-b")
	if err != nil {
		fmt.Printf("Failed to get impacted nodes: %v\n", err)
	} else {
		fmt.Println("Nodes impacted by changes to Module B:")
		for _, node := range impacted {
			fmt.Printf("  - %s (%s %s)\n", node.Name, node.Type, node.Version.String())
		}
	}
	
	fmt.Println()

	// Example 4: Version Analyzer (if example template/module files exist)
	fmt.Println("=== Example 4: Version Analyzer ===")
	
	// Create directories for example
	exampleDir := filepath.Join(localRepoPath, "version_example")
	remoteExampleDir := filepath.Join(remoteRepoPath, "version_example")
	
	if err := os.MkdirAll(exampleDir, 0755); err != nil {
		fmt.Printf("Failed to create example directory: %v\n", err)
	}
	
	if err := os.MkdirAll(remoteExampleDir, 0755); err != nil {
		fmt.Printf("Failed to create remote example directory: %v\n", err)
	}
	
	// Create example template files
	localTemplateContent := `
id: prompt-injection-basic
name: Basic Prompt Injection
version: 1.0.0
category: prompt-injection
description: Tests for basic prompt injection vulnerabilities
author: Security Team
dependencies:
  - moduleID: prompt-formatter
    version: ">=1.0.0"
`
	
	remoteTemplateContent := `
id: prompt-injection-basic
name: Basic Prompt Injection
version: 1.1.0
category: prompt-injection
description: Tests for basic prompt injection vulnerabilities with improved detection
author: Security Team
dependencies:
  - moduleID: prompt-formatter
    version: ">=1.0.0"
`
	
	localTemplatePath := filepath.Join(exampleDir, "basic_prompt_injection.yaml")
	remoteTemplatePath := filepath.Join(remoteExampleDir, "basic_prompt_injection.yaml")
	
	if err := os.WriteFile(localTemplatePath, []byte(localTemplateContent), 0644); err != nil {
		fmt.Printf("Failed to write local template file: %v\n", err)
	}
	
	if err := os.WriteFile(remoteTemplatePath, []byte(remoteTemplateContent), 0644); err != nil {
		fmt.Printf("Failed to write remote template file: %v\n", err)
	}
	
	// Create analyzer
	analyzer := version.NewAnalyzer(localRepo, remoteRepo, nil)
	
	// Analyze template
	relativePath := filepath.Join("version_example", "basic_prompt_injection.yaml")
	result, err := analyzer.AnalyzeTemplate(ctx, relativePath)
	if err != nil {
		fmt.Printf("Failed to analyze template: %v\n", err)
	} else {
		fmt.Printf("Template Analysis Result:\n")
		fmt.Printf("  Local Version: %s\n", result.LocalVersion.Version.String())
		fmt.Printf("  Remote Version: %s\n", result.RemoteVersion.Version.String())
		fmt.Printf("  Update Required: %v\n", result.UpdateRequired)
		fmt.Printf("  Changes: %s\n", result.Diff.GetChangeSummary())
	}
	
	fmt.Println()
	fmt.Println("Version Comparison and Differential Analysis Example Complete")
	
	// Clean up example files
	os.Remove(localTemplatePath)
	os.Remove(remoteTemplatePath)
	os.Remove(exampleDir)
	os.Remove(remoteExampleDir)
}
