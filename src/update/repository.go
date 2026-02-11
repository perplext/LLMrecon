package update

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/version"
)

// validateRepositoryURL validates a repository URL to prevent command injection
func validateRepositoryURL(repoURL string) error {
	// Reject URLs starting with a dash (git flag injection)
	if strings.HasPrefix(repoURL, "-") {
		return fmt.Errorf("invalid repository URL: must not start with a dash")
	}

	// Allow local file paths
	if filepath.IsAbs(repoURL) {
		// Reject path traversal
		cleaned := filepath.Clean(repoURL)
		if cleaned != repoURL {
			return fmt.Errorf("invalid repository path: must not contain path traversal")
		}
		return nil
	}

	// Validate as URL
	parsed, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	// Only allow known schemes
	switch parsed.Scheme {
	case "https", "http", "git", "ssh":
		// Valid schemes
	default:
		return fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
	}

	return nil
}

// validateLocalPath validates a local path to prevent command injection via git arguments
func validateLocalPath(localPath string) error {
	if localPath == "" {
		return fmt.Errorf("local path must not be empty")
	}
	// Reject paths starting with a dash (git flag injection)
	if strings.HasPrefix(filepath.Base(localPath), "-") {
		return fmt.Errorf("invalid local path: base name must not start with a dash")
	}
	// Ensure the path is clean (no .. traversal)
	cleaned := filepath.Clean(localPath)
	if cleaned != localPath {
		return fmt.Errorf("invalid local path: must not contain path traversal")
	}
	return nil
}

// Repository type constants (type defined in manager.go)
const (
	// GitHubRepository represents a GitHub repository
	GitHubRepository = RepositoryTypeGitHub
	// GitLabRepository represents a GitLab repository
	GitLabRepository = RepositoryTypeGitLab
	// LocalRepository represents a local repository
	LocalRepository = RepositoryTypeLocal
)

// RepositoryInfo represents information about a template repository
type RepositoryInfo struct {
	// Type of repository (github, gitlab, local)
	Type RepositoryType
	// URL of the repository
	URL string
	// Local path where the repository is cloned
	LocalPath string
	// Current version (commit hash or tag)
	CurrentVersion string
	// Latest available version
	LatestVersion string
	// Whether the repository has updates available
	HasUpdates bool
	// Last sync time
	LastSync time.Time
}

// RepositoryManager handles operations on template repositories
type RepositoryManager struct {
	// Base directory for repositories
	BaseDir string
	// Map of repository information by name
	Repositories map[string]*RepositoryInfo
}

// NewRepositoryManager creates a new RepositoryManager
func NewRepositoryManager(baseDir string) *RepositoryManager {
	return &RepositoryManager{
		BaseDir:      baseDir,
		Repositories: make(map[string]*RepositoryInfo),
	}
}

// AddRepository adds a repository to the manager
func (rm *RepositoryManager) AddRepository(name string, repoType RepositoryType, url string) (*RepositoryInfo, error) {
	// Check if repository already exists
	if _, exists := rm.Repositories[name]; exists {
		return nil, fmt.Errorf("repository with name '%s' already exists", name)
	}

	// Validate URL to prevent command injection via git subprocess
	if err := validateRepositoryURL(url); err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Create repository info
	repo := &RepositoryInfo{
		Type:      repoType,
		URL:       url,
		LocalPath: filepath.Join(rm.BaseDir, string(repoType), name),
	}

	// Add to map
	rm.Repositories[name] = repo

	return repo, nil
}

// SyncRepository syncs a repository (clone if it doesn't exist, pull if it does)
func (rm *RepositoryManager) SyncRepository(name string) error {
	repo, exists := rm.Repositories[name]
	if !exists {
		return fmt.Errorf("repository '%s' not found", name)
	}

	// Validate paths before passing to git subprocess
	if err := validateRepositoryURL(repo.URL); err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}
	if err := validateLocalPath(repo.LocalPath); err != nil {
		return fmt.Errorf("invalid local path: %w", err)
	}

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(repo.LocalPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if repository already exists locally
	if _, err := os.Stat(filepath.Join(repo.LocalPath, ".git")); os.IsNotExist(err) {
		// Repository doesn't exist, clone it
		cmd := exec.Command("git", "clone", repo.URL, repo.LocalPath) // #nosec G204 -- URL and LocalPath validated above
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	} else {
		// Repository exists, pull latest changes
		cmd := exec.Command("git", "-C", repo.LocalPath, "pull") // #nosec G204 -- LocalPath validated above via validateLocalPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to pull repository: %w", err)
		}
	}

	// Update repository information
	if err := rm.updateRepositoryInfo(name); err != nil {
		return err
	}

	return nil
}

// updateRepositoryInfo updates the version information for a repository
func (rm *RepositoryManager) updateRepositoryInfo(name string) error {
	repo, exists := rm.Repositories[name]
	if !exists {
		return fmt.Errorf("repository '%s' not found", name)
	}

	// Validate local path before passing to git subprocess
	if err := validateLocalPath(repo.LocalPath); err != nil {
		return fmt.Errorf("invalid local path: %w", err)
	}

	// Get current commit hash
	cmd := exec.Command("git", "-C", repo.LocalPath, "rev-parse", "HEAD") // #nosec G204 -- LocalPath validated above via validateLocalPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current commit hash: %w", err)
	}
	repo.CurrentVersion = strings.TrimSpace(string(output))

	// Get latest commit hash from remote
	cmd = exec.Command("git", "-C", repo.LocalPath, "ls-remote", "origin", "HEAD") // #nosec G204 -- LocalPath validated above via validateLocalPath
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get latest commit hash: %w", err)
	}
	parts := strings.Split(strings.TrimSpace(string(output)), "\t")
	if len(parts) > 0 {
		repo.LatestVersion = parts[0]
	}

	// Check if updates are available
	repo.HasUpdates = repo.CurrentVersion != repo.LatestVersion
	repo.LastSync = time.Now()

	return nil
}

// GetTemplateVersion returns the version of the templates in a repository
func (rm *RepositoryManager) GetTemplateVersion(name string) (version.Version, error) {
	repo, exists := rm.Repositories[name]
	if !exists {
		return version.Version{}, fmt.Errorf("repository '%s' not found", name)
	}

	// Check for version file in repository
	versionFilePath := filepath.Join(repo.LocalPath, "VERSION")
	if _, err := os.Stat(versionFilePath); err == nil {
		// Read version from file
		versionBytes, err := os.ReadFile(filepath.Clean(versionFilePath))
		if err != nil {
			return version.Version{}, fmt.Errorf("failed to read version file: %w", err)
		}

		// Parse version
		v, err := version.ParseVersion(strings.TrimSpace(string(versionBytes)))
		if err != nil {
			return version.Version{}, fmt.Errorf("failed to parse version: %w", err)
		}

		return v, nil
	}

	// If no version file, use short commit hash as version
	if len(repo.CurrentVersion) >= 7 {
		// Use 0.0.0+<short commit hash> as version
		return version.Version{
			Major: 0,
			Minor: 0,
			Patch: 0,
			Build: repo.CurrentVersion[:7],
		}, nil
	}

	return version.Version{}, fmt.Errorf("could not determine template version")
}

// ListRepositories returns a list of all repositories
func (rm *RepositoryManager) ListRepositories() []*RepositoryInfo {
	repos := make([]*RepositoryInfo, 0, len(rm.Repositories))
	for _, repo := range rm.Repositories {
		repos = append(repos, repo)
	}
	return repos
}

// GetRepository returns a repository by name
func (rm *RepositoryManager) GetRepository(name string) (*RepositoryInfo, error) {
	repo, exists := rm.Repositories[name]
	if !exists {
		return nil, fmt.Errorf("repository '%s' not found", name)
	}
	return repo, nil
}

// RemoveRepository removes a repository
func (rm *RepositoryManager) RemoveRepository(name string) error {
	repo, exists := rm.Repositories[name]
	if !exists {
		return fmt.Errorf("repository '%s' not found", name)
	}

	// Remove local directory if it exists
	if _, err := os.Stat(repo.LocalPath); err == nil {
		if err := os.RemoveAll(repo.LocalPath); err != nil {
			return fmt.Errorf("failed to remove repository directory: %w", err)
		}
	}

	// Remove from map
	delete(rm.Repositories, name)

	return nil
}
