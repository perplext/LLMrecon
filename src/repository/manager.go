package repository

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages multiple repositories
type Manager struct {
	// repositories is a map of repository name to repository instance
	repositories map[string]Repository
	
	// mutex protects the repositories map
	mutex sync.RWMutex

// NewManager creates a new repository manager
func NewManager() *Manager {
	return &Manager{
		repositories: make(map[string]Repository),
	}

// AddRepository adds a repository to the manager
func (m *Manager) AddRepository(repo Repository) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if repository with the same name already exists
	if _, exists := m.repositories[repo.GetName()]; exists {
		return fmt.Errorf("repository with name '%s' already exists", repo.GetName())
	}
	
	// Add repository
	m.repositories[repo.GetName()] = repo
	
	return nil

// RemoveRepository removes a repository from the manager
func (m *Manager) RemoveRepository(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if repository exists
	repo, exists := m.repositories[name]
	if !exists {
		return fmt.Errorf("repository with name '%s' does not exist", name)
	}
	
	// Disconnect repository if connected
	if repo.IsConnected() {
		if err := repo.Disconnect(); err != nil {
			return fmt.Errorf("failed to disconnect repository '%s': %w", name, err)
		}
	}
	
	// Remove repository
	delete(m.repositories, name)
	
	return nil

// GetRepository gets a repository by name
func (m *Manager) GetRepository(name string) (Repository, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Check if repository exists
	repo, exists := m.repositories[name]
	if !exists {
		return nil, fmt.Errorf("repository with name '%s' does not exist", name)
	}
	
	return repo, nil

// ListRepositories lists all repositories
func (m *Manager) ListRepositories() []Repository {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Create result slice
	result := make([]Repository, 0, len(m.repositories))
	
	// Add repositories to result
	for _, repo := range m.repositories {
		result = append(result, repo)
	}
	
	return result

// ConnectAll connects to all repositories
func (m *Manager) ConnectAll(ctx context.Context) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Connect to all repositories
	for name, repo := range m.repositories {
		if err := repo.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to repository '%s': %w", name, err)
		}
	}
	
	return nil

// DisconnectAll disconnects from all repositories
func (m *Manager) DisconnectAll() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Disconnect from all repositories
	var lastErr error
	for name, repo := range m.repositories {
		if repo.IsConnected() {
			if err := repo.Disconnect(); err != nil {
				lastErr = fmt.Errorf("failed to disconnect from repository '%s': %w", name, err)
			}
		}
	}
	
	return lastErr
	

// CreateRepository creates a new repository from a configuration
func (m *Manager) CreateRepository(config *Config) (Repository, error) {
	// Create repository
	repo, err := Create(config)
	if err != nil {
		return nil, err
	}
	
	// Add repository to manager
	if err := m.AddRepository(repo); err != nil {
		return nil, err
	}
	
	return repo, nil

// FindFile finds a file in all repositories
func (m *Manager) FindFile(ctx context.Context, path string) (Repository, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Check all repositories
	for _, repo := range m.repositories {
		// Skip if not connected
		if !repo.IsConnected() {
			continue
		}
		
		// Check if file exists
		exists, err := repo.FileExists(ctx, path)
		if err != nil {
			continue
		}
		
		if exists {
			return repo, nil
		}
	}
	
	return nil, fmt.Errorf("file '%s' not found in any repository", path)

// FindFiles finds files matching a pattern in all repositories
func (m *Manager) FindFiles(ctx context.Context, pattern string) (map[Repository][]FileInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Create result map
	result := make(map[Repository][]FileInfo)
	
	// Check all repositories
	for _, repo := range m.repositories {
		// Skip if not connected
		if !repo.IsConnected() {
			continue
		}
		
		// List files
		files, err := repo.ListFiles(ctx, pattern)
		if err != nil {
			continue
		}
		
		if len(files) > 0 {
			result[repo] = files
		}
	}
	
	return result, nil

// GetFileFromRepo gets a file from a specific repository
func (m *Manager) GetFileFromRepo(ctx context.Context, repoName, path string) (io.ReadCloser, error) {
	// Get repository
	repo, err := m.GetRepository(repoName)
	if err != nil {
		return nil, err
	}
	
	// Get file
	return repo.GetFile(ctx, path)

// GetFile gets a file from any repository that has it
func (m *Manager) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	// Find repository with the file
	repo, err := m.FindFile(ctx, path)
	if err != nil {
		return nil, err
	}
	
	// Get file
	return repo.GetFile(ctx, path)

// DefaultManager is the default repository manager
var DefaultManager = NewManager()

// AddRepository adds a repository to the default manager
func AddRepository(repo Repository) error {
	return DefaultManager.AddRepository(repo)

// RemoveRepository removes a repository from the default manager
func RemoveRepository(name string) error {
	return DefaultManager.RemoveRepository(name)

// GetRepository gets a repository by name from the default manager
func GetRepository(name string) (Repository, error) {
	return DefaultManager.GetRepository(name)

// ListRepositories lists all repositories in the default manager
func ListRepositories() []Repository {
	return DefaultManager.ListRepositories()

// ConnectAll connects to all repositories in the default manager
func ConnectAll(ctx context.Context) error {
	return DefaultManager.ConnectAll(ctx)

// DisconnectAll disconnects from all repositories in the default manager
func DisconnectAll() error {
	return DefaultManager.DisconnectAll()

// CreateRepository creates a new repository from a configuration and adds it to the default manager
func CreateRepository(config *Config) (Repository, error) {
	return DefaultManager.CreateRepository(config)

// FindFile finds a file in all repositories in the default manager
func FindFile(ctx context.Context, path string) (Repository, error) {
	return DefaultManager.FindFile(ctx, path)

// FindFiles finds files matching a pattern in all repositories in the default manager
func FindFiles(ctx context.Context, pattern string) (map[Repository][]FileInfo, error) {
	return DefaultManager.FindFiles(ctx, pattern)

// GetFileFromRepo gets a file from a specific repository in the default manager
func GetFileFromRepo(ctx context.Context, repoName, path string) (io.ReadCloser, error) {
	return DefaultManager.GetFileFromRepo(ctx, repoName, path)

// GetFile gets a file from any repository that has it in the default manager
func GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return DefaultManager.GetFile(ctx, path)
