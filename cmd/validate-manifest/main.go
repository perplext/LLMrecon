package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// Import the bundle package using the correct import path based on the module name
// This will need to be adjusted based on the actual Go module name in go.mod
import "github.com/perplext/LLMrecon/src/bundle"

func main() {
	// Define command-line flags
	manifestPath := flag.String("manifest", "", "Path to the manifest file to validate")
	schemaPath := flag.String("schema", "", "Path to the schema file (optional)")
	generateExample := flag.Bool("generate-example", false, "Generate an example manifest file")
	outputPath := flag.String("output", "example-manifest.json", "Output path for the example manifest")
	verbose := flag.Bool("verbose", false, "Enable verbose output")

	// Parse command-line flags
	flag.Parse()

	// Check if we should generate an example manifest
	if *generateExample {
		if err := bundle.SaveExampleManifest(*outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated example manifest at %s\n", *outputPath)
		return
	}

	// Check if a manifest path was provided
	if *manifestPath == "" {
		fmt.Println("Error: No manifest path provided")
		fmt.Println("Usage: validate-manifest --manifest=<path> [--schema=<path>] [--verbose]")
		fmt.Println("       validate-manifest --generate-example [--output=<path>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create a schema validator
	var validator *bundle.SchemaValidator
	var err error

	if *schemaPath != "" {
		// Use the provided schema path
		validator, err = bundle.NewSchemaValidator(*schemaPath)
	} else {
		// Use the default schema path
		validator, err = bundle.NewDefaultSchemaValidator()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Validate the manifest
	result, err := validator.ValidateManifestFile(*manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Print the validation result
	if result.IsValid {
		color.Green("✓ Manifest is valid")
	} else {
		color.Red("✗ Manifest is invalid")
		fmt.Println("Errors:")
		for _, errMsg := range result.Errors {
			fmt.Printf("  - %s\n", errMsg)
		}
	}

	// If verbose output is enabled, print the full validation result
	if *verbose {
		fmt.Println("\nValidation Result:")
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling validation result: %s\n", err)
		} else {
			fmt.Println(string(resultJSON))
		}
	}

	// Exit with an appropriate status code
	if !result.IsValid {
		os.Exit(1)
	}
}
