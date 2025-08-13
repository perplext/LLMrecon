package sandbox

import (
	"context"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
)

// ExecutionMode defines how templates are executed in the sandbox
type ExecutionMode string

const (
	// ModeStrict runs templates with maximum restrictions
	ModeStrict ExecutionMode = "strict"
	// ModeStandard runs templates with standard restrictions
	ModeStandard ExecutionMode = "standard"
	// ModeDevelopment runs templates with minimal restrictions (for development only)
	ModeDevelopment ExecutionMode = "development"
)

// ResourceLimits defines resource limits for template execution
type ResourceLimits struct {
	// MaxCPUTime is the maximum CPU time in seconds
	MaxCPUTime float64
	// MaxMemory is the maximum memory in MB
	MaxMemory int64
	// MaxExecutionTime is the maximum execution time
	MaxExecutionTime time.Duration
	// MaxFileSize is the maximum file size in bytes
	MaxFileSize int64
	// MaxOpenFiles is the maximum number of open files
	MaxOpenFiles int
	// MaxProcesses is the maximum number of processes
	MaxProcesses int
	// NetworkAccess determines if network access is allowed
	NetworkAccess bool
	// FileSystemAccess determines if file system access is allowed
	FileSystemAccess bool
}

// DefaultResourceLimits returns the default resource limits
func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		MaxCPUTime:       1.0,
		MaxMemory:        100,
		MaxExecutionTime: 5 * time.Second,
		MaxFileSize:      1024 * 1024, // 1MB
		MaxOpenFiles:     10,
		MaxProcesses:     1,
		NetworkAccess:    false,
		FileSystemAccess: false,
	}
}

// SandboxOptions defines options for the template sandbox
type SandboxOptions struct {
	// Mode is the execution mode
	Mode ExecutionMode
	// ResourceLimits defines resource limits for template execution
	ResourceLimits ResourceLimits
	// AllowedFunctions is a list of allowed functions
	AllowedFunctions []string
	// AllowedPackages is a list of allowed packages
	AllowedPackages []string
	// DisallowedFunctions is a list of disallowed functions
	DisallowedFunctions []string
	// DisallowedPackages is a list of disallowed packages
	DisallowedPackages []string
	// TimeoutDuration is the timeout duration for template execution
	TimeoutDuration time.Duration
	// EnableLogging enables logging of template execution
	EnableLogging bool
	// LogDirectory is the directory for logs
	LogDirectory string
	// ValidationOptions are the options for template validation
	ValidationOptions *security.VerificationOptions
}

// DefaultSandboxOptions returns the default sandbox options
func DefaultSandboxOptions() *SandboxOptions {
	return &SandboxOptions{
		Mode:               ModeStandard,
		ResourceLimits:     DefaultResourceLimits(),
		AllowedFunctions:   []string{},
		AllowedPackages:    []string{},
		DisallowedFunctions: []string{
			"os.Exit",
			"syscall",
			"unsafe",
			"runtime.SetFinalizer",
		},
		DisallowedPackages: []string{
			"os/exec",
			"syscall",
			"unsafe",
		},
		TimeoutDuration:    10 * time.Second,
		EnableLogging:      true,
		LogDirectory:       "",
		ValidationOptions:  security.DefaultVerificationOptions(),
	}
}

// ExecutionResult represents the result of template execution in the sandbox
type ExecutionResult struct {
	// Success indicates if the execution was successful
	Success bool
	// Output is the output of the execution
	Output string
	// Error is the error message if execution failed
	Error string
	// ExecutionTime is the execution time
	ExecutionTime time.Duration
	// ResourceUsage contains information about resource usage
	ResourceUsage ResourceUsage
	// SecurityIssues contains security issues found during execution
	SecurityIssues []*security.SecurityIssue
}

// ResourceUsage contains information about resource usage during execution
type ResourceUsage struct {
	// CPUTime is the CPU time used in seconds
	CPUTime float64
	// MemoryUsage is the memory used in MB
	MemoryUsage int64
	// ExecutionTime is the execution time
	ExecutionTime time.Duration
	// FileOperations is the number of file operations
	FileOperations int
	// NetworkOperations is the number of network operations
	NetworkOperations int
}

// TemplateSandbox is the interface for template sandboxes
type TemplateSandbox interface {
	// Execute executes a template in the sandbox
	Execute(ctx context.Context, template *format.Template, options *SandboxOptions) (*ExecutionResult, error)
	
	// ExecuteFile executes a template file in the sandbox
	ExecuteFile(ctx context.Context, templatePath string, options *SandboxOptions) (*ExecutionResult, error)
	
	// Validate validates a template against security rules
	Validate(ctx context.Context, template *format.Template, options *SandboxOptions) ([]*security.SecurityIssue, error)
	
	// ValidateFile validates a template file against security rules
	ValidateFile(ctx context.Context, templatePath string, options *SandboxOptions) ([]*security.SecurityIssue, error)
	
	// GetAllowList returns the allow list for template execution
	GetAllowList() *AllowList
	
	// SetAllowList sets the allow list for template execution
	SetAllowList(allowList *AllowList)
}

// AllowList defines allowed operations for template execution
type AllowList struct {
	// Functions is a list of allowed functions
	Functions []string
	// Packages is a list of allowed packages
	Packages []string
	// FilePatterns is a list of allowed file patterns
	FilePatterns []string
	// NetworkDomains is a list of allowed network domains
	NetworkDomains []string
	// EnvironmentVariables is a list of allowed environment variables
	EnvironmentVariables []string
}

// NewAllowList creates a new allow list with default values
func NewAllowList() *AllowList {
	return &AllowList{
		Functions: []string{
			"fmt.Print",
			"fmt.Printf",
			"fmt.Println",
			"strings.Join",
			"strings.Split",
			"strings.Replace",
			"strings.ToLower",
			"strings.ToUpper",
		},
		Packages: []string{
			"fmt",
			"strings",
			"time",
			"math",
			"encoding/json",
		},
		FilePatterns: []string{
			"*.txt",
			"*.json",
			"*.yaml",
			"*.yml",
		},
		NetworkDomains: []string{},
		EnvironmentVariables: []string{
			"PATH",
			"HOME",
			"USER",
		},
	}
}
