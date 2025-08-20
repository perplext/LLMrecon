package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// TemplateInfo represents the structure of our template files
type TemplateInfo struct {
	ID   string                 `yaml:"id"`
	Info map[string]interface{} `yaml:"info"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run verify-template-format.go <templates-dir>")
		os.Exit(1)
	}

	templatesDir := os.Args[1]
	fmt.Printf("Verifying templates in: %s\n", templatesDir)

	// Walk through the templates directory
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
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

		// Read the template file (validate path first)
		cleanPath := filepath.Clean(path)
		data, err := ioutil.ReadFile(cleanPath) // #nosec G304 - Path is cleaned
		if err != nil {
			return fmt.Errorf("error reading template %s: %v", path, err)
		}

		// Parse the template
		var template TemplateInfo
		if err := yaml.Unmarshal(data, &template); err != nil {
			fmt.Printf("❌ Error parsing template %s: %v\n", path, err)
			return nil
		}

		// Verify template structure
		if template.ID == "" {
			fmt.Printf("❌ Template %s is missing 'id' field\n", path)
			return nil
		}

		if template.Info == nil {
			fmt.Printf("❌ Template %s is missing 'info' section\n", path)
			return nil
		}

		compliance, ok := template.Info["compliance"].(map[string]interface{})
		if !ok {
			fmt.Printf("❌ Template %s is missing 'compliance' section in 'info'\n", path)
			return nil
		}

		owaspLLM, ok := compliance["owasp-llm"].([]interface{})
		if !ok {
			fmt.Printf("❌ Template %s is missing 'owasp-llm' mappings in 'compliance'\n", path)
			return nil
		}

		// Verify OWASP LLM mappings
		for i, item := range owaspLLM {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				fmt.Printf("❌ Template %s has invalid mapping format in 'owasp-llm' at index %d\n", path, i)
				continue
			}

			category, ok := itemMap["category"].(string)
			if !ok {
				fmt.Printf("❌ Template %s is missing 'category' in OWASP LLM mapping at index %d\n", path, i)
				continue
			}

			subcategory, ok := itemMap["subcategory"].(string)
			if !ok {
				fmt.Printf("❌ Template %s is missing 'subcategory' in OWASP LLM mapping at index %d\n", path, i)
				continue
			}

			coverage, ok := itemMap["coverage"].(string)
			if !ok {
				fmt.Printf("❌ Template %s is missing 'coverage' in OWASP LLM mapping at index %d\n", path, i)
				continue
			}

			fmt.Printf("✅ Template %s has valid mapping: %s/%s (%s)\n", path, category, subcategory, coverage)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking templates directory: %v\n", err)
		os.Exit(1)
	}
}
