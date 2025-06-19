package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/perplext/LLMrecon/src/config"
	"github.com/perplext/LLMrecon/src/update"
	"github.com/spf13/cobra"
)

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign [file]",
	Short: "Generate digital signatures for files",
	Long: `Generate digital signatures for files using cryptographic keys.
This command signs the specified file with a private key and outputs the signature.
It supports multiple signature algorithms including Ed25519, RSA, and ECDSA.

The private key can be provided via a file, environment variable, or configuration.
If no private key is provided, the command can generate a new key pair.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get the file path
		filePath := args[0]

		// Check if the file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File does not exist: %s\n", filePath)
			os.Exit(1)
		}

		// Get flags
		keyFile, _ := cmd.Flags().GetString("key-file")
		algorithm, _ := cmd.Flags().GetString("algorithm")
		outputFile, _ := cmd.Flags().GetString("output")
		generateKeys, _ := cmd.Flags().GetBool("generate-keys")
		keyOutputDir, _ := cmd.Flags().GetString("key-output-dir")
		calcChecksum, _ := cmd.Flags().GetBool("checksum")

		// Handle key generation if requested
		if generateKeys {
			// Convert algorithm string to SignatureAlgorithm
			var alg update.SignatureAlgorithm
			switch strings.ToLower(algorithm) {
			case "ed25519":
				alg = update.Ed25519Algorithm
			case "rsa":
				alg = update.RSAAlgorithm
			case "ecdsa":
				alg = update.ECDSAAlgorithm
			default:
				fmt.Fprintf(os.Stderr, "Error: Unsupported algorithm: %s\n", algorithm)
				os.Exit(1)
			}

			// Generate key pair
			fmt.Println("Generating new key pair...")
			privateKeyPEM, publicKeyPEM, err := update.GenerateKeyPair(alg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating key pair: %v\n", err)
				os.Exit(1)
			}

			// Ensure key output directory exists
			if keyOutputDir != "" {
				if err := os.MkdirAll(keyOutputDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating key output directory: %v\n", err)
					os.Exit(1)
				}
			}

			// Determine key file paths
			privateKeyPath := filepath.Join(keyOutputDir, fmt.Sprintf("private_key_%s.pem", strings.ToLower(algorithm)))
			publicKeyPath := filepath.Join(keyOutputDir, fmt.Sprintf("public_key_%s.pem", strings.ToLower(algorithm)))

			// Write keys to files
			if err := os.WriteFile(privateKeyPath, []byte(privateKeyPEM), 0600); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing private key: %v\n", err)
				os.Exit(1)
			}
			if err := os.WriteFile(publicKeyPath, []byte(publicKeyPEM), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing public key: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Key pair generated successfully:\n")
			fmt.Printf("Private key: %s\n", privateKeyPath)
			fmt.Printf("Public key: %s\n", publicKeyPath)

			// Use the generated private key for signing
			keyFile = privateKeyPath
		}

		// Get private key
		var privateKeyData string
		if keyFile != "" {
			// Read private key from file
			keyBytes, err := os.ReadFile(keyFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading private key file: %v\n", err)
				os.Exit(1)
			}
			privateKeyData = string(keyBytes)
		} else {
			// Try to get private key from config
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
				os.Exit(1)
			}
			privateKeyData = cfg.Security.PrivateKey
			if privateKeyData == "" {
				fmt.Fprintf(os.Stderr, "Error: No private key provided. Use --key-file or --generate-keys\n")
				os.Exit(1)
			}
		}

		// Create signature generator
		generator, err := update.NewSignatureGenerator(privateKeyData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating signature generator: %v\n", err)
			os.Exit(1)
		}

		// Generate signature
		fmt.Printf("Signing file: %s\n", filePath)
		signature, err := generator.GenerateSignature(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating signature: %v\n", err)
			os.Exit(1)
		}

		// Calculate checksum if requested
		if calcChecksum {
			checksum, err := update.CalculateChecksum(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error calculating checksum: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("SHA-256 Checksum: %s\n", checksum)
		}

		// Output signature
		if outputFile != "" {
			// Write signature to file
			if err := os.WriteFile(outputFile, []byte(signature), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing signature to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Signature written to: %s\n", outputFile)
		} else {
			// Print signature to stdout
			fmt.Printf("Signature: %s\n", signature)
		}
	},
}

func init() {
	rootCmd.AddCommand(signCmd)

	// Add flags
	signCmd.Flags().StringP("key-file", "k", "", "Path to the private key file")
	signCmd.Flags().StringP("algorithm", "a", "ed25519", "Signature algorithm (ed25519, rsa, ecdsa)")
	signCmd.Flags().StringP("output", "o", "", "Output file for the signature")
	signCmd.Flags().BoolP("generate-keys", "g", false, "Generate a new key pair")
	signCmd.Flags().String("key-output-dir", ".", "Directory to store generated keys")
	signCmd.Flags().BoolP("checksum", "c", false, "Calculate and output SHA-256 checksum")
}
