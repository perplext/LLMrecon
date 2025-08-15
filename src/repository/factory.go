package repository

import (
	"fmt"
)

// Factory creates repository instances based on configuration
type Factory struct {
	// registeredTypes maps repository types to their creator functions
	registeredTypes map[RepositoryType]RepositoryCreator

// RepositoryCreator is a function that creates a repository instance
type RepositoryCreator func(config *Config) (Repository, error)

// NewFactory creates a new repository factory
func NewFactory() *Factory {
	return &Factory{
		registeredTypes: make(map[RepositoryType]RepositoryCreator),
	}

// Register registers a repository creator for a specific type
func (f *Factory) Register(repoType RepositoryType, creator RepositoryCreator) {
	f.registeredTypes[repoType] = creator

// Create creates a repository instance based on the configuration
func (f *Factory) Create(config *Config) (Repository, error) {
	creator, exists := f.registeredTypes[config.Type]
	if !exists {
		return nil, fmt.Errorf("unsupported repository type: %s", config.Type)
	}
	
	return creator(config)

// DefaultFactory is the default repository factory with standard repository types registered
var DefaultFactory = NewFactory()

// init registers the standard repository types with the default factory
func init() {
	// Register all repository types
	DefaultFactory.Register(GitHub, NewGitHubRepository)
	DefaultFactory.Register(GitLab, NewGitLabRepository)
	DefaultFactory.Register(LocalFS, NewLocalFSRepository)
	DefaultFactory.Register(HTTP, NewHTTPRepository)
	DefaultFactory.Register(Database, NewDatabaseRepository)
	DefaultFactory.Register(S3, NewS3Repository)

// Create creates a repository instance using the default factory
func Create(config *Config) (Repository, error) {
	return DefaultFactory.Create(config)
