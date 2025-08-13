package ui

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Runner provides an interactive CLI runner with progress tracking
type Runner struct {
	terminal *Terminal
	tasks    []RunnerTask
	options  RunnerOptions
}

// RunnerTask represents a task to be executed
type RunnerTask struct {
	ID          string
	Name        string
	Description string
	Execute     func(ctx context.Context, progress ProgressReporter) error
	Retryable   bool
	Required    bool
}

// RunnerOptions configures the runner
type RunnerOptions struct {
	ShowProgress     bool
	ConcurrentTasks  int
	RetryAttempts    int
	RetryDelay       time.Duration
	ContinueOnError  bool
	ShowTaskDetails  bool
	SummaryReport    bool
}

// DefaultRunnerOptions returns default runner options
func DefaultRunnerOptions() RunnerOptions {
	return RunnerOptions{
		ShowProgress:     true,
		ConcurrentTasks:  1,
		RetryAttempts:    3,
		RetryDelay:       time.Second,
		ContinueOnError:  false,
		ShowTaskDetails:  true,
		SummaryReport:    true,
	}
}

// ProgressReporter provides progress reporting interface
type ProgressReporter interface {
	SetTotal(total int64)
	SetCurrent(current int64)
	Increment()
	SetStatus(status string)
	SetDetails(details string)
	AddSubTask(name string)
	CompleteSubTask(name string)
}

// NewRunner creates a new interactive runner
func NewRunner(terminal *Terminal, options RunnerOptions) *Runner {
	return &Runner{
		terminal: terminal,
		tasks:    make([]RunnerTask, 0),
		options:  options,
	}
}

// AddTask adds a task to the runner
func (r *Runner) AddTask(task RunnerTask) {
	r.tasks = append(r.tasks, task)
}

// Run executes all tasks with progress tracking
func (r *Runner) Run(ctx context.Context) error {
	if len(r.tasks) == 0 {
		return fmt.Errorf("no tasks to run")
	}

	r.terminal.Header("Task Execution")
	r.terminal.Info("Starting %d tasks...\n", len(r.tasks))

	startTime := time.Now()
	results := make([]TaskResult, 0, len(r.tasks))

	if r.options.ConcurrentTasks > 1 {
		results = r.runConcurrent(ctx)
	} else {
		results = r.runSequential(ctx)
	}

	// Show summary
	if r.options.SummaryReport {
		r.showSummary(results, time.Since(startTime))
	}

	// Check for failures
	for _, result := range results {
		if result.Error != nil && r.tasks[result.TaskIndex].Required {
			return fmt.Errorf("required task '%s' failed: %w", r.tasks[result.TaskIndex].Name, result.Error)
		}
	}

	return nil
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskIndex int
	Success   bool
	Error     error
	Duration  time.Duration
	Retries   int
}

// runSequential runs tasks sequentially
func (r *Runner) runSequential(ctx context.Context) []TaskResult {
	results := make([]TaskResult, 0, len(r.tasks))

	for i, task := range r.tasks {
		select {
		case <-ctx.Done():
			r.terminal.Warning("Execution cancelled")
			return results
		default:
			result := r.runTask(ctx, i, task)
			results = append(results, result)

			if result.Error != nil && !r.options.ContinueOnError && task.Required {
				r.terminal.Error("Stopping execution due to error")
				return results
			}
		}
	}

	return results
}

// runConcurrent runs tasks concurrently
func (r *Runner) runConcurrent(ctx context.Context) []TaskResult {
	var wg sync.WaitGroup
	results := make([]TaskResult, len(r.tasks))
	semaphore := make(chan struct{}, r.options.ConcurrentTasks)

	for i, task := range r.tasks {
		wg.Add(1)
		go func(index int, t RunnerTask) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				results[index] = TaskResult{
					TaskIndex: index,
					Success:   false,
					Error:     ctx.Err(),
				}
				return
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
				results[index] = r.runTask(ctx, index, t)
			}
		}(i, task)
	}

	wg.Wait()
	return results
}

// runTask runs a single task with retry logic
func (r *Runner) runTask(ctx context.Context, index int, task RunnerTask) TaskResult {
	startTime := time.Now()
	
	// Create progress reporter
	progress := &progressReporter{
		terminal: r.terminal,
		taskID:   task.ID,
		taskName: task.Name,
	}

	// Show task starting
	r.terminal.Info("Starting: %s", task.Name)
	if task.Description != "" && r.options.ShowTaskDetails {
		r.terminal.Print("  %s", task.Description)
	}

	var err error
	attempts := 1
	if task.Retryable {
		attempts = r.options.RetryAttempts
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		// Show retry info
		if attempt > 1 {
			r.terminal.Warning("Retry %d/%d for: %s", attempt, attempts, task.Name)
			time.Sleep(r.options.RetryDelay)
		}

		// Execute task
		taskCtx, cancel := context.WithCancel(ctx)
		err = task.Execute(taskCtx, progress)
		cancel()

		if err == nil {
			// Success
			duration := time.Since(startTime)
			r.terminal.Success("Completed: %s (%s)", task.Name, formatDuration(duration))
			
			return TaskResult{
				TaskIndex: index,
				Success:   true,
				Duration:  duration,
				Retries:   attempt - 1,
			}
		}

		// Failed
		if attempt < attempts && task.Retryable {
			r.terminal.Warning("Failed: %s - %v (will retry)", task.Name, err)
		}
	}

	// Final failure
	duration := time.Since(startTime)
	r.terminal.Error("Failed: %s - %v", task.Name, err)
	
	return TaskResult{
		TaskIndex: index,
		Success:   false,
		Error:     err,
		Duration:  duration,
		Retries:   attempts - 1,
	}
}

// showSummary shows execution summary
func (r *Runner) showSummary(results []TaskResult, totalDuration time.Duration) {
	r.terminal.Header("Execution Summary")

	// Count successes and failures
	succeeded := 0
	failed := 0
	totalRetries := 0

	for _, result := range results {
		if result.Success {
			succeeded++
		} else {
			failed++
		}
		totalRetries += result.Retries
	}

	// Summary stats
	r.terminal.Print("Total Tasks: %d", len(results))
	r.terminal.Success("Succeeded: %d", succeeded)
	if failed > 0 {
		r.terminal.Error("Failed: %d", failed)
	}
	if totalRetries > 0 {
		r.terminal.Warning("Total Retries: %d", totalRetries)
	}
	r.terminal.Info("Total Duration: %s", formatDuration(totalDuration))

	// Detailed results table
	if r.options.ShowTaskDetails && len(results) > 0 {
		r.terminal.Print("\nDetailed Results:")
		
		headers := []string{"Task", "Status", "Duration", "Retries"}
		rows := make([][]string, 0, len(results))

		for i, result := range results {
			task := r.tasks[result.TaskIndex]
			status := "✓ Success"
			if !result.Success {
				status = "✗ Failed"
			}
			
			row := []string{
				task.Name,
				status,
				formatDuration(result.Duration),
				fmt.Sprintf("%d", result.Retries),
			}
			rows = append(rows, row)
		}

		r.terminal.Table(headers, rows)
	}

	// Failed task details
	if failed > 0 && r.options.ShowTaskDetails {
		r.terminal.Print("\nFailed Tasks:")
		for i, result := range results {
			if !result.Success && result.Error != nil {
				task := r.tasks[result.TaskIndex]
				r.terminal.Print("  • %s: %v", task.Name, result.Error)
			}
		}
	}
}

// progressReporter implements ProgressReporter
type progressReporter struct {
	terminal  *Terminal
	taskID    string
	taskName  string
	total     int64
	current   int64
	mu        sync.Mutex
}

func (pr *progressReporter) SetTotal(total int64) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	
	pr.total = total
	if pr.terminal.progressMgr != nil {
		pr.terminal.StartProgress(pr.taskID, pr.taskName, total)
	}
}

func (pr *progressReporter) SetCurrent(current int64) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	
	pr.current = current
	if pr.terminal.progressMgr != nil {
		pr.terminal.UpdateProgress(pr.taskID, current)
	}
}

func (pr *progressReporter) Increment() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	
	pr.current++
	if pr.terminal.progressMgr != nil {
		pr.terminal.UpdateProgress(pr.taskID, pr.current)
	}
}

func (pr *progressReporter) SetStatus(status string) {
	pr.terminal.Info("  %s: %s", pr.taskName, status)
}

func (pr *progressReporter) SetDetails(details string) {
	if pr.terminal.multiProg != nil {
		pr.terminal.multiProg.UpdateTask(pr.taskID, TaskRunning, float64(pr.current)/float64(pr.total), details)
	}
}

func (pr *progressReporter) AddSubTask(name string) {
	if pr.terminal.multiProg != nil {
		pr.terminal.multiProg.AddSubTask(pr.taskID, name)
	}
}

func (pr *progressReporter) CompleteSubTask(name string) {
	// Implementation would update subtask status
	log.Debug().Str("task", pr.taskID).Str("subtask", name).Msg("Subtask completed")
}

// InteractiveRunner provides interactive task execution
type InteractiveRunner struct {
	*Runner
	selectedTasks []int
}

// NewInteractiveRunner creates a new interactive runner
func NewInteractiveRunner(terminal *Terminal, options RunnerOptions) *InteractiveRunner {
	return &InteractiveRunner{
		Runner: NewRunner(terminal, options),
	}
}

// SelectTasks allows user to interactively select tasks
func (ir *InteractiveRunner) SelectTasks() error {
	if len(ir.tasks) == 0 {
		return fmt.Errorf("no tasks available")
	}

	ir.terminal.Header("Task Selection")
	
	// Build task list
	taskNames := make([]string, len(ir.tasks))
	for i, task := range ir.tasks {
		status := ""
		if task.Required {
			status = " (required)"
		}
		taskNames[i] = fmt.Sprintf("%s%s", task.Name, status)
		if task.Description != "" {
			taskNames[i] += fmt.Sprintf("\n      %s", task.Description)
		}
	}

	// Prompt for selection
	selected, err := ir.terminal.MultiSelect("Select tasks to run", taskNames)
	if err != nil {
		return err
	}

	if len(selected) == 0 {
		return fmt.Errorf("no tasks selected")
	}

	// Store selection
	ir.selectedTasks = selected

	// Confirm selection
	ir.terminal.Print("\nSelected tasks:")
	for _, idx := range selected {
		ir.terminal.Print("  • %s", ir.tasks[idx].Name)
	}

	confirmed, err := ir.terminal.Confirm("\nProceed with execution?", true)
	if err != nil {
		return err
	}

	if !confirmed {
		return fmt.Errorf("execution cancelled by user")
	}

	return nil
}

// Run executes selected tasks
func (ir *InteractiveRunner) Run(ctx context.Context) error {
	if len(ir.selectedTasks) == 0 {
		// If no tasks selected, run all
		return ir.Runner.Run(ctx)
	}

	// Filter tasks to run only selected ones
	originalTasks := ir.tasks
	selectedTasksOnly := make([]RunnerTask, len(ir.selectedTasks))
	
	for i, idx := range ir.selectedTasks {
		selectedTasksOnly[i] = originalTasks[idx]
	}
	
	ir.tasks = selectedTasksOnly
	err := ir.Runner.Run(ctx)
	ir.tasks = originalTasks // Restore original tasks
	
	return err
}

// ExampleUsage shows how to use the runner
func ExampleUsage() {
	// Create terminal
	terminal := NewTerminal(TerminalOptions{
		Output:       os.Stdout,
		ColorOutput:  true,
		ShowProgress: true,
	})

	// Create runner
	runner := NewInteractiveRunner(terminal, DefaultRunnerOptions())

	// Add tasks
	runner.AddTask(RunnerTask{
		ID:          "init",
		Name:        "Initialize System",
		Description: "Set up required directories and configurations",
		Required:    true,
		Execute: func(ctx context.Context, progress ProgressReporter) error {
			progress.SetStatus("Creating directories...")
			time.Sleep(time.Second)
			progress.SetStatus("Loading configuration...")
			time.Sleep(time.Second)
			return nil
		},
	})

	runner.AddTask(RunnerTask{
		ID:          "download",
		Name:        "Download Templates",
		Description: "Fetch latest vulnerability test templates",
		Retryable:   true,
		Execute: func(ctx context.Context, progress ProgressReporter) error {
			progress.SetTotal(100)
			for i := 0; i <= 100; i += 10 {
				progress.SetCurrent(int64(i))
				progress.SetDetails(fmt.Sprintf("Downloaded %d/100 templates", i))
				time.Sleep(200 * time.Millisecond)
			}
			return nil
		},
	})

	runner.AddTask(RunnerTask{
		ID:          "validate",
		Name:        "Validate Templates",
		Description: "Check template syntax and compatibility",
		Execute: func(ctx context.Context, progress ProgressReporter) error {
			templates := []string{"prompt-injection", "data-leakage", "model-manipulation"}
			progress.SetTotal(int64(len(templates)))
			
			for i, template := range templates {
				progress.AddSubTask(fmt.Sprintf("Validating %s", template))
				time.Sleep(500 * time.Millisecond)
				progress.SetCurrent(int64(i + 1))
			}
			return nil
		},
	})

	// Select and run tasks
	if err := runner.SelectTasks(); err != nil {
		terminal.Error("Task selection failed: %v", err)
		return
	}

	ctx := context.Background()
	if err := runner.Run(ctx); err != nil {
		terminal.Error("Execution failed: %v", err)
	}
}