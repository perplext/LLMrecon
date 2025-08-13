package main

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/perplext/LLMrecon/src/template/manifest"
)

func main() {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		os.Exit(1)
	}
	
	// Path to examples directory
	examplesDir := filepath.Join(cwd, "examples")
	
	// Create manifest manager
	manager := manifest.NewManager(examplesDir)
	
	// Scan and register templates and modules
	fmt.Println("Scanning templates...")
	if err := manager.ScanAndRegisterTemplates(); err != nil {
		fmt.Printf("Error scanning templates: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Scanning modules...")
	if err := manager.ScanAndRegisterModules(); err != nil {
		fmt.Printf("Error scanning modules: %v\n", err)
		os.Exit(1)
	}
	
	// Save manifests
	if err := manager.SaveManifests(); err != nil {
		fmt.Printf("Error saving manifests: %v\n", err)
		os.Exit(1)
	}
	
	// Print summary
	templateManifest := manager.GetTemplateManifest()
	moduleManifest := manager.GetModuleManifest()
	
	fmt.Println("\nScan complete!")
	fmt.Printf("Found %d templates and %d modules.\n", 
		len(templateManifest.Templates), len(moduleManifest.Modules))
	
	// Print templates by category
	fmt.Println("\nTemplates by Category:")
	fmt.Println("=====================")
	
	for category, categoryInfo := range templateManifest.Categories {
		fmt.Printf("\nCategory: %s\n", category)
		fmt.Printf("Description: %s\n", categoryInfo.Description)
		fmt.Println("Templates:")
		
		for _, id := range categoryInfo.Templates {
			if template, exists := templateManifest.Templates[id]; exists {
				fmt.Printf("  - %s (v%s): %s\n", template.Name, template.Version, template.Description)
				fmt.Printf("    Severity: %s\n", template.Severity)
				fmt.Printf("    Tags: %v\n", template.Tags)
				fmt.Printf("    Path: %s\n", template.Path)
			}
		}
	}
	
	// Print modules by type
	fmt.Println("\nModules by Type:")
	fmt.Println("===============")
	
	for moduleType, typeInfo := range moduleManifest.Types {
		fmt.Printf("\nType: %s\n", moduleType)
		fmt.Printf("Description: %s\n", typeInfo.Description)
		fmt.Println("Modules:")
		
		for _, id := range typeInfo.Modules {
			if module, exists := moduleManifest.Modules[id]; exists {
				fmt.Printf("  - %s (v%s): %s\n", module.Name, module.Version, module.Description)
				fmt.Printf("    Tags: %v\n", module.Tags)
				fmt.Printf("    Path: %s\n", module.Path)
			}
		}
	}
}
