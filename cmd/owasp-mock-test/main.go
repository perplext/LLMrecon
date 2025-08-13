// Package main provides a command-line application for testing OWASP vulnerabilities with mock providers
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/testing/owasp/examples"
	"github.com/perplext/LLMrecon/src/testing/owasp"
)

func main() {
	// Define command-line flags
	modePtr := flag.String("mode", "all", "Test mode: all, vulnerability, provider")
	vulnPtr := flag.String("vulnerability", "", "Vulnerability type to test (required for vulnerability mode)")
	providerPtr := flag.String("provider", "", "Provider type to test (required for provider mode)")
	
	// Parse command-line flags
	flag.Parse()
	
	// Validate flags
	switch *modePtr {
	case "all":
		// Run all tests
		examples.RunMockProviderExample()
	case "vulnerability":
		// Validate vulnerability type
		if *vulnPtr == "" {
			fmt.Println("Error: vulnerability type is required for vulnerability mode")
			printUsage()
			os.Exit(1)
		}
		
		// Convert string to vulnerability type
		vulnerabilityType := parseVulnerabilityType(*vulnPtr)
		if vulnerabilityType == "" {
			fmt.Printf("Error: unknown vulnerability type: %s\n", *vulnPtr)
			printVulnerabilityTypes()
			os.Exit(1)
		}
		
		// Run vulnerability test
		examples.RunCustomVulnerabilityTest(vulnerabilityType)
	case "provider":
		// Validate provider type
		if *providerPtr == "" {
			fmt.Println("Error: provider type is required for provider mode")
			printUsage()
			os.Exit(1)
		}
		
		// Convert string to provider type
		providerType := parseProviderType(*providerPtr)
		if providerType == "" {
			fmt.Printf("Error: unknown provider type: %s\n", *providerPtr)
			printProviderTypes()
			os.Exit(1)
		}
		
		// Run provider test
		examples.RunCustomProviderTest(providerType)
	default:
		fmt.Printf("Error: unknown mode: %s\n", *modePtr)
		printUsage()
		os.Exit(1)
	}
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  -mode=all                    Run all tests")
	fmt.Println("  -mode=vulnerability -vulnerability=<type>  Run tests for a specific vulnerability type")
	fmt.Println("  -mode=provider -provider=<type>  Run tests for a specific provider type")
	fmt.Println("")
	printVulnerabilityTypes()
	fmt.Println("")
	printProviderTypes()
}

// printVulnerabilityTypes prints the available vulnerability types
func printVulnerabilityTypes() {
	fmt.Println("Available vulnerability types:")
	fmt.Println("  prompt-injection            LLM01: Prompt Injection")
	fmt.Println("  insecure-output             LLM02: Insecure Output Handling")
	fmt.Println("  training-data-poisoning     LLM03: Training Data Poisoning")
	fmt.Println("  model-dos                   LLM04: Model Denial of Service")
	fmt.Println("  supply-chain                LLM05: Supply Chain Vulnerabilities")
	fmt.Println("  sensitive-info              LLM06: Sensitive Information Disclosure")
	fmt.Println("  insecure-plugin             LLM07: Insecure Plugin Design")
	fmt.Println("  excessive-agency            LLM08: Excessive Agency")
	fmt.Println("  overreliance                LLM09: Overreliance")
	fmt.Println("  model-theft                 LLM10: Model Theft")
}

// printProviderTypes prints the available provider types
func printProviderTypes() {
	fmt.Println("Available provider types:")
	fmt.Println("  openai                      OpenAI (GPT models)")
	fmt.Println("  anthropic                   Anthropic (Claude models)")
	fmt.Println("  google                      Google (Gemini models)")
}

// parseVulnerabilityType converts a string to a vulnerability type
func parseVulnerabilityType(s string) owasp.VulnerabilityType {
	s = strings.ToLower(s)
	switch s {
	case "prompt-injection", "llm01":
		return owasp.PromptInjection
	case "insecure-output", "llm02":
		return owasp.InsecureOutput
	case "training-data-poisoning", "llm03":
		return owasp.TrainingDataPoisoning
	case "model-dos", "llm04":
		return owasp.ModelDOS
	case "supply-chain", "llm05":
		return owasp.SupplyChainVulnerabilities
	case "sensitive-info", "llm06":
		return owasp.SensitiveInformationDisclosure
	case "insecure-plugin", "llm07":
		return owasp.InsecurePluginDesign
	case "excessive-agency", "llm08":
		return owasp.ExcessiveAgency
	case "overreliance", "llm09":
		return owasp.Overreliance
	case "model-theft", "llm10":
		return owasp.ModelTheft
	default:
		return ""
	}
}

// parseProviderType converts a string to a provider type
func parseProviderType(s string) core.ProviderType {
	s = strings.ToLower(s)
	switch s {
	case "openai":
		return core.OpenAIProvider
	case "anthropic":
		return core.AnthropicProvider
	case "google", "gemini":
		return core.GoogleProvider
	default:
		return ""
	}
}
