package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/update"
	"github.com/spf13/cobra"
)

// bundleInfoCmd represents the bundle info command
var bundleInfoCmd = &cobra.Command{
	Use:   "info PATH",
	Short: "Display bundle information",
	Long: `Display detailed information about an offline update bundle.

This command shows:
- Bundle metadata and version
- Component inventory
- OWASP LLM Top 10 categorization
- Compliance documentation status
- Template and module listings
- Size and checksum information`,
	Example: `  # Show basic bundle information
  LLMrecon bundle info update.bundle

  # Show detailed information
  LLMrecon bundle info update.bundle --verbose

  # Show compliance status
  LLMrecon bundle info update.bundle --show-compliance

  # Output in JSON format
  LLMrecon bundle info update.bundle --json

  # List all files in bundle
  LLMrecon bundle info update.bundle --list-files`,
	Args: cobra.ExactArgs(1),
	RunE: runBundleInfo,
}

func init() {
	bundleCmd.AddCommand(bundleInfoCmd)

	// Add flags
	bundleInfoCmd.Flags().BoolP("verbose", "v", false, "Show detailed information")
	bundleInfoCmd.Flags().Bool("show-compliance", false, "Show compliance documentation status")
	bundleInfoCmd.Flags().Bool("json", false, "Output in JSON format")
	bundleInfoCmd.Flags().Bool("list-files", false, "List all files in the bundle")
	bundleInfoCmd.Flags().Bool("show-checksums", false, "Show file checksums")
}

func runBundleInfo(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]

	// Get flags
	verbose, _ := cmd.Flags().GetBool("verbose")
	showCompliance, _ := cmd.Flags().GetBool("show-compliance")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	listFiles, _ := cmd.Flags().GetBool("list-files")
	showChecksums, _ := cmd.Flags().GetBool("show-checksums")

	// Check if bundle exists
	bundleInfo, err := os.Stat(bundlePath)
	if err != nil {
		return fmt.Errorf("bundle not found: %w", err)
	}

	// Load bundle
	b, err := bundle.LoadBundle(bundlePath)
	if err != nil {
		return fmt.Errorf("loading bundle: %w", err)
	}
	manifest := &b.Manifest

	// Calculate checksum
	checksum, _ := update.CalculateChecksum(bundlePath)

	// Collect bundle information
	info := collectBundleInfo(manifest, bundleInfo, checksum)

	// Output in JSON if requested
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(info)
	}

	// Display bundle information
	displayBundleInfo(info, verbose, showCompliance, listFiles, showChecksums)

	return nil
}

// BundleInfo contains comprehensive bundle information
type BundleInfo struct {
	Path       string             `json:"path"`
	Size       int64              `json:"size"`
	Checksum   string             `json:"checksum"`
	CreatedAt  time.Time          `json:"created_at"`
	Manifest   BundleManifestInfo `json:"manifest"`
	Components ComponentsInfo     `json:"components"`
	OWASP      OWASPInfo          `json:"owasp"`
	Compliance ComplianceInfo     `json:"compliance"`
	Statistics BundleStatistics   `json:"statistics"`
}

// BundleManifestInfo contains manifest information
type BundleManifestInfo struct {
	Version     string                 `json:"version"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ComponentsInfo contains component counts
type ComponentsInfo struct {
	Templates []ComponentItem `json:"templates"`
	Modules   []ComponentItem `json:"modules"`
	Documents []ComponentItem `json:"documents"`
	Resources []ComponentItem `json:"resources"`
}

// ComponentItem represents a component in the bundle
type ComponentItem struct {
	Name     string                 `json:"name"`
	Path     string                 `json:"path"`
	Version  string                 `json:"version"`
	Size     int64                  `json:"size"`
	Checksum string                 `json:"checksum"`
	Category string                 `json:"category,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// OWASPInfo contains OWASP categorization information
type OWASPInfo struct {
	Categorized   bool           `json:"categorized"`
	Categories    map[string]int `json:"categories"`
	Uncategorized int            `json:"uncategorized"`
}

// ComplianceInfo contains compliance documentation status
type ComplianceInfo struct {
	ISO42001  BundleComplianceStatus `json:"iso42001"`
	OWASP     BundleComplianceStatus `json:"owasp"`
	Documents []string               `json:"documents"`
}

// BundleComplianceStatus represents compliance standard status
type BundleComplianceStatus struct {
	Present     bool     `json:"present"`
	Version     string   `json:"version"`
	LastUpdated string   `json:"last_updated"`
	Files       []string `json:"files"`
}

// BundleStatistics contains bundle statistics
type BundleStatistics struct {
	TotalFiles      int    `json:"total_files"`
	TotalSize       int64  `json:"total_size"`
	TemplateCount   int    `json:"template_count"`
	ModuleCount     int    `json:"module_count"`
	DocumentCount   int    `json:"document_count"`
	LargestFile     string `json:"largest_file"`
	LargestFileSize int64  `json:"largest_file_size"`
}

func collectBundleInfo(manifest *bundle.BundleManifest, fileInfo os.FileInfo, checksum string) *BundleInfo {
	info := &BundleInfo{
		Path:      fileInfo.Name(),
		Size:      fileInfo.Size(),
		Checksum:  checksum,
		CreatedAt: fileInfo.ModTime(),
		Manifest: BundleManifestInfo{
			Version:     manifest.Version,
			Name:        manifest.Name,
			Description: manifest.Description,
			Author:      manifest.Author.Name,
			CreatedAt:   manifest.CreatedAt,
			Metadata:    manifest.Metadata,
		},
		Components: ComponentsInfo{
			Templates: []ComponentItem{},
			Modules:   []ComponentItem{},
			Documents: []ComponentItem{},
			Resources: []ComponentItem{},
		},
		OWASP: OWASPInfo{
			Categories: make(map[string]int),
		},
		Compliance: ComplianceInfo{
			Documents: []string{},
		},
		Statistics: BundleStatistics{},
	}

	// Process content items
	for _, content := range manifest.Content {
		item := ComponentItem{
			Name:     filepath.Base(content.Path),
			Path:     content.Path,
			Version:  content.Version,
			Size:     content.Size,
			Checksum: content.Checksum,
			Metadata: content.Metadata,
		}

		switch content.Type {
		case bundle.TemplateContentType:
			// Extract OWASP category
			if category, ok := content.Metadata["owasp_category"].(string); ok && category != "" {
				item.Category = category
				info.OWASP.Categories[category]++
				info.OWASP.Categorized = true
			} else {
				info.OWASP.Uncategorized++
			}

			info.Components.Templates = append(info.Components.Templates, item)
			info.Statistics.TemplateCount++

		case bundle.ModuleContentType:
			// Extract module type
			if moduleType, ok := content.Metadata["type"].(string); ok {
				item.Type = moduleType
			}

			info.Components.Modules = append(info.Components.Modules, item)
			info.Statistics.ModuleCount++
		}

		info.Statistics.TotalSize += content.Size

		// Track largest file
		if content.Size > info.Statistics.LargestFileSize {
			info.Statistics.LargestFileSize = content.Size
			info.Statistics.LargestFile = content.Path
		}
	}

	info.Statistics.TotalFiles = len(manifest.Content)

	// Check for compliance files
	for _, content := range manifest.Content {
		// Check for documents
		if strings.HasPrefix(content.Path, "docs/") {
			item := ComponentItem{
				Name:     filepath.Base(content.Path),
				Path:     content.Path,
				Size:     content.Size,
				Checksum: content.Checksum,
			}
			info.Components.Documents = append(info.Components.Documents, item)
			info.Statistics.DocumentCount++

			// Check for compliance documents
			if strings.Contains(content.Path, "iso42001") {
				info.Compliance.ISO42001.Present = true
				info.Compliance.ISO42001.Files = append(info.Compliance.ISO42001.Files, content.Path)
			}
			if strings.Contains(content.Path, "owasp") {
				info.Compliance.OWASP.Present = true
				info.Compliance.OWASP.Files = append(info.Compliance.OWASP.Files, content.Path)
			}
			if strings.Contains(content.Path, "compliance") {
				info.Compliance.Documents = append(info.Compliance.Documents, content.Path)
			}
		}
	}

	// Extract compliance metadata
	if compliance, ok := manifest.Metadata["compliance"].(map[string]interface{}); ok {
		if iso, ok := compliance["iso42001"].(bool); ok && iso {
			info.Compliance.ISO42001.Present = true
		}
		if owasp, ok := compliance["owasp"].(bool); ok && owasp {
			info.Compliance.OWASP.Present = true
		}
	}

	return info
}

func displayBundleInfo(info *BundleInfo, verbose, showCompliance, listFiles, showChecksums bool) {
	// Header
	fmt.Printf("%s %s\n", bold("Bundle:"), info.Path)
	fmt.Println(strings.Repeat("─", 60))

	// Basic information
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s:\t%s\n", cyan("Version"), info.Manifest.Version)
	fmt.Fprintf(w, "%s:\t%s\n", cyan("Size"), formatSize(info.Size))
	fmt.Fprintf(w, "%s:\t%s\n", cyan("Created"), info.CreatedAt.Format("2006-01-02 15:04:05"))

	if info.Manifest.Author != "" {
		fmt.Fprintf(w, "%s:\t%s\n", cyan("Author"), info.Manifest.Author)
	}

	if info.Manifest.Description != "" {
		fmt.Fprintf(w, "%s:\t%s\n", cyan("Description"), info.Manifest.Description)
	}

	if verbose {
		fmt.Fprintf(w, "%s:\t%s\n", cyan("Checksum"), info.Checksum[:16]+"...")
	}

	w.Flush()

	// Component summary
	fmt.Println("\n" + bold("Components:"))
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if info.Statistics.TemplateCount > 0 {
		fmt.Fprintf(w, "  %s:\t%d\n", cyan("Templates"), info.Statistics.TemplateCount)
	}
	if info.Statistics.ModuleCount > 0 {
		fmt.Fprintf(w, "  %s:\t%d\n", cyan("Modules"), info.Statistics.ModuleCount)
	}
	if info.Statistics.DocumentCount > 0 {
		fmt.Fprintf(w, "  %s:\t%d\n", cyan("Documents"), info.Statistics.DocumentCount)
	}

	fmt.Fprintf(w, "  %s:\t%d files, %s total\n",
		cyan("Total"),
		info.Statistics.TotalFiles,
		formatSize(info.Statistics.TotalSize))

	w.Flush()

	// OWASP categorization
	if info.OWASP.Categorized {
		fmt.Println("\n" + bold("OWASP LLM Top 10 Categories:"))

		// Sort categories
		categories := make([]string, 0, len(info.OWASP.Categories))
		for cat := range info.OWASP.Categories {
			categories = append(categories, cat)
		}
		sort.Strings(categories)

		for _, cat := range categories {
			count := info.OWASP.Categories[cat]
			fmt.Printf("  %s %s: %d templates\n", getCategoryIcon(cat), cat, count)
		}

		if info.OWASP.Uncategorized > 0 {
			fmt.Printf("  %s Uncategorized: %d templates\n", dim("○"), info.OWASP.Uncategorized)
		}
	}

	// Compliance status
	if showCompliance || info.Compliance.ISO42001.Present || info.Compliance.OWASP.Present {
		fmt.Println("\n" + bold("Compliance Documentation:"))

		if info.Compliance.ISO42001.Present {
			fmt.Printf("  %s ISO/IEC 42001: %s\n", green("✓"), "Present")
			if verbose && len(info.Compliance.ISO42001.Files) > 0 {
				for _, file := range info.Compliance.ISO42001.Files {
					fmt.Printf("    %s %s\n", dim("•"), file)
				}
			}
		} else {
			fmt.Printf("  %s ISO/IEC 42001: %s\n", dim("○"), "Not found")
		}

		if info.Compliance.OWASP.Present {
			fmt.Printf("  %s OWASP Standards: %s\n", green("✓"), "Present")
			if verbose && len(info.Compliance.OWASP.Files) > 0 {
				for _, file := range info.Compliance.OWASP.Files {
					fmt.Printf("    %s %s\n", dim("•"), file)
				}
			}
		} else {
			fmt.Printf("  %s OWASP Standards: %s\n", dim("○"), "Not found")
		}
	}

	// List files if requested
	if listFiles {
		fmt.Println("\n" + bold("Templates:"))
		for _, template := range info.Components.Templates {
			fmt.Printf("  %s %s", getFileIcon(template.Path), template.Path)
			if template.Version != "" {
				fmt.Printf(" %s", dim(fmt.Sprintf("(v%s)", template.Version)))
			}
			if showChecksums && template.Checksum != "" {
				fmt.Printf(" %s", dim(template.Checksum[:12]+"..."))
			}
			fmt.Println()
		}

		if len(info.Components.Modules) > 0 {
			fmt.Println("\n" + bold("Modules:"))
			for _, module := range info.Components.Modules {
				fmt.Printf("  %s %s", getFileIcon(module.Path), module.Path)
				if module.Version != "" {
					fmt.Printf(" %s", dim(fmt.Sprintf("(v%s)", module.Version)))
				}
				if showChecksums && module.Checksum != "" {
					fmt.Printf(" %s", dim(module.Checksum[:12]+"..."))
				}
				fmt.Println()
			}
		}

		if len(info.Components.Documents) > 0 {
			fmt.Println("\n" + bold("Documents:"))
			for _, doc := range info.Components.Documents {
				fmt.Printf("  %s %s %s\n",
					getFileIcon(doc.Path),
					doc.Path,
					dim(formatSize(doc.Size)))
			}
		}
	}

	// Statistics in verbose mode
	if verbose {
		fmt.Println("\n" + bold("Statistics:"))
		fmt.Printf("  Largest file: %s (%s)\n",
			info.Statistics.LargestFile,
			formatSize(info.Statistics.LargestFileSize))

		avgSize := int64(0)
		if info.Statistics.TotalFiles > 0 {
			avgSize = info.Statistics.TotalSize / int64(info.Statistics.TotalFiles)
		}
		fmt.Printf("  Average file size: %s\n", formatSize(avgSize))
	}
}

func getCategoryIcon(category string) string {
	icons := map[string]string{
		"llm01-prompt-injection":        "🔐",
		"llm02-insecure-output":         "⚠️",
		"llm03-training-data-poisoning": "☠️",
		"llm04-model-denial-of-service": "🚫",
		"llm05-supply-chain":            "🔗",
		"llm06-sensitive-information":   "🔒",
		"llm07-insecure-plugin":         "🔌",
		"llm08-excessive-agency":        "🤖",
		"llm09-overreliance":            "⚖️",
		"llm10-model-theft":             "🎭",
	}

	if icon, ok := icons[category]; ok {
		return icon
	}
	return "📁"
}

func getFileIcon(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return "📄"
	case ".json":
		return "📋"
	case ".md":
		return "📝"
	case ".go":
		return "🔧"
	case ".py":
		return "🐍"
	default:
		return "📄"
	}
}
