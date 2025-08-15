package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// Terminal provides terminal UI functionality
type Terminal struct {
	output      io.Writer
	input       io.Reader
	width       int
	height      int
	isTerminal  bool
	colorOutput bool
	mu          sync.Mutex
	
	// Progress tracking
	progressMgr *ProgressManager
	multiProg   *MultiProgress
	spinner     *Spinner
	
	// Current state
	lastLines   int
	clearScreen bool
}

// TerminalOptions configures terminal behavior
type TerminalOptions struct {
	Output       io.Writer
	Input        io.Reader
	ColorOutput  bool
	ClearScreen  bool
	ShowProgress bool
}

// NewTerminal creates a new terminal UI
func NewTerminal(opts TerminalOptions) *Terminal {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}

	// Check if output is a terminal
	isTerminal := false
	width, height := 80, 24 // Default dimensions
	
	if f, ok := opts.Output.(*os.File); ok {
		if term.IsTerminal(int(f.Fd())) {
			isTerminal = true
			if w, h, err := term.GetSize(int(f.Fd())); err == nil {
				width, height = w, h
			}
		}
	}

	t := &Terminal{
		output:      opts.Output,
		input:       opts.Input,
		width:       width,
		height:      height,
		isTerminal:  isTerminal,
		colorOutput: opts.ColorOutput && isTerminal,
		clearScreen: opts.ClearScreen,
	}

	if opts.ShowProgress {
		progOpts := DefaultProgressOptions()
		progOpts.Width = width - 20 // Leave room for text
		t.progressMgr = NewProgressManager(opts.Output, progOpts)
		t.multiProg = NewMultiProgress(opts.Output, 10, true)
		t.spinner = NewSpinner(progOpts.SpinnerStyle)
	}

	return t

// Clear clears the terminal screen
func (t *Terminal) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isTerminal && t.clearScreen {
		// ANSI escape sequence to clear screen and move cursor to top
		fmt.Fprint(t.output, "\033[2J\033[H")
	}

// ClearLine clears the current line
func (t *Terminal) ClearLine() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isTerminal {
		// Move cursor to beginning of line and clear to end
		fmt.Fprint(t.output, "\r\033[K")
	}

// MoveCursorUp moves cursor up n lines
func (t *Terminal) MoveCursorUp(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isTerminal && n > 0 {
		fmt.Fprintf(t.output, "\033[%dA", n)
	}

// ClearPreviousLines clears n previous lines
func (t *Terminal) ClearPreviousLines(n int) {
	if t.isTerminal && n > 0 {
		// Move up n lines
		t.MoveCursorUp(n)
		
		// Clear each line
		for i := 0; i < n; i++ {
			t.ClearLine()
			if i < n-1 {
				fmt.Fprint(t.output, "\n")
			}
		}
		
		// Move back to start
		t.MoveCursorUp(n - 1)
	}

// Print methods with color support

// Success prints success message
func (t *Terminal) Success(format string, args ...interface{}) {
	t.printWithColor(color.FgGreen, "✓", format, args...)

// Error prints error message
func (t *Terminal) Error(format string, args ...interface{}) {
	t.printWithColor(color.FgRed, "✗", format, args...)

// Warning prints warning message
func (t *Terminal) Warning(format string, args ...interface{}) {
	t.printWithColor(color.FgYellow, "⚠", format, args...)

// Info prints info message
func (t *Terminal) Info(format string, args ...interface{}) {
	t.printWithColor(color.FgCyan, "ℹ", format, args...)

// Debug prints debug message
func (t *Terminal) Debug(format string, args ...interface{}) {
	t.printWithColor(color.FgMagenta, "🔍", format, args...)

// Print prints plain message
func (t *Terminal) Print(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	message := fmt.Sprintf(format, args...)
	fmt.Fprintln(t.output, message)
	t.lastLines = strings.Count(message, "\n") + 1

// printWithColor prints colored message with icon
func (t *Terminal) printWithColor(c color.Attribute, icon, format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	
	if t.colorOutput {
		colorFunc := color.New(c).SprintFunc()
		iconColored := colorFunc(icon)
		fmt.Fprintf(t.output, "%s %s\n", iconColored, message)
	} else {
		fmt.Fprintf(t.output, "%s %s\n", icon, message)
	}
	
	t.lastLines = strings.Count(message, "\n") + 1

// Header prints a section header
func (t *Terminal) Header(title string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	width := t.width
	if width > 80 {
		width = 80
	}

	// Create border
	border := strings.Repeat("─", width-4)
	
	if t.colorOutput {
		headerColor := color.New(color.FgCyan, color.Bold).SprintFunc()
		fmt.Fprintf(t.output, "\n%s\n", headerColor(fmt.Sprintf("┌─%s─┐", border)))
		fmt.Fprintf(t.output, "%s\n", headerColor(fmt.Sprintf("│ %-*s │", width-4, title)))
		fmt.Fprintf(t.output, "%s\n\n", headerColor(fmt.Sprintf("└─%s─┘", border)))
	} else {
		fmt.Fprintf(t.output, "\n┌─%s─┐\n", border)
		fmt.Fprintf(t.output, "│ %-*s │\n", width-4, title)
		fmt.Fprintf(t.output, "└─%s─┘\n\n", border)
	}
	
	t.lastLines = 5

// Table prints a formatted table
func (t *Terminal) Table(headers []string, rows [][]string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print header
	headerLine := ""
	separatorLine := ""
	
	for i, header := range headers {
		if i > 0 {
			headerLine += " │ "
			separatorLine += "─┼─"
		}
		headerLine += fmt.Sprintf("%-*s", colWidths[i], header)
		separatorLine += strings.Repeat("─", colWidths[i])
	}

	if t.colorOutput {
		headerColor := color.New(color.FgCyan, color.Bold).SprintFunc()
		fmt.Fprintln(t.output, headerColor(headerLine))
		fmt.Fprintln(t.output, headerColor(separatorLine))
	} else {
		fmt.Fprintln(t.output, headerLine)
		fmt.Fprintln(t.output, separatorLine)
	}

	// Print rows
	for _, row := range rows {
		rowLine := ""
		for i, cell := range row {
			if i > 0 {
				rowLine += " │ "
			}
			if i < len(colWidths) {
				rowLine += fmt.Sprintf("%-*s", colWidths[i], cell)
			} else {
				rowLine += cell
			}
		}
		fmt.Fprintln(t.output, rowLine)
	}
	
	t.lastLines = len(rows) + 2

// List prints a formatted list
func (t *Terminal) List(items []string, numbered bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, item := range items {
		if numbered {
			fmt.Fprintf(t.output, "  %d. %s\n", i+1, item)
		} else {
			fmt.Fprintf(t.output, "  • %s\n", item)
		}
	}
	
	t.lastLines = len(items)

// Progress bar methods

// StartProgress starts a progress bar
func (t *Terminal) StartProgress(id, description string, total int64) {
	if t.progressMgr != nil {
		t.progressMgr.CreateProgressBar(id, description, total)
	}

// UpdateProgress updates a progress bar
func (t *Terminal) UpdateProgress(id string, current int64) {
	if t.progressMgr != nil {
		t.progressMgr.Update(id, current)
	}

// FinishProgress finishes a progress bar
func (t *Terminal) FinishProgress(id string) {
	if t.progressMgr != nil {
		t.progressMgr.Finish(id)
	}

// StartSpinner starts an indeterminate spinner
func (t *Terminal) StartSpinner(message string) func() {
	if !t.isTerminal {
		fmt.Fprintf(t.output, "%s...\n", message)
		return func() {}
	}

	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				t.ClearLine()
				return
			case <-ticker.C:
				t.ClearLine()
				frame := t.spinner.Next()
				fmt.Fprintf(t.output, "%s %s", frame, message)
			}
		}
	}()

	return func() {
		done <- true
		close(done)
		time.Sleep(100 * time.Millisecond) // Allow spinner to clear
	}

// Interactive methods

// Prompt prompts for user input
func (t *Terminal) Prompt(prompt string) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Fprint(t.output, prompt)
	
	var response string
	_, err := fmt.Fscanln(t.input, &response)
	return response, err

// Confirm prompts for yes/no confirmation
func (t *Terminal) Confirm(prompt string, defaultYes bool) (bool, error) {
	defaultStr := "y/N"
	if defaultYes {
		defaultStr = "Y/n"
	}
	
	response, err := t.Prompt(fmt.Sprintf("%s [%s]: ", prompt, defaultStr))
	if err != nil {
			// User just pressed enter
			return defaultYes, nil
		}
		return false, err
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil

// Select prompts user to select from options
func (t *Terminal) Select(prompt string, options []string) (int, error) {
	t.Print("%s", prompt)
	t.List(options, true)
	
	for {
		response, err := t.Prompt("Select option: ")
		if err != nil {
			return -1, err
		}
		
		var index int
		if _, err := fmt.Sscanf(response, "%d", &index); err == nil {
			if index >= 1 && index <= len(options) {
				return index - 1, nil
			}
		}
		
		t.Warning("Invalid selection. Please enter a number between 1 and %d.", len(options))
	}

// MultiSelect prompts user to select multiple options
func (t *Terminal) MultiSelect(prompt string, options []string) ([]int, error) {
	t.Print("%s (comma-separated numbers or 'all'):", prompt)
	t.List(options, true)
	
	response, err := t.Prompt("Select options: ")
	if err != nil {
		return nil, err
	}
	
	response = strings.TrimSpace(response)
	if strings.ToLower(response) == "all" {
		indices := make([]int, len(options))
		for i := range indices {
			indices[i] = i
		}
		return indices, nil
	}
	
	var indices []int
	parts := strings.Split(response, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		var index int
		if _, err := fmt.Sscanf(part, "%d", &index); err == nil {
			if index >= 1 && index <= len(options) {
				indices = append(indices, index-1)
			}
		}
	}
	
	return indices, nil

// ProgressDemo demonstrates progress indicators
func (t *Terminal) ProgressDemo() {
	t.Header("Progress Indicator Demo")
	
	// Simple progress bar
	t.Info("Downloading templates...")
	stop := t.StartSpinner("Connecting to repository")
	time.Sleep(2 * time.Second)
	stop()
	
	t.StartProgress("download", "Downloading templates", 100)
	for i := 0; i <= 100; i += 5 {
		t.UpdateProgress("download", int64(i))
		time.Sleep(50 * time.Millisecond)
	}
	t.FinishProgress("download")
	t.Success("Templates downloaded successfully")
	
	// Multi-task progress
	t.Info("\nRunning security scans...")
	
	if t.multiProg != nil {
		// Add tasks
		task1 := t.multiProg.AddTask("scan1", "Prompt Injection Tests")
		task2 := t.multiProg.AddTask("scan2", "Data Leakage Tests")
		task3 := t.multiProg.AddTask("scan3", "Model Manipulation Tests")
		
		// Update task states
		t.multiProg.UpdateTask(task1.ID, TaskRunning, 0.0, "Initializing...")
		time.Sleep(500 * time.Millisecond)
		
		t.multiProg.UpdateTask(task1.ID, TaskRunning, 0.5, "Running test suite...")
		t.multiProg.UpdateTask(task2.ID, TaskRunning, 0.0, "Preparing payloads...")
		time.Sleep(1 * time.Second)
		
		t.multiProg.UpdateTask(task1.ID, TaskCompleted, 1.0, "15 tests passed")
		t.multiProg.UpdateTask(task2.ID, TaskRunning, 0.7, "Analyzing responses...")
		t.multiProg.UpdateTask(task3.ID, TaskRunning, 0.2, "Testing boundaries...")
		time.Sleep(1 * time.Second)
		
		t.multiProg.UpdateTask(task2.ID, TaskCompleted, 1.0, "8 tests passed")
		t.multiProg.UpdateTask(task3.ID, TaskFailed, 0.8, "Connection timeout")
		
		// Render final state
		t.Print("\n%s", t.multiProg.Render())
	}

// Dimensions returns terminal dimensions
func (t *Terminal) Dimensions() (width, height int) {
	return t.width, t.height

// IsTerminal returns true if output is a terminal
func (t *Terminal) IsTerminal() bool {
	return t.isTerminal
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
