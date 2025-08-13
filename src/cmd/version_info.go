package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/perplext/LLMrecon/src/config"
	"github.com/perplext/LLMrecon/src/version"
	"github.com/spf13/cobra"
)

// Build information (set via ldflags)
var (
	buildDate   = "unknown"
	gitCommit   = "unknown"
	gitBranch   = "unknown"
	buildNumber = "dev"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long: `Display detailed version information for the LLMreconing Tool and its components.

This command shows:
- Core binary version
- Template collection version
- Module versions
- Build information
- System compatibility
- Dependencies (with --verbose)`,
	Example: `  # Show basic version information
  LLMrecon version

  # Show detailed version information
  LLMrecon version --verbose

  # Show version for specific component
  LLMrecon version --component=templates

  # Output in JSON format
  LLMrecon version --json`,
	RunE: runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Add flags
	versionCmd.Flags().BoolP("verbose", "v", false, "Show detailed version information")
	versionCmd.Flags().StringP("component", "c", "all", "Component to show (all, binary, templates, modules)")
	versionCmd.Flags().Bool("json", false, "Output in JSON format")
	versionCmd.Flags().Bool("check-compatibility", false, "Check component compatibility")
}

func runVersion(cmd *cobra.Command, args []string) error {
	// Get flags
	verboseFlag, _ := cmd.Flags().GetBool("verbose")
	componentFlag, _ := cmd.Flags().GetString("component")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	checkCompatFlag, _ := cmd.Flags().GetBool("check-compatibility")

	// Load configuration to get paths
	cfg, err := config.LoadConfig()
	if err != nil && !jsonFlag {
		fmt.Fprintf(os.Stderr, yellow("Warning: Could not load configuration: %v\n"), err)
	}

	// Collect version information
	versionInfo, err := collectVersionInfo(cfg)
	if err != nil {
		return fmt.Errorf("collecting version information: %w", err)
	}

	// Filter by component if requested
	if componentFlag != "all" {
		versionInfo = filterVersionInfo(versionInfo, componentFlag)
	}

	// Check compatibility if requested
	if checkCompatFlag {
		compatReport := checkComponentCompatibility(versionInfo)
		versionInfo.Compatibility = compatReport
	}

	// Output results
	if jsonFlag {
		return outputVersionJSON(versionInfo)
	}

	return outputVersionTable(versionInfo, verboseFlag)
}

// VersionInfo contains all version information
type VersionInfo struct {
	Binary        ComponentVersion     `json:"binary"`
	Templates     ComponentVersion     `json:"templates"`
	Modules       []ComponentVersion   `json:"modules"`
	Build         BuildInfo            `json:"build"`
	System        SystemInfo           `json:"system"`
	Dependencies  []DependencyInfo     `json:"dependencies,omitempty"`
	Compatibility *CompatibilityReport `json:"compatibility,omitempty"`
}

// ComponentVersion represents version info for a component
type ComponentVersion struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	ReleaseDate time.Time `json:"release_date,omitempty"`
	Author      string    `json:"author,omitempty"`
	Description string    `json:"description,omitempty"`
	License     string    `json:"license,omitempty"`
	Homepage    string    `json:"homepage,omitempty"`
}

// BuildInfo contains build-time information
type BuildInfo struct {
	Date      string `json:"date"`
	Commit    string `json:"commit"`
	Branch    string `json:"branch"`
	Number    string `json:"number"`
	GoVersion string `json:"go_version"`
	Compiler  string `json:"compiler"`
}

// SystemInfo contains system information
type SystemInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	CPUs         int    `json:"cpus"`
	GoMaxProcs   int    `json:"go_max_procs"`
}

// DependencyInfo contains dependency information
type DependencyInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	License string `json:"license,omitempty"`
}

// CompatibilityReport contains compatibility check results
type CompatibilityReport struct {
	Compatible      bool                 `json:"compatible"`
	Issues          []CompatibilityIssue `json:"issues,omitempty"`
	Recommendations []string             `json:"recommendations,omitempty"`
}

// CompatibilityIssue represents a compatibility problem
type CompatibilityIssue struct {
	Component string `json:"component"`
	Issue     string `json:"issue"`
	Severity  string `json:"severity"`
}

func collectVersionInfo(cfg *config.Config) (*VersionInfo, error) {
	info := &VersionInfo{
		Binary: ComponentVersion{
			Name:        "LLMrecon",
			Version:     currentVersion,
			Description: "LLMreconing Tool",
			License:     "MIT",
			Homepage:    "https://github.com/perplext/LLMrecon",
		},
		Build: BuildInfo{
			Date:      buildDate,
			Commit:    gitCommit,
			Branch:    gitBranch,
			Number:    buildNumber,
			GoVersion: runtime.Version(),
			Compiler:  runtime.Compiler,
		},
		System: SystemInfo{
			OS:           runtime.GOOS,
			Architecture: runtime.GOARCH,
			CPUs:         runtime.NumCPU(),
			GoMaxProcs:   runtime.GOMAXPROCS(0),
		},
	}

	// Get template and module versions
	templateVer, moduleVersions, err := getLocalVersions()
	if err != nil {
		return info, nil // Return partial info on error
	}

	info.Templates = ComponentVersion{
		Name:        "templates",
		Version:     templateVer.String(),
		Description: "Template collection",
	}

	// Add module versions
	for id, ver := range moduleVersions {
		info.Modules = append(info.Modules, ComponentVersion{
			Name:    id,
			Version: ver.String(),
		})
	}

	return info, nil
}

func filterVersionInfo(info *VersionInfo, component string) *VersionInfo {
	filtered := &VersionInfo{
		Build:  info.Build,
		System: info.System,
	}

	switch component {
	case "binary":
		filtered.Binary = info.Binary
	case "templates":
		filtered.Templates = info.Templates
	case "modules":
		filtered.Modules = info.Modules
	default:
		// Check if it's a specific module
		for _, mod := range info.Modules {
			if mod.Name == component {
				filtered.Modules = []ComponentVersion{mod}
				break
			}
		}
	}

	return filtered
}

func checkComponentCompatibility(info *VersionInfo) *CompatibilityReport {
	report := &CompatibilityReport{
		Compatible: true,
		Issues:     []CompatibilityIssue{},
	}

	// Parse versions
	binaryVer, _ := version.ParseVersion(info.Binary.Version)
	templateVer, _ := version.ParseVersion(info.Templates.Version)

	// Check binary vs template compatibility
	if binaryVer.Major != templateVer.Major {
		report.Compatible = false
		report.Issues = append(report.Issues, CompatibilityIssue{
			Component: "templates",
			Issue:     fmt.Sprintf("Major version mismatch: binary=%d, templates=%d", binaryVer.Major, templateVer.Major),
			Severity:  "high",
		})
		report.Recommendations = append(report.Recommendations,
			"Update templates to match binary major version")
	}

	// Check module compatibility
	for _, mod := range info.Modules {
		modVer, err := version.ParseVersion(mod.Version)
		if err != nil {
			continue
		}

		if modVer.Major < binaryVer.Major-1 || modVer.Major > binaryVer.Major {
			report.Compatible = false
			report.Issues = append(report.Issues, CompatibilityIssue{
				Component: mod.Name,
				Issue:     fmt.Sprintf("Module version %s may not be compatible with binary %s", mod.Version, info.Binary.Version),
				Severity:  "medium",
			})
		}
	}

	return report
}

func outputVersionTable(info *VersionInfo, verbose bool) error {
	// Header
	fmt.Println(bold("LLMreconing Tool"))
	fmt.Println()

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Basic version info
	fmt.Fprintf(w, "%s:\t%s\n", cyan("Version"), info.Binary.Version)
	fmt.Fprintf(w, "%s:\t%s\n", cyan("Build Date"), info.Build.Date)
	fmt.Fprintf(w, "%s:\t%s\n", cyan("Git Commit"), formatCommit(info.Build.Commit))

	if verbose {
		fmt.Fprintf(w, "%s:\t%s\n", cyan("Git Branch"), info.Build.Branch)
		fmt.Fprintf(w, "%s:\t%s\n", cyan("Build Number"), info.Build.Number)
		fmt.Fprintf(w, "%s:\t%s\n", cyan("Go Version"), info.Build.GoVersion)
		fmt.Fprintf(w, "%s:\t%s\n", cyan("Compiler"), info.Build.Compiler)
	}

	w.Flush()

	// System info
	if verbose {
		fmt.Println("\n" + bold("System Information:"))
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "%s:\t%s/%s\n", cyan("Platform"), info.System.OS, info.System.Architecture)
		fmt.Fprintf(w, "%s:\t%d\n", cyan("CPUs"), info.System.CPUs)
		fmt.Fprintf(w, "%s:\t%d\n", cyan("GOMAXPROCS"), info.System.GoMaxProcs)
		w.Flush()
	}

	// Component versions
	fmt.Println("\n" + bold("Components:"))
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if info.Templates.Version != "" {
		fmt.Fprintf(w, "%s:\t%s\n", cyan("Templates"), info.Templates.Version)
	}

	if len(info.Modules) > 0 {
		fmt.Fprintf(w, "%s:\n", cyan("Modules"))
		for _, mod := range info.Modules {
			fmt.Fprintf(w, "  %s:\t%s\n", mod.Name, mod.Version)
		}
	}

	w.Flush()

	// Compatibility report
	if info.Compatibility != nil {
		fmt.Println("\n" + bold("Compatibility Check:"))
		if info.Compatibility.Compatible {
			fmt.Println(green("✓") + " All components are compatible")
		} else {
			fmt.Println(red("✗") + " Compatibility issues detected:")
			for _, issue := range info.Compatibility.Issues {
				severityColor := yellow
				if issue.Severity == "high" {
					severityColor = red
				}
				fmt.Printf("  - %s: %s [%s]\n",
					issue.Component,
					issue.Issue,
					severityColor(issue.Severity))
			}

			if len(info.Compatibility.Recommendations) > 0 {
				fmt.Println("\n" + bold("Recommendations:"))
				for _, rec := range info.Compatibility.Recommendations {
					fmt.Printf("  • %s\n", rec)
				}
			}
		}
	}

	return nil
}

func outputVersionJSON(info *VersionInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

func formatCommit(commit string) string {
	if commit == "unknown" || commit == "" {
		return dim("unknown")
	}
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

// Extended version check for CI/CD
func CheckVersionForCI() (string, error) {
	info, err := collectVersionInfo(nil)
	if err != nil {
		return "", err
	}

	// Format version string for CI
	versionStr := fmt.Sprintf("%s+%s", info.Binary.Version, info.Build.Commit)
	if info.Build.Branch != "main" && info.Build.Branch != "master" {
		versionStr = fmt.Sprintf("%s-%s", versionStr, strings.ReplaceAll(info.Build.Branch, "/", "-"))
	}

	return versionStr, nil
}
