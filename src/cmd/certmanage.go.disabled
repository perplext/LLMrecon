package cmd

import (
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/security/communication"
	"github.com/spf13/cobra"
)

var (
	// Certificate management options
	certFormat        string
	certIsRoot        bool
	certIsIntermediate bool
	certDirectory     string
	certOutputFile    string
	certValidate      bool
	certCheckCRL      bool
	certCheckOCSP     bool
	certInfo          bool
	certChain         bool
)

// certManageCmd represents the certmanage command
var certManageCmd = &cobra.Command{
	Use:   "certmanage",
	Short: "Manage certificates and trust chains",
	Long: `Manage certificates and trust chains for the LLMrecon tool.
This command allows you to import, export, validate, and manage certificates
and trust chains for secure communication.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// certImportCmd represents the certmanage import command
var certImportCmd = &cobra.Command{
	Use:   "import [certificate file]",
	Short: "Import a certificate",
	Long: `Import a certificate into the trust store.
The certificate can be a root certificate or an intermediate certificate.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Create trust chain manager
		manager := communication.NewTrustChainManager()

		// Import certificate
		var err error
		if certDirectory != "" {
			// Import from directory
			fmt.Printf("Importing certificates from %s...\n", certDirectory)
			if certIsRoot {
				err = manager.ImportCertificatesFromDirectory(certDirectory, true)
			} else if certIsIntermediate {
				err = manager.ImportCertificatesFromDirectory(certDirectory, false)
			} else {
				fmt.Println("Please specify whether to import as root or intermediate certificates.")
				return
			}
		} else {
			// Import from file
			certFile := args[0]
			fmt.Printf("Importing certificate from %s...\n", certFile)
			if certIsRoot {
				err = manager.AddRootCertificateFromFile(certFile)
			} else if certIsIntermediate {
				err = manager.AddIntermediateCertificateFromFile(certFile)
			} else {
				fmt.Println("Please specify whether to import as a root or intermediate certificate.")
				return
			}
		}

		if err != nil {
			fmt.Printf("Error importing certificate: %v\n", err)
			return
		}

		fmt.Println("Certificate imported successfully.")

		// Validate certificate if requested
		if certValidate {
			validateCertificate(manager, args[0])
		}
	},
}

// certValidateCmd represents the certmanage validate command
var certValidateCmd = &cobra.Command{
	Use:   "validate [certificate file]",
	Short: "Validate a certificate",
	Long: `Validate a certificate against the trust store.
The certificate will be checked for validity, expiration, and revocation status.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Create trust chain manager
		manager := communication.NewTrustChainManager()

		// Set CRL and OCSP checking
		manager.SetCRLCheckEnabled(certCheckCRL)
		manager.SetOCSPCheckEnabled(certCheckOCSP)

		// Validate certificate
		validateCertificate(manager, args[0])
	},
}

// certInfoCmd represents the certmanage info command
var certInfoCmd = &cobra.Command{
	Use:   "info [certificate file]",
	Short: "Display certificate information",
	Long: `Display detailed information about a certificate.
The information includes subject, issuer, validity period, and key usage.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Read certificate file
		certFile := args[0]
		pemData, err := os.ReadFile(certFile)
		if err != nil {
			fmt.Printf("Error reading certificate file: %v\n", err)
			return
		}

		// Parse certificate
		var certs []*x509.Certificate
		if certChain {
			// Parse certificate chain
			certs, err = communication.GetCertificateChainFromPEM(pemData)
			if err != nil {
				fmt.Printf("Error parsing certificate chain: %v\n", err)
				return
			}
		} else {
			// Parse single certificate
			cert, err := communication.ParseCertificateFromPEM(pemData)
			if err != nil {
				fmt.Printf("Error parsing certificate: %v\n", err)
				return
			}
			certs = []*x509.Certificate{cert}
		}

		// Display certificate information
		for i, cert := range certs {
			if len(certs) > 1 {
				fmt.Printf("Certificate %d of %d:\n", i+1, len(certs))
			}
			fmt.Println(communication.FormatCertificateInfo(cert))
			if i < len(certs)-1 {
				fmt.Println("---")
			}
		}
	},
}

// certExportCmd represents the certmanage export command
var certExportCmd = &cobra.Command{
	Use:   "export [certificate file]",
	Short: "Export a certificate",
	Long: `Export a certificate to a file.
The certificate can be exported in PEM or DER format.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Read certificate file
		certFile := args[0]
		pemData, err := os.ReadFile(certFile)
		if err != nil {
			fmt.Printf("Error reading certificate file: %v\n", err)
			return
		}

		// Parse certificate
		cert, err := communication.ParseCertificateFromPEM(pemData)
		if err != nil {
			fmt.Printf("Error parsing certificate: %v\n", err)
			return
		}

		// Determine output file
		outputFile := certOutputFile
		if outputFile == "" {
			ext := ".pem"
			if strings.ToLower(certFormat) == "der" {
				ext = ".der"
			}
			outputFile = strings.TrimSuffix(certFile, filepath.Ext(certFile)) + ext
		}

		// Determine format
		var format communication.CertificateFormat
		switch strings.ToLower(certFormat) {
		case "pem":
			format = communication.CertFormatPEM
		case "der":
			format = communication.CertFormatDER
		default:
			fmt.Printf("Unsupported certificate format: %s\n", certFormat)
			return
		}

		// Export certificate
		err = communication.ExportCertificate(cert, outputFile, format)
		if err != nil {
			fmt.Printf("Error exporting certificate: %v\n", err)
			return
		}

		fmt.Printf("Certificate exported to %s\n", outputFile)
	},
}

// validateCertificate validates a certificate and displays the result
func validateCertificate(manager *communication.TrustChainManager, certFile string) {
	// Validate certificate
	fmt.Printf("Validating certificate %s...\n", certFile)
	certInfo, err := manager.ValidateCertificateFromFile(certFile)
	if err != nil {
		fmt.Printf("Certificate validation failed: %v\n", err)
		return
	}

	// Display validation result
	fmt.Println("Certificate validation result:")
	fmt.Printf("  Status: %s\n", formatCertificateStatus(certInfo.Status))
	if certInfo.ValidationError != nil {
		fmt.Printf("  Error: %v\n", certInfo.ValidationError)
	}
	if certInfo.TrustChain != nil {
		fmt.Printf("  Trust chain length: %d\n", len(certInfo.TrustChain))
		for i, cert := range certInfo.TrustChain {
			fmt.Printf("    %d: %s (Issuer: %s)\n", i+1, cert.Subject.CommonName, cert.Issuer.CommonName)
		}
	}
}

// formatCertificateStatus formats a certificate status as a string
func formatCertificateStatus(status communication.CertificateStatus) string {
	switch status {
	case communication.CertStatusValid:
		return "Valid"
	case communication.CertStatusExpired:
		return "Expired"
	case communication.CertStatusRevoked:
		return "Revoked"
	case communication.CertStatusUntrusted:
		return "Untrusted"
	case communication.CertStatusInvalid:
		return "Invalid"
	default:
		return "Unknown"
	}
}

func init() {
	rootCmd.AddCommand(certManageCmd)
	certManageCmd.AddCommand(certImportCmd)
	certManageCmd.AddCommand(certValidateCmd)
	certManageCmd.AddCommand(certInfoCmd)
	certManageCmd.AddCommand(certExportCmd)

	// Add flags for import command
	certImportCmd.Flags().BoolVar(&certIsRoot, "root", false, "Import as a root certificate")
	certImportCmd.Flags().BoolVar(&certIsIntermediate, "intermediate", false, "Import as an intermediate certificate")
	certImportCmd.Flags().StringVar(&certDirectory, "dir", "", "Import certificates from a directory")
	certImportCmd.Flags().BoolVar(&certValidate, "validate", false, "Validate the certificate after import")

	// Add flags for validate command
	certValidateCmd.Flags().BoolVar(&certCheckCRL, "check-crl", true, "Check certificate revocation lists")
	certValidateCmd.Flags().BoolVar(&certCheckOCSP, "check-ocsp", false, "Check OCSP responders")

	// Add flags for info command
	certInfoCmd.Flags().BoolVar(&certChain, "chain", false, "Parse as a certificate chain")

	// Add flags for export command
	certExportCmd.Flags().StringVar(&certFormat, "format", "pem", "Certificate format (pem or der)")
	certExportCmd.Flags().StringVar(&certOutputFile, "output", "", "Output file")
}
