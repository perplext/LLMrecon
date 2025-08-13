package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RepositoryUpdater handles updating from various repository sources
type RepositoryUpdater struct {
	config     *Config
	downloader *Downloader
	verifier   *Verifier
	logger     Logger
}

// NewRepositoryUpdater creates a new repository updater
func NewRepositoryUpdater(config *Config, downloader *Downloader, verifier *Verifier, logger Logger) *RepositoryUpdater {
	return &RepositoryUpdater{
		config:     config,
		downloader: downloader,
		verifier:   verifier,
		logger:     logger,
	}
}

// UpdateFromGitHub updates templates/modules from GitHub repository
func (ru *RepositoryUpdater) UpdateFromGitHub(ctx context.Context, repo RepositoryConfig, targetDir string) ([]string, error) {
	ru.logger.Info(fmt.Sprintf("Updating from GitHub repository: %s", repo.URL))
	
	// Parse GitHub repository URL
	repoInfo, err := ru.parseGitHubURL(repo.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GitHub URL: %w", err)
	}
	
	// Get latest commit info
	latestCommit, err := ru.getLatestGitHubCommit(ctx, repoInfo, repo.Branch, repo.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest commit: %w", err)
	}
	
	// Check if update is needed
	currentCommit := ru.getCurrentCommit(targetDir, repo.Name)
	if currentCommit == latestCommit.SHA {
		ru.logger.Info("Repository is already up to date")
		return []string{}, nil
	}
	
	// Download repository archive
	archiveURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s", 
		repoInfo.Owner, repoInfo.Repo, repo.Branch)
	
	archivePath, err := ru.downloadRepositoryArchive(ctx, archiveURL, repo.Token, repoInfo.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to download repository: %w", err)
	}
	defer os.Remove(archivePath)
	
	// Extract and install
	updatedFiles, err := ru.extractAndInstall(archivePath, targetDir, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to extract and install: %w", err)
	}
	
	// Update commit tracking
	if err := ru.updateCommitTracking(targetDir, repo.Name, latestCommit.SHA); err != nil {
		ru.logger.Warn("Failed to update commit tracking: " + err.Error())
	}
	
	ru.logger.Info(fmt.Sprintf("Successfully updated %d files from %s", len(updatedFiles), repo.Name))
	return updatedFiles, nil
}

// UpdateFromGitLab updates templates/modules from GitLab repository
func (ru *RepositoryUpdater) UpdateFromGitLab(ctx context.Context, repo RepositoryConfig, targetDir string) ([]string, error) {
	ru.logger.Info(fmt.Sprintf("Updating from GitLab repository: %s", repo.URL))
	
	// Parse GitLab repository URL
	repoInfo, err := ru.parseGitLabURL(repo.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GitLab URL: %w", err)
	}
	
	// Get latest commit info
	latestCommit, err := ru.getLatestGitLabCommit(ctx, repoInfo, repo.Branch, repo.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest commit: %w", err)
	}
	
	// Check if update is needed
	currentCommit := ru.getCurrentCommit(targetDir, repo.Name)
	if currentCommit == latestCommit.ID {
		ru.logger.Info("Repository is already up to date")
		return []string{}, nil
	}
	
	// Download repository archive
	archiveURL := fmt.Sprintf("%s/-/archive/%s/%s-%s.zip", 
		repo.URL, repo.Branch, repoInfo.Repo, repo.Branch)
	
	archivePath, err := ru.downloadRepositoryArchive(ctx, archiveURL, repo.Token, repoInfo.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to download repository: %w", err)
	}
	defer os.Remove(archivePath)
	
	// Extract and install
	updatedFiles, err := ru.extractAndInstall(archivePath, targetDir, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to extract and install: %w", err)
	}
	
	// Update commit tracking
	if err := ru.updateCommitTracking(targetDir, repo.Name, latestCommit.ID); err != nil {
		ru.logger.Warn("Failed to update commit tracking: " + err.Error())
	}
	
	ru.logger.Info(fmt.Sprintf("Successfully updated %d files from %s", len(updatedFiles), repo.Name))
	return updatedFiles, nil
}

// UpdateFromHTTP updates from HTTP repository
func (ru *RepositoryUpdater) UpdateFromHTTP(ctx context.Context, repo RepositoryConfig, targetDir string) ([]string, error) {
	ru.logger.Info(fmt.Sprintf("Updating from HTTP repository: %s", repo.URL))
	
	// Download archive from HTTP source
	archivePath, err := ru.downloadRepositoryArchive(ctx, repo.URL, "", repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to download repository: %w", err)
	}
	defer os.Remove(archivePath)
	
	// Extract and install
	updatedFiles, err := ru.extractAndInstall(archivePath, targetDir, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to extract and install: %w", err)
	}
	
	ru.logger.Info(fmt.Sprintf("Successfully updated %d files from %s", len(updatedFiles), repo.Name))
	return updatedFiles, nil
}

// UpdateFromLocal updates from local repository
func (ru *RepositoryUpdater) UpdateFromLocal(ctx context.Context, repo RepositoryConfig, targetDir string) ([]string, error) {
	ru.logger.Info(fmt.Sprintf("Updating from local repository: %s", repo.URL))
	
	// Copy files from local repository
	updatedFiles, err := ru.copyLocalRepository(repo.URL, targetDir, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to copy local repository: %w", err)
	}
	
	ru.logger.Info(fmt.Sprintf("Successfully updated %d files from %s", len(updatedFiles), repo.Name))
	return updatedFiles, nil
}

// downloadRepositoryArchive downloads repository archive
func (ru *RepositoryUpdater) downloadRepositoryArchive(ctx context.Context, url, token, repoName string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", ru.config.UserAgent)
	
	// Add authentication if provided
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	client := &http.Client{Timeout: ru.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download archive: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %d %s", resp.StatusCode, resp.Status)
	}
	
	// Create temporary file
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.zip", repoName))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()
	
	// Download with progress
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to download archive: %w", err)
	}
	
	return tempFile.Name(), nil
}

// extractAndInstall extracts archive and installs files
func (ru *RepositoryUpdater) extractAndInstall(archivePath, targetDir, repoName string) ([]string, error) {
	// Create temporary extraction directory
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-extract-*", repoName))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Extract archive
	if err := ru.extractArchive(archivePath, tempDir); err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}
	
	// Find the extracted repository directory
	extractedDir, err := ru.findExtractedDirectory(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find extracted directory: %w", err)
	}
	
	// Create target directory with repository name
	repoTargetDir := filepath.Join(targetDir, repoName)
	if err := os.MkdirAll(repoTargetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}
	
	// Copy files and track updates
	updatedFiles, err := ru.copyAndTrackFiles(extractedDir, repoTargetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to copy files: %w", err)
	}
	
	return updatedFiles, nil
}

// extractArchive extracts various archive formats
func (ru *RepositoryUpdater) extractArchive(archivePath, destDir string) error {
	ext := strings.ToLower(filepath.Ext(archivePath))
	
	switch ext {
	case ".zip":
		return ru.extractZip(archivePath, destDir)
	case ".tar":
		return ru.extractTar(archivePath, destDir, false)
	case ".gz":
		if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
			return ru.extractTar(archivePath, destDir, true)
		}
		return fmt.Errorf("unsupported archive format: %s", ext)
	default:
		return fmt.Errorf("unsupported archive format: %s", ext)
	}
}

// extractZip extracts a ZIP archive
func (ru *RepositoryUpdater) extractZip(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()
	
	for _, file := range reader.File {
		if err := ru.extractZipFile(file, destDir); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}
	
	return nil
}

// extractZipFile extracts a single file from ZIP
func (ru *RepositoryUpdater) extractZipFile(file *zip.File, destDir string) error {
	// Clean the file path
	cleanPath := filepath.Clean(file.Name)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path: %s", file.Name)
	}
	
	destPath := filepath.Join(destDir, cleanPath)
	
	if file.FileInfo().IsDir() {
		return os.MkdirAll(destPath, file.FileInfo().Mode())
	}
	
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	
	// Extract file
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, reader)
	if err != nil {
		return err
	}
	
	// Set file permissions
	return os.Chmod(destPath, file.FileInfo().Mode())
}

// extractTar extracts a TAR archive
func (ru *RepositoryUpdater) extractTar(archivePath, destDir string, compressed bool) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open tar: %w", err)
	}
	defer file.Close()
	
	var reader io.Reader = file
	
	if compressed {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}
	
	tarReader := tar.NewReader(reader)
	
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}
		
		if err := ru.extractTarEntry(header, tarReader, destDir); err != nil {
			return fmt.Errorf("failed to extract entry %s: %w", header.Name, err)
		}
	}
	
	return nil
}

// extractTarEntry extracts a single entry from TAR
func (ru *RepositoryUpdater) extractTarEntry(header *tar.Header, reader io.Reader, destDir string) error {
	// Clean the file path
	cleanPath := filepath.Clean(header.Name)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path: %s", header.Name)
	}
	
	destPath := filepath.Join(destDir, cleanPath)
	
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(destPath, os.FileMode(header.Mode))
	case tar.TypeReg:
		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		
		// Extract file
		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()
		
		_, err = io.Copy(destFile, reader)
		if err != nil {
			return err
		}
		
		// Set file permissions
		return os.Chmod(destPath, os.FileMode(header.Mode))
	default:
		// Skip other file types (symlinks, etc.)
		return nil
	}
}

// findExtractedDirectory finds the main directory in extracted archive
func (ru *RepositoryUpdater) findExtractedDirectory(tempDir string) (string, error) {
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return "", err
	}
	
	// Look for a single directory (common in archives)
	var mainDir string
	for _, entry := range entries {
		if entry.IsDir() {
			if mainDir == "" {
				mainDir = filepath.Join(tempDir, entry.Name())
			} else {
				// Multiple directories, use temp dir itself
				return tempDir, nil
			}
		}
	}
	
	if mainDir != "" {
		return mainDir, nil
	}
	
	// No directories found, use temp dir
	return tempDir, nil
}

// copyAndTrackFiles copies files and tracks what was updated
func (ru *RepositoryUpdater) copyAndTrackFiles(srcDir, destDir string) ([]string, error) {
	var updatedFiles []string
	
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Calculate relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		
		// Skip hidden files and directories
		if strings.HasPrefix(relPath, ".") || strings.Contains(relPath, "/.") {
			return nil
		}
		
		destPath := filepath.Join(destDir, relPath)
		
		// Check if file needs updating
		needsUpdate, err := ru.fileNeedsUpdate(path, destPath)
		if err != nil {
			return err
		}
		
		if needsUpdate {
			// Create destination directory
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}
			
			// Copy file
			if err := ru.copyFile(path, destPath); err != nil {
				return err
			}
			
			updatedFiles = append(updatedFiles, relPath)
			ru.logger.Debug(fmt.Sprintf("Updated file: %s", relPath))
		}
		
		return nil
	})
	
	return updatedFiles, err
}

// fileNeedsUpdate checks if a file needs updating
func (ru *RepositoryUpdater) fileNeedsUpdate(srcPath, destPath string) (bool, error) {
	// If destination doesn't exist, update needed
	destInfo, err := os.Stat(destPath)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	
	// Get source file info
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false, err
	}
	
	// Compare modification times
	if srcInfo.ModTime().After(destInfo.ModTime()) {
		return true, nil
	}
	
	// Compare file sizes
	if srcInfo.Size() != destInfo.Size() {
		return true, nil
	}
	
	// Files appear to be the same
	return false, nil
}

// copyFile copies a file from source to destination
func (ru *RepositoryUpdater) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	// Copy file permissions
	if info, err := os.Stat(src); err == nil {
		return os.Chmod(dst, info.Mode())
	}
	
	return nil
}

// copyLocalRepository copies from local repository
func (ru *RepositoryUpdater) copyLocalRepository(srcDir, destDir, repoName string) ([]string, error) {
	repoDestDir := filepath.Join(destDir, repoName)
	if err := os.MkdirAll(repoDestDir, 0755); err != nil {
		return nil, err
	}
	
	return ru.copyAndTrackFiles(srcDir, repoDestDir)
}

// Helper functions for repository parsing and commit tracking

// parseGitHubURL parses GitHub repository URL
func (ru *RepositoryUpdater) parseGitHubURL(repoURL string) (*GitHubRepoInfo, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return nil, err
	}
	
	// Remove .git suffix if present
	path := strings.TrimSuffix(u.Path, ".git")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid GitHub URL format")
	}
	
	return &GitHubRepoInfo{
		Owner: parts[0],
		Repo:  parts[1],
	}, nil
}

// parseGitLabURL parses GitLab repository URL
func (ru *RepositoryUpdater) parseGitLabURL(repoURL string) (*GitLabRepoInfo, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return nil, err
	}
	
	// Remove .git suffix if present
	path := strings.TrimSuffix(u.Path, ".git")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid GitLab URL format")
	}
	
	return &GitLabRepoInfo{
		Host:  u.Host,
		Owner: parts[0],
		Repo:  parts[1],
	}, nil
}

// getLatestGitHubCommit gets the latest commit from GitHub
func (ru *RepositoryUpdater) getLatestGitHubCommit(ctx context.Context, repoInfo *GitHubRepoInfo, branch, token string) (*GitHubCommit, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", 
		repoInfo.Owner, repoInfo.Repo, branch)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", ru.config.UserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	client := &http.Client{Timeout: ru.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}
	
	var commit GitHubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return nil, err
	}
	
	return &commit, nil
}

// getLatestGitLabCommit gets the latest commit from GitLab
func (ru *RepositoryUpdater) getLatestGitLabCommit(ctx context.Context, repoInfo *GitLabRepoInfo, branch, token string) (*GitLabCommit, error) {
	// Encode project path for GitLab API
	projectPath := url.PathEscape(fmt.Sprintf("%s/%s", repoInfo.Owner, repoInfo.Repo))
	url := fmt.Sprintf("https://%s/api/v4/projects/%s/repository/commits/%s", 
		repoInfo.Host, projectPath, branch)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", ru.config.UserAgent)
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	client := &http.Client{Timeout: ru.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %d", resp.StatusCode)
	}
	
	var commit GitLabCommit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return nil, err
	}
	
	return &commit, nil
}

// getCurrentCommit gets the current commit hash for a repository
func (ru *RepositoryUpdater) getCurrentCommit(targetDir, repoName string) string {
	commitFile := filepath.Join(targetDir, repoName, ".commit")
	if data, err := os.ReadFile(commitFile); err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}

// updateCommitTracking updates the commit tracking file
func (ru *RepositoryUpdater) updateCommitTracking(targetDir, repoName, commit string) error {
	commitFile := filepath.Join(targetDir, repoName, ".commit")
	commitDir := filepath.Dir(commitFile)
	
	if err := os.MkdirAll(commitDir, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(commitFile, []byte(commit), 0644)
}

// Data structures for repository information

type GitHubRepoInfo struct {
	Owner string
	Repo  string
}

type GitLabRepoInfo struct {
	Host  string
	Owner string
	Repo  string
}

type GitHubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

type GitLabCommit struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	AuthoredDate time.Time `json:"authored_date"`
}