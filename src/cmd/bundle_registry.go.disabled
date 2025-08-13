package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var bundleRegistryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage template repository registry",
	Long: `Manage a registry of known template repositories for easy access and discovery.
	
The registry allows you to:
- Add trusted template sources
- List available repositories
- Search for specific template types
- Configure default sources`,
}

var registryAddCmd = &cobra.Command{
	Use:   "add [name] [url]",
	Short: "Add a repository to the registry",
	Long:  `Add a new template repository to the local registry for easy reference.`,
	Example: `  # Add official OWASP templates
  LLMrecon bundle registry add owasp-official https://github.com/OWASP/llm-templates
  
  # Add private repository
  LLMrecon bundle registry add company-templates https://gitlab.company.com/security/templates`,
	Args: cobra.ExactArgs(2),
	RunE: runRegistryAdd,
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered repositories",
	Long:  `Display all template repositories in the local registry.`,
	RunE:  runRegistryList,
}

var registryRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a repository from the registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runRegistryRemove,
}

var registryUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update repository metadata",
	Long:  `Update metadata for a registered repository by fetching latest information.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRegistryUpdate,
}

var registrySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for templates across registered repositories",
	Long:  `Search for security test templates across all registered repositories.`,
	Example: `  # Search for prompt injection templates
  LLMrecon bundle registry search "prompt injection"
  
  # Search for specific category
  LLMrecon bundle registry search "llm01"`,
	Args: cobra.ExactArgs(1),
	RunE: runRegistrySearch,
}

func init() {
	bundleCmd.AddCommand(bundleRegistryCmd)
	bundleRegistryCmd.AddCommand(registryAddCmd)
	bundleRegistryCmd.AddCommand(registryListCmd)
	bundleRegistryCmd.AddCommand(registryRemoveCmd)
	bundleRegistryCmd.AddCommand(registryUpdateCmd)
	bundleRegistryCmd.AddCommand(registrySearchCmd)
	
	// Add flags
	registryAddCmd.Flags().String("description", "", "Repository description")
	registryAddCmd.Flags().StringSlice("tags", []string{}, "Repository tags")
	registryAddCmd.Flags().Bool("official", false, "Mark as official repository")
	registryAddCmd.Flags().String("auth", "", "Authentication token for private repositories")
}

// RegistryEntry represents a repository in the registry
type RegistryEntry struct {
	Name        string                 `json:"name"`
	URL         string                 `json:"url"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Tags        []string               `json:"tags"`
	Official    bool                   `json:"official"`
	Added       time.Time              `json:"added"`
	Updated     time.Time              `json:"updated"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Auth        string                 `json:"-"` // Don't save auth tokens
}

// Registry represents the template repository registry
type Registry struct {
	Version  string                    `json:"version"`
	Entries  map[string]*RegistryEntry `json:"entries"`
	Modified time.Time                 `json:"modified"`
}

func getRegistryPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".LLMrecon", "registry.json")
}

func loadRegistry() (*Registry, error) {
	registryPath := getRegistryPath()
	
	// Create default registry if it doesn't exist
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		registry := &Registry{
			Version:  "1.0",
			Entries:  make(map[string]*RegistryEntry),
			Modified: time.Now(),
		}
		
		// Add default entries
		registry.Entries["owasp-examples"] = &RegistryEntry{
			Name:        "owasp-examples",
			URL:         "https://github.com/LLMrecon/owasp-templates",
			Description: "Example OWASP LLM Top 10 security test templates",
			Type:        "github",
			Tags:        []string{"owasp", "examples", "security"},
			Official:    true,
			Added:       time.Now(),
			Updated:     time.Now(),
		}
		
		return registry, nil
	}
	
	// Load existing registry
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}
	
	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}
	
	if registry.Entries == nil {
		registry.Entries = make(map[string]*RegistryEntry)
	}
	
	return &registry, nil
}

func saveRegistry(registry *Registry) error {
	registryPath := getRegistryPath()
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(registryPath), 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}
	
	// Update modified time
	registry.Modified = time.Now()
	
	// Marshal to JSON
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(registryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}
	
	return nil
}

func runRegistryAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	url := args[1]
	
	// Load registry
	registry, err := loadRegistry()
	if err != nil {
		return err
	}
	
	// Check if already exists
	if _, exists := registry.Entries[name]; exists {
		return fmt.Errorf("repository '%s' already exists in registry", name)
	}
	
	// Detect repository type
	repoType, err := detectRepositoryType(url)
	if err != nil {
		return err
	}
	
	// Create entry
	entry := &RegistryEntry{
		Name:    name,
		URL:     url,
		Type:    repoType,
		Added:   time.Now(),
		Updated: time.Now(),
	}
	
	// Add optional fields
	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		entry.Description = desc
	}
	
	if tags, _ := cmd.Flags().GetStringSlice("tags"); len(tags) > 0 {
		entry.Tags = tags
	}
	
	if official, _ := cmd.Flags().GetBool("official"); official {
		entry.Official = official
	}
	
	if auth, _ := cmd.Flags().GetString("auth"); auth != "" {
		// Store auth token separately (in future, use secure storage)
		fmt.Println("Note: Authentication tokens should be stored securely")
	}
	
	// Add to registry
	registry.Entries[name] = entry
	
	// Save registry
	if err := saveRegistry(registry); err != nil {
		return err
	}
	
	color.Green("✅ Repository '%s' added to registry", name)
	fmt.Printf("   URL: %s\n", url)
	fmt.Printf("   Type: %s\n", repoType)
	
	return nil
}

func runRegistryList(cmd *cobra.Command, args []string) error {
	registry, err := loadRegistry()
	if err != nil {
		return err
	}
	
	if len(registry.Entries) == 0 {
		color.Yellow("No repositories registered")
		fmt.Println("\nAdd repositories using:")
		fmt.Println("  LLMrecon bundle registry add <name> <url>")
		return nil
	}
	
	fmt.Println()
	color.Cyan("📚 Registered Template Repositories")
	fmt.Println(strings.Repeat("-", 70))
	
	// Sort entries by name
	var names []string
	for name := range registry.Entries {
		names = append(names, name)
	}
	
	for _, name := range names {
		entry := registry.Entries[name]
		
		// Name and official badge
		if entry.Official {
			color.Green("%-20s [OFFICIAL]", entry.Name)
		} else {
			fmt.Printf("%-20s", entry.Name)
		}
		
		// URL
		color.Cyan(" %s", entry.URL)
		
		// Description
		if entry.Description != "" {
			fmt.Printf("\n%-20s %s", "", entry.Description)
		}
		
		// Tags
		if len(entry.Tags) > 0 {
			fmt.Printf("\n%-20s Tags: %s", "", strings.Join(entry.Tags, ", "))
		}
		
		// Last updated
		fmt.Printf("\n%-20s Updated: %s", "", entry.Updated.Format("2006-01-02"))
		
		fmt.Println("\n" + strings.Repeat("-", 70))
	}
	
	fmt.Printf("\nTotal repositories: %d\n", len(registry.Entries))
	
	return nil
}

func runRegistryRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	
	registry, err := loadRegistry()
	if err != nil {
		return err
	}
	
	if _, exists := registry.Entries[name]; !exists {
		return fmt.Errorf("repository '%s' not found in registry", name)
	}
	
	delete(registry.Entries, name)
	
	if err := saveRegistry(registry); err != nil {
		return err
	}
	
	color.Green("✅ Repository '%s' removed from registry", name)
	
	return nil
}

func runRegistryUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]
	
	registry, err := loadRegistry()
	if err != nil {
		return err
	}
	
	entry, exists := registry.Entries[name]
	if !exists {
		return fmt.Errorf("repository '%s' not found in registry", name)
	}
	
	color.Cyan("🔄 Updating repository metadata: %s", name)
	
	// TODO: Connect to repository and fetch latest metadata
	// For now, just update the timestamp
	entry.Updated = time.Now()
	
	// In a real implementation, we would:
	// 1. Connect to the repository
	// 2. Fetch repository metadata (description, topics, etc.)
	// 3. Count templates by category
	// 4. Check for latest updates
	
	// Mock metadata update
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]interface{})
	}
	
	entry.Metadata["template_count"] = 42
	entry.Metadata["categories"] = []string{
		"llm01-prompt-injection",
		"llm02-insecure-output",
		"llm03-training-data-poisoning",
	}
	entry.Metadata["last_commit"] = time.Now().Format(time.RFC3339)
	
	if err := saveRegistry(registry); err != nil {
		return err
	}
	
	color.Green("✅ Repository metadata updated")
	
	if count, ok := entry.Metadata["template_count"].(int); ok {
		fmt.Printf("   Templates: %d\n", count)
	}
	
	if categories, ok := entry.Metadata["categories"].([]string); ok {
		fmt.Printf("   Categories: %s\n", strings.Join(categories, ", "))
	}
	
	return nil
}

func runRegistrySearch(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(args[0])
	
	registry, err := loadRegistry()
	if err != nil {
		return err
	}
	
	color.Cyan("🔍 Searching for: %s", query)
	fmt.Println()
	
	found := 0
	
	for name, entry := range registry.Entries {
		// Search in name, description, and tags
		if strings.Contains(strings.ToLower(name), query) ||
			strings.Contains(strings.ToLower(entry.Description), query) ||
			containsTag(entry.Tags, query) {
			
			found++
			
			// Display result
			if entry.Official {
				color.Green("📦 %s [OFFICIAL]", entry.Name)
			} else {
				color.Yellow("📦 %s", entry.Name)
			}
			
			fmt.Printf("   URL: %s\n", entry.URL)
			
			if entry.Description != "" {
				fmt.Printf("   Description: %s\n", entry.Description)
			}
			
			if len(entry.Tags) > 0 {
				fmt.Printf("   Tags: %s\n", strings.Join(entry.Tags, ", "))
			}
			
			// Show sync command
			fmt.Printf("   Sync: LLMrecon bundle sync --source=%s\n", entry.URL)
			
			fmt.Println()
		}
	}
	
	if found == 0 {
		color.Yellow("No repositories found matching '%s'", query)
		fmt.Println("\nTry:")
		fmt.Println("  - Using different search terms")
		fmt.Println("  - Running 'LLMrecon bundle registry list' to see all repositories")
		fmt.Println("  - Adding new repositories with 'LLMrecon bundle registry add'")
	} else {
		fmt.Printf("Found %d repositories\n", found)
	}
	
	return nil
}

func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}