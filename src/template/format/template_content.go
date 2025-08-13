package format

// TemplateContent represents the content of a template
type TemplateContent struct {
	// Sections is a list of template sections
	Sections []TemplateSection
	// Variables is a map of variable names to values
	Variables map[string]interface{}
}

// TemplateSection represents a section of a template
type TemplateSection struct {
	// Type is the type of the section
	Type string
	// Content is the content of the section
	Content string
}

// NewTemplateContent creates a new template content
func NewTemplateContent() *TemplateContent {
	return &TemplateContent{
		Sections:  make([]TemplateSection, 0),
		Variables: make(map[string]interface{}),
	}
}

// AddSection adds a section to the template content
func (c *TemplateContent) AddSection(sectionType, content string) {
	c.Sections = append(c.Sections, TemplateSection{
		Type:    sectionType,
		Content: content,
	})
}

// AddVariable adds a variable to the template content
func (c *TemplateContent) AddVariable(name string, value interface{}) {
	c.Variables[name] = value
}

// GetVariable gets a variable from the template content
func (c *TemplateContent) GetVariable(name string) (interface{}, bool) {
	value, ok := c.Variables[name]
	return value, ok
}

// GetSections gets all sections of a specific type
func (c *TemplateContent) GetSections(sectionType string) []TemplateSection {
	var result []TemplateSection
	for _, section := range c.Sections {
		if section.Type == sectionType {
			result = append(result, section)
		}
	}
	return result
}

// GetSectionContent gets the content of the first section of a specific type
func (c *TemplateContent) GetSectionContent(sectionType string) (string, bool) {
	for _, section := range c.Sections {
		if section.Type == sectionType {
			return section.Content, true
		}
	}
	return "", false
}

// Clone creates a deep copy of the template content
func (c *TemplateContent) Clone() *TemplateContent {
	clone := &TemplateContent{
		Sections:  make([]TemplateSection, len(c.Sections)),
		Variables: make(map[string]interface{}, len(c.Variables)),
	}

	// Clone sections
	for i, section := range c.Sections {
		clone.Sections[i] = TemplateSection{
			Type:    section.Type,
			Content: section.Content,
		}
	}

	// Clone variables
	for name, value := range c.Variables {
		clone.Variables[name] = value
	}

	return clone
}
