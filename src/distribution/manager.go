package distribution

import (
	"context"
	"fmt"
	"runtime"
)

// DistributionManager orchestrates cross-platform CLI distribution
type DistributionManager struct {
	config           *Config
	buildPipeline    BuildPipeline
	packageManagers  map[string]PackageManager
	channels         map[string]DistributionChannel
	verifier         UpdateVerifier
	analytics        InstallationAnalytics
	releaseManager   ReleaseManager
	logger           Logger

// Config defines distribution configuration
type Config struct {
	// Build configuration
	BuildTargets         []BuildTarget         `json:"build_targets"`
	ArtifactStorage      ArtifactStorageConfig `json:"artifact_storage"`
	SigningConfig        SigningConfig         `json:"signing_config"`
	
	// Distribution configuration
	Channels             []ChannelConfig       `json:"channels"`
	PackageManagers      []PackageManagerConfig `json:"package_managers"`
	
	// Update configuration
	UpdateServer         UpdateServerConfig    `json:"update_server"`
	VerificationConfig   VerificationConfig    `json:"verification_config"`
	
	// Analytics configuration
	Analytics            AnalyticsConfig       `json:"analytics"`
	
	// Release configuration
	ReleaseStrategy      ReleaseStrategy       `json:"release_strategy"`
	RollbackConfig       RollbackConfig        `json:"rollback_config"`
}

// BuildTarget defines a specific platform/architecture build target
type BuildTarget struct {
	Platform     Platform     `json:"platform"`
	Architecture Architecture `json:"architecture"`
	GoOS         string       `json:"goos"`
	GoArch       string       `json:"goarch"`
	CGOEnabled   bool         `json:"cgo_enabled"`
	BuildFlags   []string     `json:"build_flags"`
	LDFlags      []string     `json:"ldflags"`
	OutputName   string       `json:"output_name"`
	Compress     bool         `json:"compress"`
	Notarize     bool         `json:"notarize"`     // For macOS

// Platform represents target platforms
type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
	PlatformMacOS   Platform = "darwin"
	PlatformFreeBSD Platform = "freebsd"
	PlatformOpenBSD Platform = "openbsd"
	PlatformNetBSD  Platform = "netbsd"
)

// Architecture represents target architectures
type Architecture string

const (
	ArchAMD64 Architecture = "amd64"
	ArchARM64 Architecture = "arm64"
	ArchARM   Architecture = "arm"
	Arch386   Architecture = "386"
)

// ArtifactStorageConfig defines where build artifacts are stored
type ArtifactStorageConfig struct {
	Type        StorageType `json:"type"`
	S3Config    S3Config    `json:"s3_config,omitempty"`
	GCSConfig   GCSConfig   `json:"gcs_config,omitempty"`
	LocalConfig LocalConfig `json:"local_config,omitempty"`
	Retention   time.Duration `json:"retention"`
}

// StorageType represents different storage backends
type StorageType string

const (
	StorageTypeS3    StorageType = "s3"
	StorageTypeGCS   StorageType = "gcs"
	StorageTypeLocal StorageType = "local"
)

// SigningConfig defines code signing configuration
type SigningConfig struct {
	Enabled      bool              `json:"enabled"`
	KeyPath      string            `json:"key_path"`
	CertPath     string            `json:"cert_path"`
	Algorithm    string            `json:"algorithm"`
	Timestamping bool              `json:"timestamping"`
	Platforms    map[Platform]PlatformSigningConfig `json:"platforms"`

// PlatformSigningConfig defines platform-specific signing
type PlatformSigningConfig struct {
	Enabled       bool   `json:"enabled"`
	Identity      string `json:"identity"`       // Code signing identity
	Entitlements  string `json:"entitlements"`   // macOS entitlements file
	BundleID      string `json:"bundle_id"`      // macOS bundle identifier
	TeamID        string `json:"team_id"`        // Apple Developer Team ID
}

// ChannelConfig defines a distribution channel
type ChannelConfig struct {
	Name        string                 `json:"name"`
	Type        ChannelType            `json:"type"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Platforms   []Platform             `json:"platforms"`
	Config      map[string]interface{} `json:"config"`
	UpdateFreq  time.Duration          `json:"update_frequency"`
}

// ChannelType represents different distribution channels
type ChannelType string

const (
	ChannelTypeGitHub      ChannelType = "github"
	ChannelTypeGitLab      ChannelType = "gitlab"
	ChannelTypeHomebrew    ChannelType = "homebrew"
	ChannelTypeChocolatey  ChannelType = "chocolatey"
	ChannelTypeAPT         ChannelType = "apt"
	ChannelTypeRPM         ChannelType = "rpm"
	ChannelTypeSnap        ChannelType = "snap"
	ChannelTypeFlatpak     ChannelType = "flatpak"
	ChannelTypeDockerHub   ChannelType = "dockerhub"
	ChannelTypeAUR         ChannelType = "aur"
	ChannelTypeWinget      ChannelType = "winget"
	ChannelTypeScoop       ChannelType = "scoop"
)

// PackageManagerConfig defines package manager integration
type PackageManagerConfig struct {
	Name        string                 `json:"name"`
	Type        PackageManagerType     `json:"type"`
	Enabled     bool                   `json:"enabled"`
	Repository  string                 `json:"repository"`
	Credentials map[string]string      `json:"credentials"`
	Metadata    PackageMetadata        `json:"metadata"`
	AutoUpdate  bool                   `json:"auto_update"`
}

// PackageManagerType represents different package managers
type PackageManagerType string

const (
	PackageManagerHomebrew   PackageManagerType = "homebrew"
	PackageManagerChocolatey PackageManagerType = "chocolatey"
	PackageManagerAPT        PackageManagerType = "apt"
	PackageManagerRPM        PackageManagerType = "rpm"
	PackageManagerSnap       PackageManagerType = "snap"
	PackageManagerFlatpak    PackageManagerType = "flatpak"
	PackageManagerWinget     PackageManagerType = "winget"
	PackageManagerScoop      PackageManagerType = "scoop"
	PackageManagerAUR        PackageManagerType = "aur"
	PackageManagerNPM        PackageManagerType = "npm"
	PackageManagerPyPI       PackageManagerType = "pypi"
)

// PackageMetadata defines package metadata
type PackageMetadata struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Homepage     string            `json:"homepage"`
	License      string            `json:"license"`
	Authors      []string          `json:"authors"`
	Keywords     []string          `json:"keywords"`
	Categories   []string          `json:"categories"`
	Dependencies []Dependency      `json:"dependencies"`
	Conflicts    []string          `json:"conflicts"`
	Provides     []string          `json:"provides"`
	Replaces     []string          `json:"replaces"`
	Extras       map[string]interface{} `json:"extras"`
}

// Dependency represents a package dependency
type Dependency struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Optional    bool   `json:"optional"`
	Platform    string `json:"platform,omitempty"`

// UpdateServerConfig defines update server configuration
type UpdateServerConfig struct {
	Enabled     bool   `json:"enabled"`
	BaseURL     string `json:"base_url"`
	Port        int    `json:"port"`
	TLSEnabled  bool   `json:"tls_enabled"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
	RateLimit   int    `json:"rate_limit"`
}

// VerificationConfig defines update verification settings
type VerificationConfig struct {
	Required       bool     `json:"required"`
	ChecksumAlgo   string   `json:"checksum_algorithm"`
	SignatureAlgo  string   `json:"signature_algorithm"`
	TrustedKeys    []string `json:"trusted_keys"`
	CertPinning    bool     `json:"cert_pinning"`
	PinnedCerts    []string `json:"pinned_certs"`

// AnalyticsConfig defines installation analytics configuration
type AnalyticsConfig struct {
	Enabled         bool   `json:"enabled"`
	Endpoint        string `json:"endpoint"`
	CollectUsage    bool   `json:"collect_usage"`
	CollectErrors   bool   `json:"collect_errors"`
	CollectTelemetry bool  `json:"collect_telemetry"`
	RetentionDays   int    `json:"retention_days"`
	AnonymizeIPs    bool   `json:"anonymize_ips"`
}

// ReleaseStrategy defines release management strategy
type ReleaseStrategy struct {
	Type           ReleaseType     `json:"type"`
	Channels       []ReleaseChannel `json:"channels"`
	RolloutPercent int             `json:"rollout_percent"`
	CanaryDuration time.Duration   `json:"canary_duration"`
	AutoPromote    bool            `json:"auto_promote"`
	HealthChecks   []HealthCheck   `json:"health_checks"`

// ReleaseType represents different release strategies
type ReleaseType string

const (
	ReleaseTypeImmediate ReleaseType = "immediate"
	ReleaseTypeStaged    ReleaseType = "staged"
	ReleaseTypeCanary    ReleaseType = "canary"
	ReleaseTypeBlueGreen ReleaseType = "blue_green"
)

// ReleaseChannel represents different release channels
type ReleaseChannel struct {
	Name        string        `json:"name"`
	Stability   StabilityLevel `json:"stability"`
	Audience    string        `json:"audience"`
	Percentage  int           `json:"percentage"`
	Requirements []string     `json:"requirements"`
}

// StabilityLevel represents release stability
type StabilityLevel string

const (
	StabilityAlpha  StabilityLevel = "alpha"
	StabilityBeta   StabilityLevel = "beta"
	StabilityRC     StabilityLevel = "rc"
	StabilityStable StabilityLevel = "stable"
)

// HealthCheck defines release health verification
type HealthCheck struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Endpoint    string        `json:"endpoint"`
	Interval    time.Duration `json:"interval"`
	Timeout     time.Duration `json:"timeout"`
	Threshold   int           `json:"threshold"`
	Required    bool          `json:"required"`
}

// RollbackConfig defines rollback configuration
type RollbackConfig struct {
	Enabled         bool          `json:"enabled"`
	AutoRollback    bool          `json:"auto_rollback"`
	TriggerThreshold float64      `json:"trigger_threshold"`
	MaxVersions     int           `json:"max_versions"`
	RollbackWindow  time.Duration `json:"rollback_window"`
}

// NewDistributionManager creates a new distribution manager
func NewDistributionManager(config *Config, logger Logger) *DistributionManager {
	manager := &DistributionManager{
		config:          config,
		packageManagers: make(map[string]PackageManager),
		channels:        make(map[string]DistributionChannel),
		logger:          logger,
	}
	
	// Initialize components
	manager.buildPipeline = NewBuildPipeline(config, logger)
	manager.verifier = NewUpdateVerifier(config.VerificationConfig, logger)
	manager.analytics = NewInstallationAnalytics(config.Analytics, logger)
	manager.releaseManager = NewReleaseManager(config.ReleaseStrategy, logger)
	
	// Register package managers
	manager.registerPackageManagers()
	
	// Register distribution channels
	manager.registerDistributionChannels()
	
	return manager

// GetSupportedPlatforms returns list of supported platforms
func (dm *DistributionManager) GetSupportedPlatforms() []PlatformInfo {
	var platforms []PlatformInfo
	
	for _, target := range dm.config.BuildTargets {
		platforms = append(platforms, PlatformInfo{
			Platform:     target.Platform,
			Architecture: target.Architecture,
			GoOS:         target.GoOS,
			GoArch:       target.GoArch,
			Available:    true,
		})
	}
	
	return platforms

// GetCurrentPlatform returns information about the current platform
func (dm *DistributionManager) GetCurrentPlatform() PlatformInfo {
	return PlatformInfo{
		Platform:     Platform(runtime.GOOS),
		Architecture: Architecture(runtime.GOARCH),
		GoOS:         runtime.GOOS,
		GoArch:       runtime.GOARCH,
		Available:    true,
	}

// GetAvailableChannels returns list of available distribution channels
func (dm *DistributionManager) GetAvailableChannels() []ChannelInfo {
	var channels []ChannelInfo
	
	for _, config := range dm.config.Channels {
		if config.Enabled {
			channels = append(channels, ChannelInfo{
				Name:        config.Name,
				Type:        config.Type,
				Platforms:   config.Platforms,
				Priority:    config.Priority,
				UpdateFreq:  config.UpdateFreq,
			})
		}
	}
	
	return channels

// GetPackageManagers returns list of supported package managers
func (dm *DistributionManager) GetPackageManagers() []PackageManagerInfo {
	var managers []PackageManagerInfo
	
	for _, config := range dm.config.PackageManagers {
		if config.Enabled {
			managers = append(managers, PackageManagerInfo{
				Name:       config.Name,
				Type:       config.Type,
				Repository: config.Repository,
				AutoUpdate: config.AutoUpdate,
			})
		}
	}
	
	return managers

// ValidateDistributionConfig validates the distribution configuration
func (dm *DistributionManager) ValidateDistributionConfig() []ValidationError {
	var errors []ValidationError
	
	// Validate build targets
	if len(dm.config.BuildTargets) == 0 {
		errors = append(errors, ValidationError{
			Field:   "build_targets",
			Message: "At least one build target must be specified",
		})
	}
	
	// Validate channels
	for i, channel := range dm.config.Channels {
		if channel.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("channels[%d].name", i),
				Message: "Channel name cannot be empty",
			})
		}
		
		if len(channel.Platforms) == 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("channels[%d].platforms", i),
				Message: "At least one platform must be specified for channel",
			})
		}
	}
	
	// Validate package managers
	for i, pm := range dm.config.PackageManagers {
		if pm.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("package_managers[%d].name", i),
				Message: "Package manager name cannot be empty",
			})
		}
		
		if pm.Repository == "" && pm.Type != PackageManagerAUR {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("package_managers[%d].repository", i),
				Message: "Repository must be specified for package manager",
			})
		}
	}
	
	// Validate signing config
	if dm.config.SigningConfig.Enabled {
		if dm.config.SigningConfig.KeyPath == "" {
			errors = append(errors, ValidationError{
				Field:   "signing_config.key_path",
				Message: "Key path must be specified when signing is enabled",
			})
		}
	}
	
	return errors

// GetDistributionMatrix returns the distribution matrix
func (dm *DistributionManager) GetDistributionMatrix() DistributionMatrix {
	matrix := DistributionMatrix{
		Platforms: make(map[Platform][]Architecture),
		Channels:  make(map[ChannelType][]Platform),
		Packages:  make(map[PackageManagerType][]Platform),
	}
	
	// Build platform matrix
	for _, target := range dm.config.BuildTargets {
		if _, exists := matrix.Platforms[target.Platform]; !exists {
			matrix.Platforms[target.Platform] = make([]Architecture, 0)
		}
		matrix.Platforms[target.Platform] = append(matrix.Platforms[target.Platform], target.Architecture)
	}
	
	// Build channel matrix
	for _, channel := range dm.config.Channels {
		if channel.Enabled {
			matrix.Channels[channel.Type] = channel.Platforms
		}
	}
	
	// Build package manager matrix
	for _, pm := range dm.config.PackageManagers {
		if pm.Enabled {
			platforms := dm.getPackageManagerPlatforms(pm.Type)
			matrix.Packages[pm.Type] = platforms
		}
	}
	
	return matrix

// Internal methods

func (dm *DistributionManager) registerPackageManagers() {
	for _, config := range dm.config.PackageManagers {
		if !config.Enabled {
			continue
		}
		
		var pm PackageManager
		switch config.Type {
		case PackageManagerHomebrew:
			pm = NewHomebrewManager(config, dm.logger)
		case PackageManagerChocolatey:
			pm = NewChocolateyManager(config, dm.logger)
		case PackageManagerAPT:
			pm = NewAPTManager(config, dm.logger)
		case PackageManagerRPM:
			pm = NewRPMManager(config, dm.logger)
		case PackageManagerSnap:
			pm = NewSnapManager(config, dm.logger)
		case PackageManagerWinget:
			pm = NewWingetManager(config, dm.logger)
		case PackageManagerScoop:
			pm = NewScoopManager(config, dm.logger)
		default:
			dm.logger.Warn("Unsupported package manager type", "type", config.Type)
			continue
		}
		
		dm.packageManagers[config.Name] = pm
		dm.logger.Info("Registered package manager", "name", config.Name, "type", config.Type)
	}

func (dm *DistributionManager) registerDistributionChannels() {
	for _, config := range dm.config.Channels {
		if !config.Enabled {
			continue
		}
		
		var channel DistributionChannel
		switch config.Type {
		case ChannelTypeGitHub:
			channel = NewGitHubChannel(config, dm.logger)
		case ChannelTypeGitLab:
			channel = NewGitLabChannel(config, dm.logger)
		case ChannelTypeDockerHub:
			channel = NewDockerHubChannel(config, dm.logger)
		default:
			dm.logger.Warn("Unsupported channel type", "type", config.Type)
			continue
		}
		
		dm.channels[config.Name] = channel
		dm.logger.Info("Registered distribution channel", "name", config.Name, "type", config.Type)
	}

func (dm *DistributionManager) getPackageManagerPlatforms(pmType PackageManagerType) []Platform {
	switch pmType {
	case PackageManagerHomebrew:
		return []Platform{PlatformMacOS, PlatformLinux}
	case PackageManagerChocolatey, PackageManagerWinget, PackageManagerScoop:
		return []Platform{PlatformWindows}
	case PackageManagerAPT, PackageManagerRPM, PackageManagerSnap, PackageManagerFlatpak, PackageManagerAUR:
		return []Platform{PlatformLinux}
	default:
		return []Platform{PlatformLinux, PlatformMacOS, PlatformWindows}
	}

// Supporting types

type PlatformInfo struct {
	Platform     Platform     `json:"platform"`
	Architecture Architecture `json:"architecture"`
	GoOS         string       `json:"goos"`
	GoArch       string       `json:"goarch"`
	Available    bool         `json:"available"`
}

type ChannelInfo struct {
	Name       string        `json:"name"`
	Type       ChannelType   `json:"type"`
	Platforms  []Platform    `json:"platforms"`
	Priority   int           `json:"priority"`
	UpdateFreq time.Duration `json:"update_frequency"`
}

type PackageManagerInfo struct {
	Name       string             `json:"name"`
	Type       PackageManagerType `json:"type"`
	Repository string             `json:"repository"`
	AutoUpdate bool               `json:"auto_update"`

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`

type DistributionMatrix struct {
	Platforms map[Platform][]Architecture          `json:"platforms"`
	Channels  map[ChannelType][]Platform           `json:"channels"`
	Packages  map[PackageManagerType][]Platform    `json:"packages"`
}

// Storage configurations
type S3Config struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Prefix    string `json:"prefix"`
}

type GCSConfig struct {
	Bucket      string `json:"bucket"`
	ProjectID   string `json:"project_id"`
	Credentials string `json:"credentials"`
	Prefix      string `json:"prefix"`

type LocalConfig struct {
	Path string `json:"path"`
