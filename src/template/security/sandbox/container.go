package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
)

// ContainerSandbox is a sandbox implementation that uses containers for isolation
type ContainerSandbox struct {
	DefaultSandbox
	containerEngine string // "docker" or "podman"
	imageName       string
	networkMode     string
	volumeBinds     []string
}

// ContainerSandboxOptions contains options specific to container-based sandboxes
type ContainerSandboxOptions struct {
	ContainerEngine string   // "docker" or "podman"
	ImageName       string   // Container image to use
	NetworkMode     string   // Network mode (none, host, bridge)
	VolumeBinds     []string // Volume bindings
	CleanupTimeout  time.Duration
}

// DefaultContainerSandboxOptions returns the default container sandbox options
func DefaultContainerSandboxOptions() *ContainerSandboxOptions {
	return &ContainerSandboxOptions{
		ContainerEngine: "docker",
		ImageName:       "alpine:latest",
		NetworkMode:     "none",
		VolumeBinds:     []string{},
		CleanupTimeout:  5 * time.Second,
	}

// NewContainerSandbox creates a new container-based sandbox
func NewContainerSandbox(verifier security.TemplateVerifier, options *SandboxOptions, containerOptions *ContainerSandboxOptions) (*ContainerSandbox, error) {
	if options == nil {
		options = DefaultSandboxOptions()
	}
	
	if containerOptions == nil {
		containerOptions = DefaultContainerSandboxOptions()
	}
	
	// Check if container engine is available
	if err := checkContainerEngine(containerOptions.ContainerEngine); err != nil {
		return nil, fmt.Errorf("container engine not available: %w", err)
	}
	
	// Create the default sandbox
	defaultSandbox := NewSandbox(verifier, options)
	
	return &ContainerSandbox{
		DefaultSandbox:  *defaultSandbox,
		containerEngine: containerOptions.ContainerEngine,
		imageName:       containerOptions.ImageName,
		networkMode:     containerOptions.NetworkMode,
		volumeBinds:     containerOptions.VolumeBinds,
	}, nil

// checkContainerEngine checks if the specified container engine is available
func checkContainerEngine(engine string) error {
	var cmd *exec.Cmd
	
	switch engine {
	case "docker":
		cmd = exec.Command("docker", "version")
	case "podman":
		cmd = exec.Command("podman", "version")
	default:
		return fmt.Errorf("unsupported container engine: %s", engine)
	}
	
	return cmd.Run()

// Execute executes a template in a container sandbox
func (s *ContainerSandbox) Execute(ctx context.Context, template *format.Template, options *SandboxOptions) (*ExecutionResult, error) {
	if options == nil {
		options = s.options
	}
	
	// Validate the template first
	issues, err := s.Validate(ctx, template, options)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Template validation failed: %v", err),
		}, err
	}
	
	// If there are critical security issues, don't execute the template
	for _, issue := range issues {
		if issue.Severity == "critical" {
			return &ExecutionResult{
				Success:       false,
				Error:         fmt.Sprintf("Critical security issue found: %s", issue.Description),
				SecurityIssues: issues,
			}, fmt.Errorf("critical security issue found: %s", issue.Description)
		}
	}
	
	// Create a context with timeout
	execCtx, cancel := context.WithTimeout(ctx, options.TimeoutDuration)
	defer cancel()
	
	startTime := time.Now()
	
	// Execute the template in a container
	result, err := s.executeInContainer(execCtx, template, options)
	
	executionTime := time.Since(startTime)
	
	if err != nil {
		return &ExecutionResult{
			Success:       false,
			Error:         fmt.Sprintf("Template execution failed: %v", err),
			ExecutionTime: executionTime,
			SecurityIssues: issues,
		}, err
	}
	
	result.ExecutionTime = executionTime
	result.SecurityIssues = issues
	
	return result, nil
	

// executeInContainer executes a template in a container
func (s *ContainerSandbox) executeInContainer(ctx context.Context, template *format.Template, options *SandboxOptions) (*ExecutionResult, error) {
	// Create a temporary directory for container execution
	tempDir, err := ioutil.TempDir("", "template-container-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Write the template to a file
	templateFile := filepath.Join(tempDir, "template.json")
	templateData, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template: %w", err)
	}
	
	if err := ioutil.WriteFile(templateFile, templateData, 0600); err != nil {
		return nil, fmt.Errorf("failed to write template file: %w", err)
	}
	
	// Create a script to execute the template
	scriptFile := filepath.Join(tempDir, "execute.sh")
	scriptContent := `#!/bin/sh
cat template.json | jq -r '.Content' > template.txt
echo "Executing template..."
cat template.txt
echo "Done."
`
	
	if err := ioutil.WriteFile(scriptFile, []byte(scriptContent), 0700); err != nil {
		return nil, fmt.Errorf("failed to write script file: %w", err)
	}
	
	// Build the container command
	var cmd *exec.Cmd
	containerName := fmt.Sprintf("template-sandbox-%d", time.Now().UnixNano())
	
	args := []string{
		"run",
		"--name", containerName,
		"--rm",
		"--network", s.networkMode,
		"--memory", fmt.Sprintf("%dm", options.ResourceLimits.MaxMemory),
		"--memory-swap", fmt.Sprintf("%dm", options.ResourceLimits.MaxMemory),
		"--cpus", fmt.Sprintf("%.2f", options.ResourceLimits.MaxCPUTime),
		"--pids-limit", fmt.Sprintf("%d", options.ResourceLimits.MaxProcesses),
		"-v", fmt.Sprintf("%s:/workspace", tempDir),
		"-w", "/workspace",
	}
	
	// Add volume binds
	for _, bind := range s.volumeBinds {
		args = append(args, "-v", bind)
	}
	
	// Add resource limits
	if !options.ResourceLimits.NetworkAccess {
		args = append(args, "--network", "none")
	}
	
	// Add the image and command
	args = append(args, s.imageName, "/bin/sh", "./execute.sh")
	
	// Create the command
	switch s.containerEngine {
	case "docker":
		cmd = exec.CommandContext(ctx, "docker", args...)
	case "podman":
		cmd = exec.CommandContext(ctx, "podman", args...)
	default:
		return nil, fmt.Errorf("unsupported container engine: %s", s.containerEngine)
	}
	
	// Capture output
	output, err := cmd.CombinedOutput()
	
	// Check if the context is done (timeout or cancellation)
	select {
	case <-ctx.Done():
		// Cleanup the container
		s.cleanupContainer(containerName)
		
		return &ExecutionResult{
			Success: false,
			Error:   "Template execution timed out",
			ResourceUsage: ResourceUsage{
				ExecutionTime: options.TimeoutDuration,
			},
		}, ctx.Err()
	default:
		// Continue
	}
	
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Template execution failed: %v\nOutput: %s", err, string(output)),
			ResourceUsage: ResourceUsage{
				ExecutionTime: time.Since(time.Now().Add(-options.TimeoutDuration)),
			},
		}, err
	}
	
	return &ExecutionResult{
		Success: true,
		Output:  string(output),
		ResourceUsage: ResourceUsage{
			ExecutionTime: time.Since(time.Now().Add(-options.TimeoutDuration)),
		},
	}, nil

// cleanupContainer cleans up a container
func (s *ContainerSandbox) cleanupContainer(containerName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	var cmd *exec.Cmd
	
	switch s.containerEngine {
	case "docker":
		cmd = exec.CommandContext(ctx, "docker", "rm", "-f", containerName)
	case "podman":
		cmd = exec.CommandContext(ctx, "podman", "rm", "-f", containerName)
	default:
		return
	}
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}
}
}
}
}
