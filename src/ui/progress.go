package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressManager manages multiple progress indicators
type ProgressManager struct {
	bars       map[string]*ProgressBar
	mu         sync.RWMutex
	output     io.Writer
	showETA    bool
	showRate   bool
	showBytes  bool
	width      int
}

// ProgressBar represents an individual progress bar
type ProgressBar struct {
	bar         *progressbar.ProgressBar
	description string
	total       int64
	current     int64
	startTime   time.Time
	lastUpdate  time.Time
}

// ProgressOptions configures progress display
type ProgressOptions struct {
	ShowETA      bool
	ShowRate     bool
	ShowBytes    bool
	Width        int
	RefreshRate  time.Duration
	SpinnerStyle []string
}

// DefaultProgressOptions returns default progress options
func DefaultProgressOptions() ProgressOptions {
	return ProgressOptions{
		ShowETA:     true,
		ShowRate:    true,
		ShowBytes:   false,
		Width:       40,
		RefreshRate: 100 * time.Millisecond,
		SpinnerStyle: []string{
			"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
		},
	}
}

// NewProgressManager creates a new progress manager
func NewProgressManager(output io.Writer, opts ProgressOptions) *ProgressManager {
	return &ProgressManager{
		bars:      make(map[string]*ProgressBar),
		output:    output,
		showETA:   opts.ShowETA,
		showRate:  opts.ShowRate,
		showBytes: opts.ShowBytes,
		width:     opts.Width,
	}
}

// CreateProgressBar creates a new progress bar
func (pm *ProgressManager) CreateProgressBar(id, description string, total int64) *ProgressBar {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	options := []progressbar.Option{
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(pm.output),
		progressbar.OptionSetWidth(pm.width),
		progressbar.OptionThrottle(100 * time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(pm.output, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	}

	if pm.showETA {
		options = append(options, progressbar.OptionShowElapsedTimeOnFinish())
	}

	if pm.showBytes {
		options = append(options, progressbar.OptionShowBytes(true))
	}

	bar := progressbar.NewOptions64(total, options...)
	
	pb := &ProgressBar{
		bar:         bar,
		description: description,
		total:       total,
		current:     0,
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
	}

	pm.bars[id] = pb
	return pb
}

// Update updates progress for a specific bar
func (pm *ProgressManager) Update(id string, current int64) error {
	pm.mu.RLock()
	bar, exists := pm.bars[id]
	pm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("progress bar '%s' not found", id)
	}

	delta := current - bar.current
	if delta > 0 {
		bar.bar.Add64(delta)
		bar.current = current
		bar.lastUpdate = time.Now()
	}

	return nil
}

// Increment increments progress by 1
func (pm *ProgressManager) Increment(id string) error {
	pm.mu.RLock()
	bar, exists := pm.bars[id]
	pm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("progress bar '%s' not found", id)
	}

	bar.bar.Add(1)
	bar.current++
	bar.lastUpdate = time.Now()

	return nil
}

// Finish marks a progress bar as complete
func (pm *ProgressManager) Finish(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	bar, exists := pm.bars[id]
	if !exists {
		return fmt.Errorf("progress bar '%s' not found", id)
	}

	bar.bar.Finish()
	delete(pm.bars, id)

	return nil
}

// Clear removes all progress bars
func (pm *ProgressManager) Clear() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, bar := range pm.bars {
		bar.bar.Clear()
	}
	pm.bars = make(map[string]*ProgressBar)
}

// MultiProgress manages multiple concurrent progress indicators
type MultiProgress struct {
	tasks      []*Task
	mu         sync.RWMutex
	output     io.Writer
	maxTasks   int
	showDetail bool
}

// Task represents a tracked task
type Task struct {
	ID          string
	Name        string
	Status      TaskStatus
	Progress    float64
	StartTime   time.Time
	EndTime     *time.Time
	Details     string
	SubTasks    []*SubTask
	Error       error
}

// SubTask represents a sub-task
type SubTask struct {
	Name     string
	Status   TaskStatus
	Progress float64
}

// TaskStatus represents task status
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
	TaskSkipped
)

// String returns string representation of task status
func (ts TaskStatus) String() string {
	switch ts {
	case TaskPending:
		return "Pending"
	case TaskRunning:
		return "Running"
	case TaskCompleted:
		return "Completed"
	case TaskFailed:
		return "Failed"
	case TaskSkipped:
		return "Skipped"
	default:
		return "Unknown"
	}
}

// Symbol returns symbol for task status
func (ts TaskStatus) Symbol() string {
	switch ts {
	case TaskPending:
		return "⏸"
	case TaskRunning:
		return "▶"
	case TaskCompleted:
		return "✓"
	case TaskFailed:
		return "✗"
	case TaskSkipped:
		return "⏭"
	default:
		return "?"
	}
}

// NewMultiProgress creates a new multi-progress tracker
func NewMultiProgress(output io.Writer, maxTasks int, showDetail bool) *MultiProgress {
	return &MultiProgress{
		tasks:      make([]*Task, 0),
		output:     output,
		maxTasks:   maxTasks,
		showDetail: showDetail,
	}
}

// AddTask adds a new task to track
func (mp *MultiProgress) AddTask(id, name string) *Task {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	task := &Task{
		ID:        id,
		Name:      name,
		Status:    TaskPending,
		Progress:  0,
		StartTime: time.Now(),
		SubTasks:  make([]*SubTask, 0),
	}

	mp.tasks = append(mp.tasks, task)
	
	// Keep only the most recent tasks
	if len(mp.tasks) > mp.maxTasks {
		mp.tasks = mp.tasks[len(mp.tasks)-mp.maxTasks:]
	}

	return task
}

// UpdateTask updates task status
func (mp *MultiProgress) UpdateTask(id string, status TaskStatus, progress float64, details string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for _, task := range mp.tasks {
		if task.ID == id {
			task.Status = status
			task.Progress = progress
			task.Details = details
			
			if status == TaskCompleted || status == TaskFailed {
				now := time.Now()
				task.EndTime = &now
			}
			
			return nil
		}
	}

	return fmt.Errorf("task '%s' not found", id)
}

// AddSubTask adds a sub-task to a task
func (mp *MultiProgress) AddSubTask(taskID, name string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for _, task := range mp.tasks {
		if task.ID == taskID {
			subTask := &SubTask{
				Name:     name,
				Status:   TaskPending,
				Progress: 0,
			}
			task.SubTasks = append(task.SubTasks, subTask)
			return nil
		}
	}

	return fmt.Errorf("task '%s' not found", taskID)
}

// Render renders the current progress state
func (mp *MultiProgress) Render() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var output strings.Builder

	for i, task := range mp.tasks {
		// Task header
		output.WriteString(fmt.Sprintf("%s %s", task.Status.Symbol(), task.Name))
		
		// Progress bar for running tasks
		if task.Status == TaskRunning && task.Progress > 0 {
			bar := mp.renderProgressBar(task.Progress)
			output.WriteString(fmt.Sprintf(" %s %.0f%%", bar, task.Progress*100))
		}
		
		// Duration for completed tasks
		if task.EndTime != nil {
			duration := task.EndTime.Sub(task.StartTime)
			output.WriteString(fmt.Sprintf(" (%s)", formatDuration(duration)))
		}
		
		// Error for failed tasks
		if task.Status == TaskFailed && task.Error != nil {
			output.WriteString(fmt.Sprintf(" - Error: %s", task.Error.Error()))
		}
		
		// Details if enabled
		if mp.showDetail && task.Details != "" {
			output.WriteString(fmt.Sprintf("\n  └─ %s", task.Details))
		}
		
		// Sub-tasks if any
		if len(task.SubTasks) > 0 && mp.showDetail {
			for j, subTask := range task.SubTasks {
				prefix := "├─"
				if j == len(task.SubTasks)-1 {
					prefix = "└─"
				}
				output.WriteString(fmt.Sprintf("\n  %s %s %s", prefix, subTask.Status.Symbol(), subTask.Name))
			}
		}
		
		if i < len(mp.tasks)-1 {
			output.WriteString("\n")
		}
	}

	return output.String()
}

// renderProgressBar renders a simple progress bar
func (mp *MultiProgress) renderProgressBar(progress float64) string {
	width := 20
	filled := int(progress * float64(width))
	
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"
	
	return bar
}

// formatDuration formats a duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", mins, secs)
	}
	
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", hours, mins)
}

// Spinner provides animated spinner for indeterminate progress
type Spinner struct {
	frames  []string
	current int
	mu      sync.Mutex
}

// NewSpinner creates a new spinner
func NewSpinner(style []string) *Spinner {
	if len(style) == 0 {
		style = DefaultProgressOptions().SpinnerStyle
	}
	
	return &Spinner{
		frames:  style,
		current: 0,
	}
}

// Next returns the next frame
func (s *Spinner) Next() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	frame := s.frames[s.current]
	s.current = (s.current + 1) % len(s.frames)
	return frame
}

// Reset resets the spinner
func (s *Spinner) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = 0
}