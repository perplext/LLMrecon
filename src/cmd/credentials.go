package cmd

import (
	"os"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/audit"
	securityaudit "github.com/perplext/LLMrecon/src/security/audit"
	"github.com/perplext/LLMrecon/src/security/vault"
	"github.com/spf13/cobra"
)

var (
	// Credential command flags
	credentialService  string
	credentialType     string
	credentialValue    string
	credentialDesc     string
	credentialTags     string
	credentialRotation int
	credentialWarning  int
)

// credentialCmd represents the credential command
var credentialCmd = &cobra.Command{
	Use:   "credential",
	Short: "Manage credentials securely",
	Long: `Manage credentials securely for the LLMreconing Tool.
This command allows you to list, add, update, and delete credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},

// credentialListCmd represents the credential list command
var credentialListCmd = &cobra.Command{
	Use:   "list",
	Short: "List credentials",
	Long:  `List all credentials or filter by service, type, or tag.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize credential manager
		if err := initCredentialManager(); err != nil {
			fmt.Printf("Error initializing credential manager: %v\n", err)
			os.Exit(1)
		}

		var credentials []*vault.Credential
		var err error

		// Filter by service, type, or tag
		if credentialService != "" {
			credentials, err = vault.DefaultManager.ListCredentialsByService(credentialService)
		} else if credentialType != "" {
			credentials, err = vault.DefaultManager.ListCredentialsByType(vault.CredentialType(credentialType))
		} else if credentialTags != "" {
			// Split tags and list by each tag
			tags := strings.Split(credentialTags, ",")
			for _, tag := range tags {
				tagCredentials, tagErr := vault.DefaultManager.ListCredentialsByTag(strings.TrimSpace(tag))
				if tagErr != nil {
					fmt.Printf("Error listing credentials by tag '%s': %v\n", tag, tagErr)
					continue
				}
				credentials = append(credentials, tagCredentials...)
			}
		} else {
			credentials, err = vault.DefaultManager.ListCredentials()
		}

		if err != nil {
			fmt.Printf("Error listing credentials: %v\n", err)
			os.Exit(1)
		}

		// Display credentials
		if len(credentials) == 0 {
			fmt.Println("No credentials found.")
			return
		}

		fmt.Println("ID\tName\tService\tType\tLast Used\tExpires")
		fmt.Println("--\t----\t-------\t----\t---------\t-------")
		for _, cred := range credentials {
			lastUsed := "Never"
			if !cred.LastUsedAt.IsZero() {
				lastUsed = cred.LastUsedAt.Format("2006-01-02")
			}

			expires := "Never"
			if !cred.ExpiresAt.IsZero() {
				expires = cred.ExpiresAt.Format("2006-01-02")
			} else if cred.RotationPolicy != nil && cred.RotationPolicy.Enabled {
				var nextRotation time.Time
				if !cred.RotationPolicy.LastRotation.IsZero() {
					nextRotation = cred.RotationPolicy.LastRotation.AddDate(0, 0, cred.RotationPolicy.IntervalDays)
				} else {
					nextRotation = cred.CreatedAt.AddDate(0, 0, cred.RotationPolicy.IntervalDays)
				}
				expires = nextRotation.Format("2006-01-02")
			}

			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n",
				cred.ID,
				cred.Name,
				cred.Service,
				cred.Type,
				lastUsed,
				expires,
			)
		}
	},

// credentialShowCmd represents the credential show command
var credentialShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show credential details",
	Long:  `Show detailed information about a credential.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize credential manager
		if err := initCredentialManager(); err != nil {
			fmt.Printf("Error initializing credential manager: %v\n", err)
			os.Exit(1)
		}

		// Get credential
		cred, err := vault.DefaultManager.GetCredential(args[0])
		if err != nil {
			fmt.Printf("Error getting credential: %v\n", err)
			os.Exit(1)
		}

		// Display credential details
		fmt.Println("ID:", cred.ID)
		fmt.Println("Name:", cred.Name)
		fmt.Println("Service:", cred.Service)
		fmt.Println("Type:", cred.Type)
		fmt.Println("Description:", cred.Description)
		fmt.Println("Tags:", strings.Join(cred.Tags, ", "))
		fmt.Println("Created:", cred.CreatedAt.Format(time.RFC3339))
		fmt.Println("Updated:", cred.UpdatedAt.Format(time.RFC3339))

		if !cred.LastUsedAt.IsZero() {
			fmt.Println("Last Used:", cred.LastUsedAt.Format(time.RFC3339))
		}

		if !cred.ExpiresAt.IsZero() {
			fmt.Println("Expires:", cred.ExpiresAt.Format(time.RFC3339))
		}

		if cred.RotationPolicy != nil && cred.RotationPolicy.Enabled {
			fmt.Println("Rotation Policy:")
			fmt.Println("  Enabled:", cred.RotationPolicy.Enabled)
			fmt.Println("  Interval:", cred.RotationPolicy.IntervalDays, "days")
			fmt.Println("  Warning:", cred.RotationPolicy.WarningDays, "days before")

			if !cred.RotationPolicy.LastRotation.IsZero() {
				fmt.Println("  Last Rotation:", cred.RotationPolicy.LastRotation.Format(time.RFC3339))
				nextRotation := cred.RotationPolicy.LastRotation.AddDate(0, 0, cred.RotationPolicy.IntervalDays)
				fmt.Println("  Next Rotation:", nextRotation.Format(time.RFC3339))
			}
		}

		// Only show value if explicitly requested
		if cmd.Flag("show-value").Value.String() == "true" {
			fmt.Println("Value:", cred.Value)
		} else {
			fmt.Println("Value: [hidden, use --show-value to display]")
		}
	},

// credentialAddCmd represents the credential add command
var credentialAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new credential",
	Long:  `Add a new credential to the secure vault.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize credential manager
		if err := initCredentialManager(); err != nil {
			fmt.Printf("Error initializing credential manager: %v\n", err)
			os.Exit(1)
		}

		// Validate required fields
		if credentialService == "" {
			fmt.Println("Error: service is required")
			cmd.Help()
			os.Exit(1)
		}
		if credentialType == "" {
			fmt.Println("Error: type is required")
			cmd.Help()
			os.Exit(1)
		}
		if credentialValue == "" {
			fmt.Println("Error: value is required")
			cmd.Help()
			os.Exit(1)
		}

		// Create credential
		cred := &vault.Credential{
			ID:          vault.GenerateCredentialID(credentialService, credentialType),
			Name:        fmt.Sprintf("%s %s", strings.Title(credentialService), strings.Title(credentialType)),
			Type:        vault.CredentialType(credentialType),
			Service:     credentialService,
			Value:       credentialValue,
			Description: credentialDesc,
			Tags:        []string{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Add tags
		if credentialTags != "" {
			cred.Tags = strings.Split(credentialTags, ",")
			for i, tag := range cred.Tags {
				cred.Tags[i] = strings.TrimSpace(tag)
			}
		}

		// Add rotation policy if specified
		if credentialRotation > 0 {
			cred.RotationPolicy = &vault.RotationPolicy{
				Enabled:      true,
				IntervalDays: credentialRotation,
				WarningDays:  credentialWarning,
				LastRotation: time.Now(),
			}
		}

		// Store credential
		if err := vault.DefaultManager.StoreCredential(cred); err != nil {
			fmt.Printf("Error storing credential: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Credential added with ID: %s\n", cred.ID)
	},
// credentialUpdateCmd represents the credential update command
var credentialUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update a credential",
	Long:  `Update an existing credential in the secure vault.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize credential manager
		if err := initCredentialManager(); err != nil {
			fmt.Printf("Error initializing credential manager: %v\n", err)
			os.Exit(1)
		}

		// Get existing credential
		cred, err := vault.DefaultManager.GetCredential(args[0])
		if err != nil {
			fmt.Printf("Error getting credential: %v\n", err)
			os.Exit(1)
		}

		// Update fields if specified
		if credentialValue != "" {
			cred.Value = credentialValue
		}
		if credentialDesc != "" {
			cred.Description = credentialDesc
		}
		if credentialTags != "" {
			cred.Tags = strings.Split(credentialTags, ",")
			for i, tag := range cred.Tags {
				cred.Tags[i] = strings.TrimSpace(tag)
			}
		}

		// Update rotation policy if specified
		if credentialRotation > 0 {
			if cred.RotationPolicy == nil {
				cred.RotationPolicy = &vault.RotationPolicy{
					Enabled:      true,
					LastRotation: time.Now(),
				}
			}
			cred.RotationPolicy.IntervalDays = credentialRotation
			cred.RotationPolicy.WarningDays = credentialWarning
		}

		// Update timestamp
		cred.UpdatedAt = time.Now()

		// Store updated credential
		if err := vault.DefaultManager.StoreCredential(cred); err != nil {
			fmt.Printf("Error updating credential: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Credential updated: %s\n", cred.ID)
	},

// credentialDeleteCmd represents the credential delete command
var credentialDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a credential",
	Long:  `Delete a credential from the secure vault.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize credential manager
		if err := initCredentialManager(); err != nil {
			fmt.Printf("Error initializing credential manager: %v\n", err)
			os.Exit(1)
		}
		// Delete credential
		if err := vault.DefaultManager.DeleteCredential(args[0]); err != nil {
			fmt.Printf("Error deleting credential: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Credential deleted: %s\n", args[0])
	},

// credentialRotateCmd represents the credential rotate command
var credentialRotateCmd = &cobra.Command{
	Use:   "rotate [id]",
	Short: "Rotate a credential",
	Long:  `Rotate a credential by updating its value and rotation timestamp.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize credential manager
		if err := initCredentialManager(); err != nil {
			fmt.Printf("Error initializing credential manager: %v\n", err)
			os.Exit(1)
		}

		// Validate required fields
		if credentialValue == "" {
			fmt.Println("Error: value is required for rotation")
			cmd.Help()
			os.Exit(1)
		}

		// Rotate credential
		if err := vault.DefaultManager.RotateCredential(args[0], credentialValue); err != nil {
			fmt.Printf("Error rotating credential: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Credential rotated: %s\n", args[0])
	},

// credentialCheckRotationCmd represents the credential check-rotation command
var credentialCheckRotationCmd = &cobra.Command{
	Use:   "check-rotation",
	Short: "Check for credentials that need rotation",
	Long:  `Check for credentials that need rotation based on their rotation policy.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize credential manager
		if err := initCredentialManager(); err != nil {
			fmt.Printf("Error initializing credential manager: %v\n", err)
			os.Exit(1)
		}

		// Get credentials needing rotation
		credentials, err := vault.DefaultManager.GetCredentialsNeedingRotation()
		if err != nil {
			fmt.Printf("Error checking credentials for rotation: %v\n", err)
			os.Exit(1)
		}

		// Display credentials
		if len(credentials) == 0 {
			fmt.Println("No credentials need rotation.")
			return
		}

		fmt.Println("The following credentials need rotation:")
		fmt.Println("ID\tName\tService\tType\tLast Rotation\tDays Overdue")
		fmt.Println("--\t----\t-------\t----\t-------------\t------------")
		for _, cred := range credentials {
			var lastRotation time.Time
			if cred.RotationPolicy != nil && !cred.RotationPolicy.LastRotation.IsZero() {
				lastRotation = cred.RotationPolicy.LastRotation
			} else {
				lastRotation = cred.CreatedAt
			}

			nextRotation := lastRotation.AddDate(0, 0, cred.RotationPolicy.IntervalDays)
			daysOverdue := int(time.Since(nextRotation).Hours() / 24)

			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%d\n",
				cred.ID,
				cred.Name,
				cred.Service,
				cred.Type,
				lastRotation.Format("2006-01-02"),
				daysOverdue,
			)
		}
	},

// getConfigDir returns the configuration directory path
func getConfigDir() string {
	// Check for environment variable first
	configDir := os.Getenv("LLMRT_CONFIG_DIR")
	if configDir != "" {
		return configDir
	}

	// Default to user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory can't be determined
		return "./.LLMrecon"
	}

	return filepath.Join(homeDir, ".LLMrecon")

// initCredentialManager initializes the credential manager
func initCredentialManager() error {
	// Get configuration directory
	configDir := getConfigDir()

	// Get passphrase
	passphrase := os.Getenv("LLMRT_VAULT_PASSPHRASE")
	if passphrase == "" {
		// Use a default passphrase for development
		// In production, this should be securely provided
		passphrase = os.Getenv("LLMRT_VAULT_PASSPHRASE")
	}

	// Initialize audit logger
	auditLogger, err := audit.NewFileAuditLogger(filepath.Join(configDir, "audit.log"))
	if err != nil {
		return fmt.Errorf("failed to initialize audit logger: %w", err)
	}

	// Create a security credential audit logger
	credentialLogPath := filepath.Join(configDir, "credential_audit.log")
	credentialLogger, err := securityaudit.NewCredentialAuditLogger(credentialLogPath, securityaudit.CredentialAuditLoggerOptions{})
	if err != nil {
		return fmt.Errorf("failed to initialize credential audit logger: %w", err)
	}

	// Create the security audit logger adapter
	// We need to pass both the credential logger and our standard audit logger
	securityAdapter := securityaudit.NewAuditLoggerAdapter(credentialLogger, auditLogger.Writer, auditLogger.User)

	// Initialize credential manager
	return vault.InitDefaultIntegration(configDir, passphrase, securityAdapter)

// No longer needed - using the built-in security audit logger adapter

func init() {
	rootCmd.AddCommand(credentialCmd)
	credentialCmd.AddCommand(credentialListCmd)
	credentialCmd.AddCommand(credentialShowCmd)
	credentialCmd.AddCommand(credentialAddCmd)
	credentialCmd.AddCommand(credentialUpdateCmd)
	credentialCmd.AddCommand(credentialDeleteCmd)
	credentialCmd.AddCommand(credentialRotateCmd)
	credentialCmd.AddCommand(credentialCheckRotationCmd)

	// Add flags for credential list command
	credentialListCmd.Flags().StringVarP(&credentialService, "service", "s", "", "Filter by service")
	credentialListCmd.Flags().StringVarP(&credentialType, "type", "t", "", "Filter by type")
	credentialListCmd.Flags().StringVarP(&credentialTags, "tags", "", "", "Filter by tags (comma-separated)")

	// Add flags for credential show command
	credentialShowCmd.Flags().Bool("show-value", false, "Show the credential value")

	// Add flags for credential add command
	credentialAddCmd.Flags().StringVarP(&credentialService, "service", "s", "", "Service name (required)")
	credentialAddCmd.Flags().StringVarP(&credentialType, "type", "t", "", "Credential type (required)")
	credentialAddCmd.Flags().StringVarP(&credentialValue, "value", "v", "", "Credential value (required)")
	credentialAddCmd.Flags().StringVarP(&credentialDesc, "description", "d", "", "Credential description")
	credentialAddCmd.Flags().StringVarP(&credentialTags, "tags", "", "", "Tags (comma-separated)")
	credentialAddCmd.Flags().IntVarP(&credentialRotation, "rotation-days", "r", 0, "Rotation interval in days")
	credentialAddCmd.Flags().IntVarP(&credentialWarning, "warning-days", "w", 14, "Warning days before rotation")

	// Add flags for credential update command
	credentialUpdateCmd.Flags().StringVarP(&credentialValue, "value", "v", "", "New credential value")
	credentialUpdateCmd.Flags().StringVarP(&credentialDesc, "description", "d", "", "New credential description")
	credentialUpdateCmd.Flags().StringVarP(&credentialTags, "tags", "", "", "New tags (comma-separated)")
	credentialUpdateCmd.Flags().IntVarP(&credentialRotation, "rotation-days", "r", 0, "New rotation interval in days")
	credentialUpdateCmd.Flags().IntVarP(&credentialWarning, "warning-days", "w", 14, "New warning days before rotation")

	// Add flags for credential rotate command
