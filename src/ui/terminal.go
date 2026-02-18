package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

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
}

// Clear clears the terminal screen
func (t *Terminal) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isTerminal && t.clearScreen {
		// ANSI escape sequence to clear screen and move cursor to top
		_, _ = fmt.Fprint(t.output, "\033[2J\033[H")
	}
}

// ClearLine clears the current line
func (t *Terminal) ClearLine() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isTerminal {
		// Move cursor to beginning of line and clear to end
		_, _ = fmt.Fprint(t.output, "\r\033[K")
	}
}

// MoveCursorUp moves cursor up n lines
func (t *Terminal) MoveCursorUp(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isTerminal && n > 0 {
		_, _ = fmt.Fprintf(t.output, "\033[%dA", n)
	}
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
				_, _ = fmt.Fprint(t.output, "\n")
			}
		}

		// Move back to start
		t.MoveCursorUp(n - 1)
	}
}

// Print methods with color support

// Success prints success message
func (t *Terminal) Success(format string, args ...interface{}) {
	t.printWithColor(color.FgGreen, "✓", format, args...)
}

// Error prints error message
func (t *Terminal) Error(format string, args ...interface{}) {
	t.printWithColor(color.FgRed, "✗", format, args...)
}

// Warning prints warning message
func (t *Terminal) Warning(format string, args ...interface{}) {
	t.printWithColor(color.FgYellow, "⚠", format, args...)
}

// Info prints info message
func (t *Terminal) Info(format string, args ...interface{}) {
	t.printWithColor(color.FgCyan, "ℹ", format, args...)
}

// Debug prints debug message
func (t *Terminal) Debug(format string, args ...interface{}) {
	t.printWithColor(color.FgMagenta, "🔍", format, args...)
}

// Print prints plain message
func (t *Terminal) Print(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintln(t.output, message)
	t.lastLines = strings.Count(message, "\n") + 1
}

// printWithColor prints colored message with icon
func (t *Terminal) printWithColor(c color.Attribute, icon, format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	message := fmt.Sprintf(format, args...)

	if t.colorOutput {
		colorFunc := color.New(c).SprintFunc()
		iconColored := colorFunc(icon)
		_, _ = fmt.Fprintf(t.output, "%s %s\n", iconColored, message)
	} else {
		_, _ = fmt.Fprintf(t.output, "%s %s\n", icon, message)
	}

	t.lastLines = strings.Count(message, "\n") + 1
}

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
		_, _ = fmt.Fprintf(t.output, "\n%s\n", headerColor(fmt.Sprintf("┌─%s─┐", border)))
		_, _ = fmt.Fprintf(t.output, "%s\n", headerColor(fmt.Sprintf("│ %-*s │", width-4, title)))
		_, _ = fmt.Fprintf(t.output, "%s\n\n", headerColor(fmt.Sprintf("└─%s─┘", border)))
	} else {
		_, _ = fmt.Fprintf(t.output, "\n┌─%s─┐\n", border)
		_, _ = fmt.Fprintf(t.output, "│ %-*s │\n", width-4, title)
		_, _ = fmt.Fprintf(t.output, "└─%s─┘\n\n", border)
	}

	t.lastLines = 5
}

// Table prints a formatted table with variadic arguments
// Can be called as:
// - Table(headers, rows) - separate headers and rows
// - Table(table) - table where first row contains headers
func (t *Terminal) Table(args ...interface{}) {
	if len(args) == 1 {
		// Single argument - table with headers as first row
		if table, ok := args[0].([][]string); ok && len(table) > 0 {
			headers := table[0]
			rows := table[1:]
			t.tableImpl(headers, rows)
		}
	} else if len(args) == 2 {
		// Two arguments - separate headers and rows
		if headers, ok := args[0].([]string); ok {
			if rows, ok := args[1].([][]string); ok {
				t.tableImpl(headers, rows)
			}
		}
	}
}

// tableImpl is the actual table implementation
func (t *Terminal) tableImpl(headers []string, rows [][]string) {
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
		_, _ = fmt.Fprintln(t.output, headerColor(headerLine))
		_, _ = fmt.Fprintln(t.output, headerColor(separatorLine))
	} else {
		_, _ = fmt.Fprintln(t.output, headerLine)
		_, _ = fmt.Fprintln(t.output, separatorLine)
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
		_, _ = fmt.Fprintln(t.output, rowLine)
	}

	t.lastLines = len(rows) + 2
}

// List prints a formatted list
func (t *Terminal) List(items []string, numbered bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, item := range items {
		if numbered {
			_, _ = fmt.Fprintf(t.output, "  %d. %s\n", i+1, item)
		} else {
			_, _ = fmt.Fprintf(t.output, "  • %s\n", item)
		}
	}

	t.lastLines = len(items)
}

// Progress bar methods

// StartProgress starts a progress bar
func (t *Terminal) StartProgress(id, description string, total int64) {
	if t.progressMgr != nil {
		t.progressMgr.CreateProgressBar(id, description, total)
	}
}

// UpdateProgress updates a progress bar
func (t *Terminal) UpdateProgress(id string, current int64) {
	if t.progressMgr != nil {
		t.progressMgr.Update(id, current)
	}
}

// FinishProgress finishes a progress bar
func (t *Terminal) FinishProgress(id string) {
	if t.progressMgr != nil {
		t.progressMgr.Finish(id)
	}
}

// StartSpinner starts an indeterminate spinner
func (t *Terminal) StartSpinner(message string) func() {
	if !t.isTerminal {
		_, _ = fmt.Fprintf(t.output, "%s...\n", message)
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
				_, _ = fmt.Fprintf(t.output, "%s %s", frame, message)
			}
		}
	}()

	return func() {
		done <- true
		close(done)
		time.Sleep(100 * time.Millisecond) // Allow spinner to clear
	}
}

// Interactive methods

// Prompt prompts for user input
func (t *Terminal) Prompt(prompt string) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, _ = fmt.Fprint(t.output, prompt)

	var response string
	_, err := fmt.Fscanln(t.input, &response)
	return response, err
}

// Confirm prompts for yes/no confirmation
func (t *Terminal) Confirm(prompt string, defaultYes bool) (bool, error) {
	defaultStr := "y/N"
	if defaultYes {
		defaultStr = "Y/n"
	}

	response, err := t.Prompt(fmt.Sprintf("%s [%s]: ", prompt, defaultStr))
	if err != nil {
		if response == "" {
			// User just pressed enter
			return defaultYes, nil
		}
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

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
}

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
}

// Dimensions returns terminal dimensions
func (t *Terminal) Dimensions() (width, height int) {
	return t.width, t.height
}

// IsTerminal returns true if output is a terminal
func (t *Terminal) IsTerminal() bool {
	return t.isTerminal
}

// Section prints a section header
func (t *Terminal) Section(title string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.colorOutput {
		headerColor := color.New(color.FgCyan, color.Bold).SprintFunc()
		_, _ = fmt.Fprintf(t.output, "\n%s %s\n%s\n",
			headerColor("▶"),
			headerColor(title),
			strings.Repeat("─", len(title)+2),
		)
	} else {
		_, _ = fmt.Fprintf(t.output, "\n▶ %s\n%s\n",
			title,
			strings.Repeat("─", len(title)+2),
		)
	}

	t.lastLines = 3
}

// KeyValue prints a key-value pair
func (t *Terminal) KeyValue(key string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.colorOutput {
		keyColor := color.New(color.FgCyan).SprintFunc()
		valueColor := color.New(color.FgWhite).SprintFunc()
		_, _ = fmt.Fprintf(t.output, "%s: %s\n", keyColor(key), valueColor(fmt.Sprintf("%v", value)))
	} else {
		_, _ = fmt.Fprintf(t.output, "%s: %v\n", key, value)
	}

	t.lastLines = 1
}

// Subheader prints a subheader
func (t *Terminal) Subheader(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.colorOutput {
		subheaderColor := color.New(color.FgCyan, color.Bold).SprintFunc()
		_, _ = fmt.Fprintln(t.output, subheaderColor(text))
	} else {
		_, _ = fmt.Fprintln(t.output, text)
	}

	t.lastLines = 1
}

// HasInput checks if there is input available
func (t *Terminal) HasInput() bool {
	if f, ok := t.input.(*os.File); ok {
		// Use term package to check for input availability
		// This is a simplified implementation
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// ReadKey reads a single key from input
func (t *Terminal) ReadKey() string {
	if f, ok := t.input.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		// Set terminal to raw mode for single character reading
		oldState, err := term.MakeRaw(int(f.Fd()))
		if err != nil {
			return ""
		}
		defer term.Restore(int(f.Fd()), oldState)

		// Read a single byte
		var b [1]byte
		_, err = f.Read(b[:])
		if err != nil {
			return ""
		}
		return string(rune(b[0]))
	}

	// Fallback for non-terminal input
	var input string
	_, err := fmt.Fscanln(t.input, &input)
	if err != nil {
		return ""
	}
	if len(input) > 0 {
		return string(rune(input[0]))
	}
	return ""
}

// Box prints content in a bordered box
func (t *Terminal) Box(title, content string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	lines := strings.Split(content, "\n")
	maxWidth := len(title)

	// Find the maximum line width
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	// Ensure minimum width and add padding
	maxWidth += 4
	if maxWidth < 20 {
		maxWidth = 20
	}

	border := strings.Repeat("─", maxWidth-2)

	if t.colorOutput {
		boxColor := color.New(color.FgCyan).SprintFunc()
		titleColor := color.New(color.FgCyan, color.Bold).SprintFunc()

		// Top border with title
		_, _ = fmt.Fprintf(t.output, "%s\n", boxColor("┌─"+border+"─┐"))
		_, _ = fmt.Fprintf(t.output, "%s %s %s\n", boxColor("│"), titleColor(fmt.Sprintf("%-*s", maxWidth-4, title)), boxColor("│"))
		_, _ = fmt.Fprintf(t.output, "%s\n", boxColor("├─"+border+"─┤"))

		// Content lines
		for _, line := range lines {
			_, _ = fmt.Fprintf(t.output, "%s %-*s %s\n", boxColor("│"), maxWidth-4, line, boxColor("│"))
		}

		// Bottom border
		_, _ = fmt.Fprintf(t.output, "%s\n", boxColor("└─"+border+"─┘"))
	} else {
		// Top border with title
		_, _ = fmt.Fprintf(t.output, "┌─%s─┐\n", border)
		_, _ = fmt.Fprintf(t.output, "│ %-*s │\n", maxWidth-4, title)
		_, _ = fmt.Fprintf(t.output, "├─%s─┤\n", border)

		// Content lines
		for _, line := range lines {
			_, _ = fmt.Fprintf(t.output, "│ %-*s │\n", maxWidth-4, line)
		}

		// Bottom border
		_, _ = fmt.Fprintf(t.output, "└─%s─┘\n", border)
	}

	t.lastLines = len(lines) + 4
}

// HeaderBox prints a header in a box format
func (t *Terminal) HeaderBox(title string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	width := len(title) + 4
	if width < 20 {
		width = 20
	}

	border := strings.Repeat("─", width-2)

	if t.colorOutput {
		headerColor := color.New(color.FgCyan, color.Bold).SprintFunc()
		_, _ = fmt.Fprintf(t.output, "\n%s\n", headerColor("┌─"+border+"─┐"))
		_, _ = fmt.Fprintf(t.output, "%s\n", headerColor(fmt.Sprintf("│ %-*s │", width-4, title)))
		_, _ = fmt.Fprintf(t.output, "%s\n\n", headerColor("└─"+border+"─┘"))
	} else {
		_, _ = fmt.Fprintf(t.output, "\n┌─%s─┐\n", border)
		_, _ = fmt.Fprintf(t.output, "│ %-*s │\n", width-4, title)
		_, _ = fmt.Fprintf(t.output, "└─%s─┘\n\n", border)
	}

	t.lastLines = 5
}

// Printf prints formatted text
func (t *Terminal) Printf(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprint(t.output, message)
	t.lastLines = strings.Count(message, "\n")
	if !strings.HasSuffix(message, "\n") {
		t.lastLines++
	}
}

// Input prompts for user input with an optional default value
func (t *Terminal) Input(prompt string, defaultValue string) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Display prompt with default value if provided
	if defaultValue != "" {
		_, _ = fmt.Fprintf(t.output, "%s [%s]: ", prompt, defaultValue)
	} else {
		_, _ = fmt.Fprintf(t.output, "%s: ", prompt)
	}

	var response string
	_, err := fmt.Fscanln(t.input, &response)
	if err != nil {
		// If there was an error reading (e.g., just pressing Enter), return default
		if defaultValue != "" {
			return defaultValue, nil
		}
		return "", err
	}

	// If user entered nothing and we have a default, use it
	if strings.TrimSpace(response) == "" && defaultValue != "" {
		return defaultValue, nil
	}

	return response, nil
}

// Subsection prints a subsection header (smaller than Section)
func (t *Terminal) Subsection(title string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.colorOutput {
		subheaderColor := color.New(color.FgCyan).SprintFunc()
		_, _ = fmt.Fprintf(t.output, "\n%s %s\n%s\n",
			subheaderColor("▸"),
			subheaderColor(title),
			strings.Repeat("─", len(title)+2),
		)
	} else {
		_, _ = fmt.Fprintf(t.output, "\n▸ %s\n%s\n",
			title,
			strings.Repeat("─", len(title)+2),
		)
	}

	t.lastLines = 3
}

// Muted prints text in a muted/gray color
func (t *Terminal) Muted(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	message := fmt.Sprintf(format, args...)

	if t.colorOutput {
		mutedColor := color.New(color.FgHiBlack).SprintFunc()
		_, _ = fmt.Fprintln(t.output, mutedColor(message))
	} else {
		_, _ = fmt.Fprintln(t.output, message)
	}

	t.lastLines = strings.Count(message, "\n") + 1
}

// Code prints formatted code with syntax highlighting
func (t *Terminal) Code(code string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.colorOutput {
		codeColor := color.New(color.FgGreen).SprintFunc()
		_, _ = fmt.Fprintf(t.output, "  %s\n", codeColor(code))
	} else {
		_, _ = fmt.Fprintf(t.output, "  %s\n", code)
	}

	t.lastLines = strings.Count(code, "\n") + 1
}

// Bold prints text in bold
func (t *Terminal) Bold(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	message := fmt.Sprintf(format, args...)

	if t.colorOutput {
		boldColor := color.New(color.Bold).SprintFunc()
		_, _ = fmt.Fprintln(t.output, boldColor(message))
	} else {
		_, _ = fmt.Fprintln(t.output, message)
	}

	t.lastLines = strings.Count(message, "\n") + 1
}
