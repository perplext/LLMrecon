package ui

import (
	"fmt"
	"sort"
	"strings"
)

// TemplateSelector provides interactive template selection
type TemplateSelector struct {
	terminal  *Terminal
	templates []Template
	filters   TemplateFilters
}

// Template represents a test template
type Template struct {
	ID          string
	Name        string
	Description string
	Category    string
	Severity    string
	Author      string
	Tags        []string
	Path        string
}

// TemplateFilters for filtering templates
type TemplateFilters struct {
	Categories []string
	Severities []string
	Tags       []string
	Author     string
	Search     string
}

// NewTemplateSelector creates a new template selector
func NewTemplateSelector(terminal *Terminal) *TemplateSelector {
	return &TemplateSelector{
		terminal:  terminal,
		templates: make([]Template, 0),
		filters:   TemplateFilters{},
	}
}

// LoadTemplates loads available templates
func (ts *TemplateSelector) LoadTemplates(templates []Template) {
	ts.templates = templates
}

// SelectTemplates runs interactive template selection
func (ts *TemplateSelector) SelectTemplates() ([]Template, error) {
	if len(ts.templates) == 0 {
		return nil, fmt.Errorf("no templates available")
	}

	ts.terminal.Header("Interactive Template Selection")
	
	for {
		// Show main menu
		options := []string{
			"Browse all templates",
			"Filter by category",
			"Filter by severity",
			"Filter by tags",
			"Search templates",
			"Quick select (OWASP Top 10)",
			"View selected templates",
			"Clear filters",
			"Done selecting",
		}

		choice, err := ts.terminal.Select("What would you like to do?", options)
		if err != nil {
			return nil, err
		}

		switch choice {
		case 0:
			if err := ts.browseTemplates(); err != nil {
				return nil, err
			}
		case 1:
			if err := ts.filterByCategory(); err != nil {
				return nil, err
			}
		case 2:
			if err := ts.filterBySeverity(); err != nil {
				return nil, err
			}
		case 3:
			if err := ts.filterByTags(); err != nil {
				return nil, err
			}
		case 4:
			if err := ts.searchTemplates(); err != nil {
				return nil, err
			}
		case 5:
			if err := ts.quickSelectOWASP(); err != nil {
				return nil, err
			}
		case 6:
			ts.viewSelectedTemplates()
		case 7:
			ts.clearFilters()
			ts.terminal.Success("All filters cleared")
		case 8:
			// Done selecting
			selected := ts.getSelectedTemplates()
			if len(selected) == 0 {
				confirm, _ := ts.terminal.Confirm("No templates selected. Are you sure you want to exit?", false)
				if confirm {
					return []Template{}, nil
				}
			} else {
				return selected, nil
			}
		}
		
		fmt.Println()
	}
}

// browseTemplates allows browsing and selecting templates
func (ts *TemplateSelector) browseTemplates() error {
	filtered := ts.getFilteredTemplates()
	
	if len(filtered) == 0 {
		ts.terminal.Warning("No templates match current filters")
		return nil
	}

	// Group by category
	categorized := make(map[string][]Template)
	for _, tmpl := range filtered {
		categorized[tmpl.Category] = append(categorized[tmpl.Category], tmpl)
	}

	// Show categories
	categories := make([]string, 0, len(categorized))
	for cat := range categorized {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	ts.terminal.Info("Found %d templates in %d categories", len(filtered), len(categories))
	
	for _, category := range categories {
		templates := categorized[category]
		ts.terminal.Section(fmt.Sprintf("%s (%d templates)", category, len(templates)))
		
		// Create template list for selection
		templateNames := make([]string, len(templates))
		for i, tmpl := range templates {
			selected := ""
			if ts.isSelected(tmpl.ID) {
				selected = " [SELECTED]"
			}
			
			severityBadge := ts.formatSeverity(tmpl.Severity)
			templateNames[i] = fmt.Sprintf("%s %s%s\n   %s", 
				severityBadge, tmpl.Name, selected, tmpl.Description)
		}

		// Multi-select templates
		selected, err := ts.terminal.MultiSelect(
			fmt.Sprintf("Select templates from %s:", category), 
			templateNames,
		)
		if err != nil {
			continue
		}

		// Toggle selection
		for _, idx := range selected {
			ts.toggleTemplate(templates[idx])
		}
	}

	return nil
}

// filterByCategory filters templates by category
func (ts *TemplateSelector) filterByCategory() error {
	// Get unique categories
	categoryMap := make(map[string]int)
	for _, tmpl := range ts.templates {
		categoryMap[tmpl.Category]++
	}

	categories := make([]string, 0, len(categoryMap))
	for cat, count := range categoryMap {
		categories = append(categories, fmt.Sprintf("%s (%d templates)", cat, count))
	}
	sort.Strings(categories)

	selected, err := ts.terminal.MultiSelect("Select categories to filter by:", categories)
	if err != nil {
		return err
	}

	// Extract category names
	ts.filters.Categories = make([]string, len(selected))
	for i, idx := range selected {
		// Extract category name (before the count)
		parts := strings.Split(categories[idx], " (")
		ts.filters.Categories[i] = parts[0]
	}

	ts.terminal.Success("Filter applied: %d categories selected", len(ts.filters.Categories))
	return nil
}

// filterBySeverity filters templates by severity
func (ts *TemplateSelector) filterBySeverity() error {
	severities := []string{
		"Critical - Highest risk vulnerabilities",
		"High - Significant security risks",
		"Medium - Moderate security concerns",
		"Low - Minor security issues",
		"Info - Informational findings",
	}

	selected, err := ts.terminal.MultiSelect("Select severity levels:", severities)
	if err != nil {
		return err
	}

	ts.filters.Severities = make([]string, len(selected))
	for i, idx := range selected {
		switch idx {
		case 0:
			ts.filters.Severities[i] = "critical"
		case 1:
			ts.filters.Severities[i] = "high"
		case 2:
			ts.filters.Severities[i] = "medium"
		case 3:
			ts.filters.Severities[i] = "low"
		case 4:
			ts.filters.Severities[i] = "info"
		}
	}

	ts.terminal.Success("Filter applied: %d severity levels selected", len(ts.filters.Severities))
	return nil
}

// filterByTags filters templates by tags
func (ts *TemplateSelector) filterByTags() error {
	// Get unique tags
	tagMap := make(map[string]int)
	for _, tmpl := range ts.templates {
		for _, tag := range tmpl.Tags {
			tagMap[tag]++
		}
	}

	if len(tagMap) == 0 {
		ts.terminal.Warning("No tags available for filtering")
		return nil
	}

	tags := make([]string, 0, len(tagMap))
	for tag, count := range tagMap {
		tags = append(tags, fmt.Sprintf("%s (%d templates)", tag, count))
	}
	sort.Strings(tags)

	selected, err := ts.terminal.MultiSelect("Select tags to filter by:", tags)
	if err != nil {
		return err
	}

	ts.filters.Tags = make([]string, len(selected))
	for i, idx := range selected {
		parts := strings.Split(tags[idx], " (")
		ts.filters.Tags[i] = parts[0]
	}

	ts.terminal.Success("Filter applied: %d tags selected", len(ts.filters.Tags))
	return nil
}

// searchTemplates searches templates by keyword
func (ts *TemplateSelector) searchTemplates() error {
	search, err := ts.terminal.Prompt("Enter search keyword: ")
	if err != nil {
		return err
	}

	ts.filters.Search = strings.TrimSpace(search)
	
	if ts.filters.Search == "" {
		ts.terminal.Info("Search filter cleared")
	} else {
		matched := ts.getFilteredTemplates()
		ts.terminal.Success("Search filter applied: %d templates match '%s'", len(matched), ts.filters.Search)
	}

	return nil
}

// quickSelectOWASP quickly selects OWASP Top 10 templates
func (ts *TemplateSelector) quickSelectOWASP() error {
	owaspTemplates := []string{
		"LLM01 - Prompt Injection",
		"LLM02 - Insecure Output Handling",
		"LLM03 - Training Data Poisoning",
		"LLM04 - Model Denial of Service",
		"LLM05 - Supply Chain Vulnerabilities",
		"LLM06 - Sensitive Information Disclosure",
		"LLM07 - Insecure Plugin Design",
		"LLM08 - Excessive Agency",
		"LLM09 - Overreliance",
		"LLM10 - Model Theft",
	}

	selected, err := ts.terminal.MultiSelect("Select OWASP Top 10 categories:", owaspTemplates)
	if err != nil {
		return err
	}

	// Find and select matching templates
	selectedCount := 0
	for _, idx := range selected {
		category := fmt.Sprintf("LLM%02d", idx+1)
		for _, tmpl := range ts.templates {
			if strings.Contains(tmpl.Category, category) || 
			   strings.Contains(tmpl.ID, strings.ToLower(category)) ||
			   ts.hasTag(tmpl, fmt.Sprintf("owasp-llm-%02d", idx+1)) {
				ts.selectTemplate(tmpl)
				selectedCount++
			}
		}
	}

	ts.terminal.Success("Selected %d OWASP templates", selectedCount)
	return nil
}

// getFilteredTemplates returns templates matching current filters
func (ts *TemplateSelector) getFilteredTemplates() []Template {
	filtered := make([]Template, 0)
	
	for _, tmpl := range ts.templates {
		// Category filter
		if len(ts.filters.Categories) > 0 {
			found := false
			for _, cat := range ts.filters.Categories {
				if tmpl.Category == cat {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Severity filter
		if len(ts.filters.Severities) > 0 {
			found := false
			for _, sev := range ts.filters.Severities {
				if strings.EqualFold(tmpl.Severity, sev) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Tag filter
		if len(ts.filters.Tags) > 0 {
			found := false
			for _, filterTag := range ts.filters.Tags {
				if ts.hasTag(tmpl, filterTag) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Search filter
		if ts.filters.Search != "" {
			search := strings.ToLower(ts.filters.Search)
			if !strings.Contains(strings.ToLower(tmpl.Name), search) &&
			   !strings.Contains(strings.ToLower(tmpl.Description), search) &&
			   !strings.Contains(strings.ToLower(tmpl.ID), search) {
				continue
			}
		}

		filtered = append(filtered, tmpl)
	}

	return filtered
}

// Template selection tracking
var selectedTemplates = make(map[string]bool)

func (ts *TemplateSelector) isSelected(id string) bool {
	return selectedTemplates[id]
}

func (ts *TemplateSelector) selectTemplate(tmpl Template) {
	selectedTemplates[tmpl.ID] = true
}

func (ts *TemplateSelector) deselectTemplate(tmpl Template) {
	delete(selectedTemplates, tmpl.ID)
}

func (ts *TemplateSelector) toggleTemplate(tmpl Template) {
	if ts.isSelected(tmpl.ID) {
		ts.deselectTemplate(tmpl)
	} else {
		ts.selectTemplate(tmpl)
	}
}

func (ts *TemplateSelector) getSelectedTemplates() []Template {
	selected := make([]Template, 0)
	for _, tmpl := range ts.templates {
		if ts.isSelected(tmpl.ID) {
			selected = append(selected, tmpl)
		}
	}
	return selected
}

func (ts *TemplateSelector) viewSelectedTemplates() {
	selected := ts.getSelectedTemplates()
	
	if len(selected) == 0 {
		ts.terminal.Warning("No templates selected yet")
		return
	}

	ts.terminal.Header(fmt.Sprintf("Selected Templates (%d)", len(selected)))
	
	// Group by category
	categorized := make(map[string][]Template)
	for _, tmpl := range selected {
		categorized[tmpl.Category] = append(categorized[tmpl.Category], tmpl)
	}

	for category, templates := range categorized {
		ts.terminal.Subheader(category)
		for _, tmpl := range templates {
			ts.terminal.Print("  • %s %s", ts.formatSeverity(tmpl.Severity), tmpl.Name)
		}
	}
}

func (ts *TemplateSelector) clearFilters() {
	ts.filters = TemplateFilters{}
	selectedTemplates = make(map[string]bool)
}

func (ts *TemplateSelector) formatSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "[CRIT]"
	case "high":
		return "[HIGH]"
	case "medium":
		return "[MED]"
	case "low":
		return "[LOW]"
	case "info":
		return "[INFO]"
	default:
		return "[?]"
	}
}

func (ts *TemplateSelector) hasTag(tmpl Template, tag string) bool {
	for _, t := range tmpl.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// TemplateBrowser provides advanced template browsing
type TemplateBrowser struct {
	selector     *TemplateSelector
	currentView  string
	sortBy       string
	groupBy      string
	showDetails  bool
}

// NewTemplateBrowser creates a new template browser
func NewTemplateBrowser(selector *TemplateSelector) *TemplateBrowser {
	return &TemplateBrowser{
		selector:    selector,
		currentView: "list",
		sortBy:      "name",
		groupBy:     "category",
		showDetails: true,
	}
}

// Browse starts the browsing interface
func (tb *TemplateBrowser) Browse() error {
	for {
		tb.selector.terminal.Header("Template Browser")
		tb.showCurrentView()
		
		options := []string{
			"Change view (current: " + tb.currentView + ")",
			"Sort templates (by: " + tb.sortBy + ")",
			"Group templates (by: " + tb.groupBy + ")",
			"Toggle details (current: " + fmt.Sprintf("%v", tb.showDetails) + ")",
			"Select/deselect templates",
			"Back to main menu",
		}

		choice, err := tb.selector.terminal.Select("Browser options:", options)
		if err != nil {
			return err
		}

		switch choice {
		case 0:
			tb.changeView()
		case 1:
			tb.changeSorting()
		case 2:
			tb.changeGrouping()
		case 3:
			tb.showDetails = !tb.showDetails
		case 4:
			return tb.selectFromCurrentView()
		case 5:
			return nil
		}
	}
}

// showCurrentView displays templates in the current view
func (tb *TemplateBrowser) showCurrentView() {
	templates := tb.selector.getFilteredTemplates()
	
	switch tb.currentView {
	case "list":
		tb.showListView(templates)
	case "grid":
		tb.showGridView(templates)
	case "tree":
		tb.showTreeView(templates)
	}
}

// showListView shows templates in a list
func (tb *TemplateBrowser) showListView(templates []Template) {
	// Sort templates
	tb.sortTemplates(templates)
	
	// Group templates
	grouped := tb.groupTemplates(templates)
	
	for group, tmpls := range grouped {
		tb.selector.terminal.Section(fmt.Sprintf("%s (%d)", group, len(tmpls)))
		
		for _, tmpl := range tmpls {
			selected := ""
			if tb.selector.isSelected(tmpl.ID) {
				selected = " ✓"
			}
			
			tb.selector.terminal.Print("  %s %s%s", 
				tb.selector.formatSeverity(tmpl.Severity),
				tmpl.Name,
				selected,
			)
			
			if tb.showDetails {
				tb.selector.terminal.Print("    %s", tmpl.Description)
				if len(tmpl.Tags) > 0 {
					tb.selector.terminal.Print("    Tags: %s", strings.Join(tmpl.Tags, ", "))
				}
			}
		}
	}
}

// showGridView shows templates in a grid
func (tb *TemplateBrowser) showGridView(templates []Template) {
	// Implementation would show templates in a grid format
	tb.selector.terminal.Info("Grid view showing %d templates", len(templates))
	
	// Show summary counts by category and severity
	categoryCounts := make(map[string]int)
	severityCounts := make(map[string]int)
	
	for _, tmpl := range templates {
		categoryCounts[tmpl.Category]++
		severityCounts[tmpl.Severity]++
	}
	
	tb.selector.terminal.Print("\nBy Category:")
	for cat, count := range categoryCounts {
		tb.selector.terminal.Print("  %s: %d", cat, count)
	}
	
	tb.selector.terminal.Print("\nBy Severity:")
	for sev, count := range severityCounts {
		tb.selector.terminal.Print("  %s: %d", tb.selector.formatSeverity(sev), count)
	}
}

// showTreeView shows templates in a tree structure
func (tb *TemplateBrowser) showTreeView(templates []Template) {
	// Build tree structure
	tree := make(map[string]map[string][]Template)
	
	for _, tmpl := range templates {
		if tree[tmpl.Category] == nil {
			tree[tmpl.Category] = make(map[string][]Template)
		}
		tree[tmpl.Category][tmpl.Severity] = append(tree[tmpl.Category][tmpl.Severity], tmpl)
	}
	
	// Display tree
	for category, severities := range tree {
		tb.selector.terminal.Print("%s %s", "📁", category)
		
		for severity, tmpls := range severities {
			tb.selector.terminal.Print("  %s %s (%d)", "📂", tb.selector.formatSeverity(severity), len(tmpls))
			
			if tb.showDetails {
				for _, tmpl := range tmpls {
					selected := ""
					if tb.selector.isSelected(tmpl.ID) {
						selected = " ✓"
					}
					tb.selector.terminal.Print("    %s %s%s", "📄", tmpl.Name, selected)
				}
			}
		}
	}
}

// Helper methods
func (tb *TemplateBrowser) changeView() {
	views := []string{"list", "grid", "tree"}
	choice, _ := tb.selector.terminal.Select("Select view:", views)
	tb.currentView = views[choice]
}

func (tb *TemplateBrowser) changeSorting() {
	options := []string{"name", "severity", "category", "author"}
	choice, _ := tb.selector.terminal.Select("Sort by:", options)
	tb.sortBy = options[choice]
}

func (tb *TemplateBrowser) changeGrouping() {
	options := []string{"category", "severity", "author", "none"}
	choice, _ := tb.selector.terminal.Select("Group by:", options)
	tb.groupBy = options[choice]
}

func (tb *TemplateBrowser) sortTemplates(templates []Template) {
	sort.Slice(templates, func(i, j int) bool {
		switch tb.sortBy {
		case "name":
			return templates[i].Name < templates[j].Name
		case "severity":
			return getSeverityWeight(templates[i].Severity) > getSeverityWeight(templates[j].Severity)
		case "category":
			return templates[i].Category < templates[j].Category
		case "author":
			return templates[i].Author < templates[j].Author
		default:
			return templates[i].Name < templates[j].Name
		}
	})
}

func (tb *TemplateBrowser) groupTemplates(templates []Template) map[string][]Template {
	grouped := make(map[string][]Template)
	
	for _, tmpl := range templates {
		var key string
		switch tb.groupBy {
		case "category":
			key = tmpl.Category
		case "severity":
			key = tmpl.Severity
		case "author":
			key = tmpl.Author
		case "none":
			key = "All Templates"
		default:
			key = tmpl.Category
		}
		
		grouped[key] = append(grouped[key], tmpl)
	}
	
	return grouped
}

func (tb *TemplateBrowser) selectFromCurrentView() error {
	// Implementation would allow selecting templates from current view
	return nil
}

func getSeverityWeight(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 5
	case "high":
		return 4
	case "medium":
		return 3
	case "low":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}