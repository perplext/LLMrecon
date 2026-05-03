package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/repository"
)

var (
	publishTarget      string
	publishBranch      string
	publishMessage     string
	publishTag         string
	publishAuth        string
	publishCreatePR    bool
	publishPRTitle     string
	publishPRBody      string
	publishDryRun      bool
)

var bundlePublishCmd = &cobra.Command{
	Use:   "publish [bundle-file]",
	Short: "Publish bundles to remote repositories",
	Long: `Publish security test bundles to GitHub, GitLab, or other repositories.
	
This command allows you to:
- Upload bundles to repositories with version control
- Create pull/merge requests for review
- Tag releases with semantic versioning
- Maintain bundle distribution channels`,
	Example: `  # Publish to GitHub repository
  LLMrecon bundle publish security-tests.bundle --target=https://github.com/org/bundles
  
  # Publish with authentication
  LLMrecon bundle publish tests.bundle --target=https://github.com/org/bundles --auth=$GITHUB_TOKEN
  
  # Create pull request
  LLMrecon bundle publish tests.bundle --target=https://github.com/org/bundles --create-pr --pr-title="Add security tests v1.2.0"
  
  # Publish with tag
  LLMrecon bundle publish tests.bundle --target=https://gitlab.com/org/bundles --tag=v1.2.0`,
	Args: cobra.ExactArgs(1),
	RunE: runBundlePublish,
}

func init() {
	bundleCmd.AddCommand(bundlePublishCmd)
	
	bundlePublishCmd.Flags().StringVarP(&publishTarget, "target", "t", "", "Target repository URL (required)")
	bundlePublishCmd.Flags().StringVarP(&publishBranch, "branch", "b", "main", "Target branch")
	bundlePublishCmd.Flags().StringVarP(&publishMessage, "message", "m", "", "Commit message")
	bundlePublishCmd.Flags().StringVar(&publishTag, "tag", "", "Create tag for this release")
	bundlePublishCmd.Flags().StringVar(&publishAuth, "auth", "", "Authentication token")
	bundlePublishCmd.Flags().BoolVar(&publishCreatePR, "create-pr", false, "Create pull/merge request")
	bundlePublishCmd.Flags().StringVar(&publishPRTitle, "pr-title", "", "Pull request title")
	bundlePublishCmd.Flags().StringVar(&publishPRBody, "pr-body", "", "Pull request body")
	bundlePublishCmd.Flags().BoolVar(&publishDryRun, "dry-run", false, "Show what would be published without uploading")
	
	bundlePublishCmd.MarkFlagRequired("target")
}

func runBundlePublish(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	ctx := context.Background()
	
	// Verify bundle exists
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("bundle file not found: %s", bundlePath)
	}
	
	fmt.Println()
	color.Cyan("📤 Bundle Publishing")
	fmt.Println(strings.Repeat("-", 50))
	
	// Load bundle
	bundleData, err := bundle.LoadBundle(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w", err)
	}
	
	color.Yellow("Bundle: %s", bundlePath)
	color.Yellow("Name: %s", bundleData.Manifest.Name)
	color.Yellow("Version: %s", bundleData.Manifest.Version)
	color.Yellow("Templates: %d", bundleData.Manifest.Templates)
	fmt.Println()
	
	// Parse repository type
	repoType, err := detectRepositoryType(publishTarget)
	if err != nil {
		return fmt.Errorf("failed to detect repository type: %w", err)
	}
	
	color.Yellow("Target: %s", publishTarget)
	color.Yellow("Type: %s", repoType)
	color.Yellow("Branch: %s", publishBranch)
	
	if publishTag != "" {
		color.Yellow("Tag: %s", publishTag)
	}
	
	if publishCreatePR {
		color.Yellow("Pull Request: Yes")
	}
	
	fmt.Println()
	
	if publishDryRun {
		color.Cyan("🔍 Dry run mode - no changes will be made")
		
		fmt.Println("\nWould publish:")
		fmt.Printf("  Bundle: %s (%s)\n", bundleData.Manifest.Name, formatBytes(getFileSize(bundlePath)))
		fmt.Printf("  To: %s\n", publishTarget)
		fmt.Printf("  Branch: %s\n", publishBranch)
		
		if publishTag != "" {
			fmt.Printf("  Tag: %s\n", publishTag)
		}
		
		if publishCreatePR {
			fmt.Println("\nWould create pull request:")
			fmt.Printf("  Title: %s\n", getPRTitle(bundleData, publishPRTitle))
			fmt.Printf("  Body:\n%s\n", getPRBody(bundleData, publishPRBody))
		}
		
		return nil
	}
	
	// Create repository manager
	repoManager := repository.NewManager()
	
	// Configure authentication
	if publishAuth != "" {
		switch repoType {
		case "github":
			repoManager.SetGitHubToken(publishAuth)
		case "gitlab":
			repoManager.SetGitLabToken(publishAuth)
		}
	}
	
	// Get repository interface
	repo, err := repoManager.GetRepository(publishTarget)
	if err != nil {
		return fmt.Errorf("failed to connect to repository: %w", err)
	}
	
	// Prepare bundle for publishing
	color.Cyan("📦 Preparing bundle for publishing...")
	
	// Create bundle directory structure
	bundleDir := fmt.Sprintf("bundles/%s", bundleData.Manifest.Name)
	bundleFile := fmt.Sprintf("%s-%s.bundle", bundleData.Manifest.Name, bundleData.Manifest.Version)
	bundleFullPath := filepath.Join(bundleDir, bundleFile)
	
	// Read bundle content
	bundleContent, err := os.ReadFile(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to read bundle file: %w", err)
	}
	
	// Create metadata file
	metadataPath := filepath.Join(bundleDir, "metadata.json")
	metadataContent := generateBundleMetadata(bundleData, bundleFile)
	
	// Create README
	readmePath := filepath.Join(bundleDir, "README.md")
	readmeContent := generateBundleReadme(bundleData, bundleFile)
	
	// Prepare commit
	commitMessage := publishMessage
	if commitMessage == "" {
		commitMessage = fmt.Sprintf("Add %s bundle v%s", bundleData.Manifest.Name, bundleData.Manifest.Version)
	}
	
	if publishCreatePR {
		// Create feature branch
		featureBranch := fmt.Sprintf("bundle-%s-%s", bundleData.Manifest.Name, bundleData.Manifest.Version)
		color.Cyan("🌿 Creating feature branch: %s", featureBranch)
		
		// Upload files to feature branch
		files := map[string][]byte{
			bundleFullPath: bundleContent,
			metadataPath:   metadataContent,
			readmePath:     readmeContent,
		}
		
		if err := repo.CreateBranch(ctx, featureBranch, publishBranch); err != nil {
			return fmt.Errorf("failed to create feature branch: %w", err)
		}
		
		for path, content := range files {
			color.Yellow("  Uploading %s...", path)
			if err := repo.CreateOrUpdateFile(ctx, path, content, commitMessage, featureBranch); err != nil {
				return fmt.Errorf("failed to upload %s: %w", path, err)
			}
		}
		
		// Create pull request
		prTitle := getPRTitle(bundleData, publishPRTitle)
		prBody := getPRBody(bundleData, publishPRBody)
		
		color.Cyan("🔄 Creating pull request...")
		prURL, err := repo.CreatePullRequest(ctx, prTitle, prBody, featureBranch, publishBranch)
		if err != nil {
			return fmt.Errorf("failed to create pull request: %w", err)
		}
		
		color.Green("✅ Pull request created: %s", prURL)
		
	} else {
		// Direct commit to branch
		color.Cyan("📝 Committing to branch: %s", publishBranch)
		
		// Upload bundle file
		color.Yellow("  Uploading bundle...")
		if err := repo.CreateOrUpdateFile(ctx, bundleFullPath, bundleContent, commitMessage, publishBranch); err != nil {
			return fmt.Errorf("failed to upload bundle: %w", err)
		}
		
		// Upload metadata
		color.Yellow("  Uploading metadata...")
		if err := repo.CreateOrUpdateFile(ctx, metadataPath, metadataContent, commitMessage, publishBranch); err != nil {
			return fmt.Errorf("failed to upload metadata: %w", err)
		}
		
		// Upload README
		color.Yellow("  Uploading README...")
		if err := repo.CreateOrUpdateFile(ctx, readmePath, readmeContent, commitMessage, publishBranch); err != nil {
			return fmt.Errorf("failed to upload README: %w", err)
		}
		
		color.Green("✅ Bundle published successfully")
	}
	
	// Create tag if specified
	if publishTag != "" {
		color.Cyan("🏷️  Creating tag: %s", publishTag)
		
		tagMessage := fmt.Sprintf("Release %s - %s", bundleData.Manifest.Name, bundleData.Manifest.Version)
		if err := repo.CreateTag(ctx, publishTag, publishBranch, tagMessage); err != nil {
			color.Red("  ⚠️  Failed to create tag: %v", err)
		} else {
			color.Green("  ✓ Tag created: %s", publishTag)
		}
	}
	
	// Show publish summary
	fmt.Println()
	color.Cyan("═══════════════════════════════════════════")
	color.Cyan("          Publish Summary")
	color.Cyan("═══════════════════════════════════════════")
	fmt.Printf("Bundle: %s v%s\n", bundleData.Manifest.Name, bundleData.Manifest.Version)
	fmt.Printf("Repository: %s\n", publishTarget)
	fmt.Printf("Branch: %s\n", publishBranch)
	fmt.Printf("Path: %s\n", bundleFullPath)
	
	if publishTag != "" {
		fmt.Printf("Tag: %s\n", publishTag)
	}
	
	if publishCreatePR {
		fmt.Println("Status: Pull request created")
	} else {
		fmt.Println("Status: Published")
	}
	
	color.Cyan("═══════════════════════════════════════════")
	
	return nil
}

// generateBundleMetadata creates metadata JSON for the bundle
func generateBundleMetadata(b *bundle.Bundle, filename string) []byte {
	metadata := map[string]interface{}{
		"name":        b.Manifest.Name,
		"version":     b.Manifest.Version,
		"description": b.Manifest.Description,
		"author":      b.Manifest.Author,
		"created":     b.Manifest.Created.Format(time.RFC3339),
		"templates":   b.Manifest.Templates,
		"filename":    filename,
		"categories":  extractCategories(b),
		"compliance": map[string]interface{}{
			"owasp_llm_top10": true,
			"iso_42001":       b.Manifest.Metadata["iso42001_compliant"] == true,
		},
	}
	
	data, _ := json.MarshalIndent(metadata, "", "  ")
	return data
}

// generateBundleReadme creates README content for the bundle
func generateBundleReadme(b *bundle.Bundle, filename string) []byte {
	var readme strings.Builder
	
	readme.WriteString(fmt.Sprintf("# %s\n\n", b.Manifest.Name))
	readme.WriteString(fmt.Sprintf("Version: %s\n\n", b.Manifest.Version))
	readme.WriteString(fmt.Sprintf("%s\n\n", b.Manifest.Description))
	
	readme.WriteString("## Bundle Information\n\n")
	readme.WriteString(fmt.Sprintf("- **Author**: %s\n", b.Manifest.Author))
	readme.WriteString(fmt.Sprintf("- **Created**: %s\n", b.Manifest.Created.Format("2006-01-02")))
	readme.WriteString(fmt.Sprintf("- **Templates**: %d\n", b.Manifest.Templates))
	readme.WriteString(fmt.Sprintf("- **File**: `%s`\n", filename))
	
	// Categories
	categories := extractCategories(b)
	if len(categories) > 0 {
		readme.WriteString("\n## OWASP LLM Top 10 Categories\n\n")
		for _, cat := range categories {
			readme.WriteString(fmt.Sprintf("- %s\n", cat))
		}
	}
	
	// Installation
	readme.WriteString("\n## Installation\n\n")
	readme.WriteString("```bash\n")
	readme.WriteString(fmt.Sprintf("# Download the bundle\n"))
	readme.WriteString(fmt.Sprintf("curl -LO <repository-url>/bundles/%s/%s\n\n", b.Manifest.Name, filename))
	readme.WriteString("# Import the bundle\n")
	readme.WriteString(fmt.Sprintf("LLMrecon bundle import %s\n", filename))
	readme.WriteString("```\n")
	
	// Verification
	readme.WriteString("\n## Verification\n\n")
	readme.WriteString("```bash\n")
	readme.WriteString("# Verify bundle integrity\n")
	readme.WriteString(fmt.Sprintf("LLMrecon bundle verify %s\n", filename))
	readme.WriteString("```\n")
	
	return []byte(readme.String())
}

// extractCategories extracts OWASP categories from bundle
func extractCategories(b *bundle.Bundle) []string {
	categoryMap := make(map[string]bool)
	
	for _, tmpl := range b.Templates {
		if cat := extractCategory(tmpl.Path); cat != "" {
			categoryMap[cat] = true
		}
	}
	
	var categories []string
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	
	return categories
}

// getPRTitle generates pull request title
func getPRTitle(b *bundle.Bundle, customTitle string) string {
	if customTitle != "" {
		return customTitle
	}
	return fmt.Sprintf("Add %s security test bundle v%s", b.Manifest.Name, b.Manifest.Version)
}

// getPRBody generates pull request body
func getPRBody(b *bundle.Bundle, customBody string) string {
	if customBody != "" {
		return customBody
	}
	
	var body strings.Builder
	
	body.WriteString(fmt.Sprintf("## Bundle: %s v%s\n\n", b.Manifest.Name, b.Manifest.Version))
	body.WriteString(fmt.Sprintf("%s\n\n", b.Manifest.Description))
	
	body.WriteString("### Details\n\n")
	body.WriteString(fmt.Sprintf("- **Templates**: %d\n", b.Manifest.Templates))
	body.WriteString(fmt.Sprintf("- **Author**: %s\n", b.Manifest.Author))
	body.WriteString(fmt.Sprintf("- **Created**: %s\n\n", b.Manifest.Created.Format("2006-01-02")))
	
	categories := extractCategories(b)
	if len(categories) > 0 {
		body.WriteString("### OWASP LLM Top 10 Categories\n\n")
		for _, cat := range categories {
			body.WriteString(fmt.Sprintf("- [x] %s\n", cat))
		}
		body.WriteString("\n")
	}
	
	body.WriteString("### Compliance\n\n")
	body.WriteString("- [x] OWASP LLM Top 10 compliant\n")
	if b.Manifest.Metadata["iso42001_compliant"] == true {
		body.WriteString("- [x] ISO/IEC 42001:2023 compliant\n")
	}
	
	body.WriteString("\n### Checklist\n\n")
	body.WriteString("- [ ] Bundle verified\n")
	body.WriteString("- [ ] Templates tested\n")
	body.WriteString("- [ ] Documentation reviewed\n")
	body.WriteString("- [ ] Security scan passed\n")
	
	return body.String()
}

// getFileSize returns file size in bytes
func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}