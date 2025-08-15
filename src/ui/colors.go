package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// ColorScheme defines colors for different UI elements
type ColorScheme struct {
	// Status colors
	Success   *color.Color
	Error     *color.Color
	Warning   *color.Color
	Info      *color.Color
	Debug     *color.Color
	
	// Severity colors for vulnerabilities
	Critical  *color.Color
	High      *color.Color
	Medium    *color.Color
	Low       *color.Color
	
	// UI element colors
	Header    *color.Color
	Subheader *color.Color
	Label     *color.Color
	Value     *color.Color
	Muted     *color.Color
	
	// Special colors
	Highlight *color.Color
	Link      *color.Color
	Code      *color.Color
	Quote     *color.Color
}

// DefaultColorScheme returns the default color scheme
func DefaultColorScheme() *ColorScheme {
	return &ColorScheme{
		// Status colors
		Success: color.New(color.FgGreen, color.Bold),
		Error:   color.New(color.FgRed, color.Bold),
		Warning: color.New(color.FgYellow, color.Bold),
		Info:    color.New(color.FgCyan),
		Debug:   color.New(color.FgMagenta),
		
		// Severity colors
		Critical: color.New(color.FgRed, color.Bold, color.BgWhite),
		High:     color.New(color.FgRed),
		Medium:   color.New(color.FgYellow),
		Low:      color.New(color.FgBlue),
		
		// UI elements
		Header:    color.New(color.FgCyan, color.Bold, color.Underline),
		Subheader: color.New(color.FgCyan, color.Bold),
		Label:     color.New(color.FgWhite, color.Bold),
		Value:     color.New(color.FgGreen),
		Muted:     color.New(color.FgHiBlack),
		
		// Special
		Highlight: color.New(color.FgYellow, color.Bold),
		Link:      color.New(color.FgBlue, color.Underline),
		Code:      color.New(color.FgMagenta, color.BgHiBlack),
		Quote:     color.New(color.FgHiWhite),
	}

// DarkColorScheme returns a color scheme optimized for dark terminals
func DarkColorScheme() *ColorScheme {
	return &ColorScheme{
		Success:   color.New(color.FgHiGreen, color.Bold),
		Error:     color.New(color.FgHiRed, color.Bold),
		Warning:   color.New(color.FgHiYellow, color.Bold),
		Info:      color.New(color.FgHiCyan),
		Debug:     color.New(color.FgHiMagenta),
		Critical:  color.New(color.FgHiRed, color.Bold, color.BgHiWhite),
		High:      color.New(color.FgHiRed),
		Medium:    color.New(color.FgHiYellow),
		Low:       color.New(color.FgHiBlue),
		Header:    color.New(color.FgHiCyan, color.Bold, color.Underline),
		Subheader: color.New(color.FgHiCyan, color.Bold),
		Label:     color.New(color.FgHiWhite, color.Bold),
		Value:     color.New(color.FgHiGreen),
		Muted:     color.New(color.FgWhite),
		Highlight: color.New(color.FgHiYellow, color.Bold),
		Link:      color.New(color.FgHiBlue, color.Underline),
		Code:      color.New(color.FgHiMagenta),
		Quote:     color.New(color.FgHiWhite),
	}

// LightColorScheme returns a color scheme optimized for light terminals
func LightColorScheme() *ColorScheme {
	return &ColorScheme{
		Success:   color.New(color.FgGreen, color.Bold),
		Error:     color.New(color.FgRed, color.Bold),
		Warning:   color.New(color.FgYellow, color.Bold),
		Info:      color.New(color.FgBlue),
		Debug:     color.New(color.FgMagenta),
		Critical:  color.New(color.FgWhite, color.Bold, color.BgRed),
		High:      color.New(color.FgRed),
		Medium:    color.New(color.FgYellow),
		Low:       color.New(color.FgBlue),
		Header:    color.New(color.FgBlue, color.Bold, color.Underline),
		Subheader: color.New(color.FgBlue, color.Bold),
		Label:     color.New(color.FgBlack, color.Bold),
		Value:     color.New(color.FgGreen),
		Muted:     color.New(color.FgHiBlack),
		Highlight: color.New(color.FgYellow, color.Bold, color.BgHiWhite),
		Link:      color.New(color.FgBlue, color.Underline),
		Code:      color.New(color.FgWhite, color.BgBlack),
		Quote:     color.New(color.FgHiBlack),
	}

// Formatter provides color formatting functions
type Formatter struct {
	scheme  *ColorScheme
	enabled bool

// NewFormatter creates a new formatter
func NewFormatter(scheme *ColorScheme, enabled bool) *Formatter {
	if scheme == nil {
		scheme = DefaultColorScheme()
	}
	return &Formatter{
		scheme:  scheme,
		enabled: enabled,
	}

// Format methods for different elements

// Success formats success messages
func (f *Formatter) Success(format string, args ...interface{}) string {
	return f.format(f.scheme.Success, format, args...)

// Error formats error messages
func (f *Formatter) Error(format string, args ...interface{}) string {
	return f.format(f.scheme.Error, format, args...)

// Warning formats warning messages
func (f *Formatter) Warning(format string, args ...interface{}) string {
	return f.format(f.scheme.Warning, format, args...)

// Info formats info messages
func (f *Formatter) Info(format string, args ...interface{}) string {
	return f.format(f.scheme.Info, format, args...)

// Debug formats debug messages
func (f *Formatter) Debug(format string, args ...interface{}) string {
	return f.format(f.scheme.Debug, format, args...)

// Severity formats severity levels
func (f *Formatter) Severity(level string) string {
	switch strings.ToLower(level) {
	case "critical":
		return f.format(f.scheme.Critical, level)
	case "high":
		return f.format(f.scheme.High, level)
	case "medium":
		return f.format(f.scheme.Medium, level)
	case "low":
		return f.format(f.scheme.Low, level)
	default:
		return level
	}

// Header formats headers
func (f *Formatter) Header(text string) string {
	return f.format(f.scheme.Header, text)

// Subheader formats subheaders
func (f *Formatter) Subheader(text string) string {
	return f.format(f.scheme.Subheader, text)

// Label formats labels
func (f *Formatter) Label(text string) string {
	return f.format(f.scheme.Label, text)

// Value formats values
func (f *Formatter) Value(format string, args ...interface{}) string {
	return f.format(f.scheme.Value, format, args...)

// Muted formats muted text
func (f *Formatter) Muted(format string, args ...interface{}) string {
	return f.format(f.scheme.Muted, format, args...)

// Highlight formats highlighted text
func (f *Formatter) Highlight(text string) string {
	return f.format(f.scheme.Highlight, text)

// Link formats links
func (f *Formatter) Link(url string) string {
	return f.format(f.scheme.Link, url)

// Code formats code snippets
func (f *Formatter) Code(code string) string {
	return f.format(f.scheme.Code, code)

// Quote formats quotes
func (f *Formatter) Quote(text string) string {
	return f.format(f.scheme.Quote, text)

// format applies color if enabled
func (f *Formatter) format(c *color.Color, format string, args ...interface{}) string {
	text := fmt.Sprintf(format, args...)
	if f.enabled && c != nil {
		return c.Sprint(text)
	}
	return text

// Icons provides colored icons
type Icons struct {
	Success   string
	Error     string
	Warning   string
	Info      string
	Debug     string
	Question  string
	Arrow     string
	Bullet    string
	Check     string
	Cross     string
	Star      string
	Heart     string
	Lightning string
	Fire      string
	Shield    string
	Lock      string
	Key       string
	Globe     string
	Cloud     string
	Database  string

// DefaultIcons returns default icon set
func DefaultIcons() *Icons {
	return &Icons{
		Success:   "✓",
		Error:     "✗",
		Warning:   "⚠",
		Info:      "ℹ",
		Debug:     "🔍",
		Question:  "?",
		Arrow:     "→",
		Bullet:    "•",
		Check:     "✓",
		Cross:     "✗",
		Star:      "★",
		Heart:     "♥",
		Lightning: "⚡",
		Fire:      "🔥",
		Shield:    "🛡",
		Lock:      "🔒",
		Key:       "🔑",
		Globe:     "🌐",
		Cloud:     "☁",
		Database:  "🗄",
	}

// ASCIIIcons returns ASCII-only icons for compatibility
func ASCIIIcons() *Icons {
	return &Icons{
		Success:   "[OK]",
		Error:     "[X]",
		Warning:   "[!]",
		Info:      "[i]",
		Debug:     "[D]",
		Question:  "[?]",
		Arrow:     "->",
		Bullet:    "*",
		Check:     "[v]",
		Cross:     "[x]",
		Star:      "[*]",
		Heart:     "<3",
		Lightning: "[!]",
		Fire:      "[F]",
		Shield:    "[S]",
		Lock:      "[L]",
		Key:       "[K]",
		Globe:     "[G]",
		Cloud:     "[C]",
		Database:  "[DB]",
	}

// Box drawing characters
type BoxChars struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	Cross       string
	TeeLeft     string
	TeeRight    string
	TeeTop      string
	TeeBottom   string
}

// DefaultBoxChars returns Unicode box drawing characters
func DefaultBoxChars() *BoxChars {
	return &BoxChars{
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
		Cross:       "┼",
		TeeLeft:     "├",
		TeeRight:    "┤",
		TeeTop:      "┬",
		TeeBottom:   "┴",
	}

// ASCIIBoxChars returns ASCII box drawing characters
func ASCIIBoxChars() *BoxChars {
	return &BoxChars{
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		Horizontal:  "-",
		Vertical:    "|",
		Cross:       "+",
		TeeLeft:     "+",
		TeeRight:    "+",
		TeeTop:      "+",
		TeeBottom:   "+",
	}

// RenderBox renders a box with content
func RenderBox(title, content string, width int, boxChars *BoxChars, formatter *Formatter) string {
	if boxChars == nil {
		boxChars = DefaultBoxChars()
	}

	var result strings.Builder

	// Top border
	result.WriteString(boxChars.TopLeft)
	if title != "" {
		titleFormatted := formatter.Header(fmt.Sprintf(" %s ", title))
		result.WriteString(titleFormatted)
		remaining := width - len(title) - 4
		if remaining > 0 {
			result.WriteString(strings.Repeat(boxChars.Horizontal, remaining))
		}
	} else {
		result.WriteString(strings.Repeat(boxChars.Horizontal, width-2))
	}
	result.WriteString(boxChars.TopRight)
	result.WriteString("\n")

	// Content lines
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		result.WriteString(boxChars.Vertical)
		result.WriteString(" ")
		
		// Pad line to width
		if len(line) < width-3 {
			result.WriteString(line)
			result.WriteString(strings.Repeat(" ", width-3-len(line)))
		} else {
			result.WriteString(line[:width-3])
		}
		
		result.WriteString(" ")
		result.WriteString(boxChars.Vertical)
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString(boxChars.BottomLeft)
	result.WriteString(strings.Repeat(boxChars.Horizontal, width-2))
	result.WriteString(boxChars.BottomRight)

	return result.String()

// RenderSeverityBar renders a colored severity indicator bar
func RenderSeverityBar(critical, high, medium, low int, width int, formatter *Formatter) string {
	total := critical + high + medium + low
	if total == 0 {
		return strings.Repeat("─", width)
	}

	// Calculate proportions
	criticalWidth := (critical * width) / total
	highWidth := (high * width) / total
	mediumWidth := (medium * width) / total
	lowWidth := width - criticalWidth - highWidth - mediumWidth

	var bar strings.Builder
	
	// Critical (red)
	if criticalWidth > 0 {
		bar.WriteString(formatter.format(formatter.scheme.Critical, strings.Repeat("█", criticalWidth)))
	}
	
	// High (red)
	if highWidth > 0 {
		bar.WriteString(formatter.format(formatter.scheme.High, strings.Repeat("█", highWidth)))
	}
	
	// Medium (yellow)
	if mediumWidth > 0 {
		bar.WriteString(formatter.format(formatter.scheme.Medium, strings.Repeat("█", mediumWidth)))
	}
	
	// Low (blue)
	if lowWidth > 0 {
		bar.WriteString(formatter.format(formatter.scheme.Low, strings.Repeat("█", lowWidth)))
	}

	return bar.String()

// RenderProgressBar renders a colored progress bar
func RenderProgressBar(current, total int, width int, formatter *Formatter) string {
	if total <= 0 {
		return strings.Repeat("─", width)
	}

	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))
	
	var bar strings.Builder
	bar.WriteString("[")
	
	// Determine color based on percentage
	var barColor *color.Color
	if percentage >= 1.0 {
		barColor = formatter.scheme.Success
	} else if percentage >= 0.7 {
		barColor = formatter.scheme.Value
	} else if percentage >= 0.4 {
		barColor = formatter.scheme.Warning
	} else {
		barColor = formatter.scheme.Error
	}
	
	// Filled portion
	if filled > 0 {
		bar.WriteString(formatter.format(barColor, strings.Repeat("=", filled-1)))
		if filled < width {
			bar.WriteString(formatter.format(barColor, ">"))
		} else {
			bar.WriteString(formatter.format(barColor, "="))
		}
	}
	
	// Empty portion
	if filled < width {
		bar.WriteString(strings.Repeat(" ", width-filled))
	}
	
	bar.WriteString("] ")
	bar.WriteString(formatter.format(barColor, "%.1f%%", percentage*100))
	
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
}
