// Package cmd provides command-line interfaces for the LLMrecon tool
package cmd

import (
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/api"
	"github.com/perplext/LLMrecon/src/security"
	securityapi "github.com/perplext/LLMrecon/src/security/api"
	"github.com/perplext/LLMrecon/src/security/communication"
	"github.com/spf13/cobra"
)

// apiServerCmd represents the api server command
var apiServerCmd = &cobra.Command{
	Use:     "api",
	Aliases: []string{"serve", "server"},
	Short:   "Start the API server",
	Long:    `Start the HTTP API server for managing red-team scans with enhanced security features.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		addr, _ := cmd.Flags().GetString("addr")
		useTLS, _ := cmd.Flags().GetBool("tls")
		certFile, _ := cmd.Flags().GetString("cert")
		keyFile, _ := cmd.Flags().GetString("key")
		logPath, _ := cmd.Flags().GetString("log")
		rateLimit, _ := cmd.Flags().GetInt("rate-limit")
		ipAllowlist, _ := cmd.Flags().GetString("ip-allowlist")
		developmentMode, _ := cmd.Flags().GetBool("dev-mode")

		// Create server config
		config := &api.ServerConfig{
			Address:  addr,
			UseTLS:   useTLS,
			CertFile: certFile,
			KeyFile:  keyFile,
			SecurityConfig: &security.SecurityConfig{
				TLSConfig: &communication.TLSConfig{
					MinVersion: 0, // Will use default (TLS 1.2)
					CertFile:   certFile,
					KeyFile:    keyFile,
				},
				DevelopmentMode: developmentMode,
				LogFilePath:     logPath,
			},
		}

		// Configure rate limiting if specified
		if rateLimit > 0 {
			config.SecurityConfig.RateLimiterConfig = &securityapi.RateLimiterConfig{
				RequestsPerMinute: rateLimit,
				BurstSize:         rateLimit / 10, // 10% of rate limit
			}
		}

		// Configure IP allowlist if specified
		if ipAllowlist != "" {
			// Check if it's a file
			if _, err := os.Stat(ipAllowlist); err == nil {
				config.SecurityConfig.IPAllowlistConfig = &securityapi.IPAllowlistConfig{
					Enabled:    true,
					ConfigFile: ipAllowlist,
				}
			} else {
				// Assume it's a comma-separated list of IPs
				config.SecurityConfig.IPAllowlistConfig = &securityapi.IPAllowlistConfig{
					Enabled:    true,
					AllowedIPs: splitAndTrim(ipAllowlist, ","),
				}
			}
		}

		// Create and start server
		server, err := api.NewServer(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating server: %v\n", err)
			os.Exit(1)
		}

		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
			os.Exit(1)
		}
	},
}

// splitAndTrim splits a string by a separator and trims spaces
func splitAndTrim(s string, sep string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

func init() {
	rootCmd.AddCommand(apiServerCmd)

	// Add basic flags
	apiServerCmd.Flags().StringP("addr", "a", ":8080", "Address to listen on (e.g., :8080)")

	// Add TLS flags
	apiServerCmd.Flags().Bool("tls", false, "Enable TLS/HTTPS")
	apiServerCmd.Flags().String("cert", "certs/server.crt", "TLS certificate file path")
	apiServerCmd.Flags().String("key", "certs/server.key", "TLS key file path")

	// Add security flags
	apiServerCmd.Flags().String("log", "logs/security.log", "Security log file path")
	apiServerCmd.Flags().Int("rate-limit", 60, "Rate limit in requests per minute (0 to disable)")
	apiServerCmd.Flags().String("ip-allowlist", "", "IP allowlist (comma-separated list or file path)")
	apiServerCmd.Flags().Bool("dev-mode", false, "Enable development mode (more verbose errors)")
}
