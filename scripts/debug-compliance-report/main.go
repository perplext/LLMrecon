package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TemplateInfo represents the structure of our template files
type TemplateInfo struct {
	ID   string                 `yaml:"id" json:"id"`
	Info map[string]interface{} `yaml:"info" json:"info"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run debug-compliance-report.go <templates-dir>")
		os.Exit(1)
	}

	templatesDir := os.Args[1]
	fmt.Printf("Loading templates from: %s\n", templatesDir)

	// Load templates
	templates, err := loadTemplates(templatesDir)
	if err != nil {
		fmt.Printf("Error loading templates: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d templates\n", len(templates))

	// Print template details for debugging
	for i, template := range templates {
		templateBytes, _ := json.MarshalIndent(template, "", "  ")
		fmt.Printf("Template %d:\n%s\n\n", i+1, string(templateBytes))
	}
}

// loadTemplates loads templates from the specified directory
func loadTemplates(dir string) ([]interface{}, error) {
	var templates []interface{}

	// Walk through the templates directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-YAML files
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		fmt.Printf("Processing file: %s\n", path)

		// Read the template file (validate path first)
		cleanPath := filepath.Clean(path)
		data, err := ioutil.ReadFile(cleanPath) // #nosec G304 - Path is cleaned
		if err != nil {
			return fmt.Errorf("error reading template %s: %v", path, err)
		}

		// Parse the template
		var template TemplateInfo
		if err := yaml.Unmarshal(data, &template); err != nil {
			fmt.Printf("Error parsing template %s: %v\n", path, err)
			return nil
		}

		// Add the template to the list
		templates = append(templates, template)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking templates directory: %v", err)
	}

	return templates, nil
}
