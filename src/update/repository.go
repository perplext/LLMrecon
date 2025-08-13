package update

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/perplext/LLMrecon/src/version"
)

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

	// Create repository info
	repo := &RepositoryInfo{
		Type:     repoType,
		URL:      url,
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

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(repo.LocalPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if repository already exists locally
	if _, err := os.Stat(filepath.Join(repo.LocalPath, ".git")); os.IsNotExist(err) {
		// Repository doesn't exist, clone it
		cmd := exec.Command("git", "clone", repo.URL, repo.LocalPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	} else {
		// Repository exists, pull latest changes
		cmd := exec.Command("git", "-C", repo.LocalPath, "pull")
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

	// Get current commit hash
	cmd := exec.Command("git", "-C", repo.LocalPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current commit hash: %w", err)
	}
	repo.CurrentVersion = strings.TrimSpace(string(output))

	// Get latest commit hash from remote
	cmd = exec.Command("git", "-C", repo.LocalPath, "ls-remote", "origin", "HEAD")
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
		versionBytes, err := os.ReadFile(versionFilePath)
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
			Major:     0,
			Minor:     0,
			Patch:     0,
			Build:     repo.CurrentVersion[:7],
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
