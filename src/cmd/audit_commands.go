// Package cmd provides command-line interface functionality
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/perplext/LLMrecon/src/security/access"
)

var (
	auditUserID      string
	auditUsername    string
	auditAction      string
	auditResource    string
	auditResourceID  string
	auditIPAddress   string
	auditSeverity    string
	auditStatus      string
	auditSessionID   string
	auditStartTime   string
	auditEndTime     string
	auditLimit       int
	auditOffset      int
	auditOutputFormat string
)

// initAuditCommands initializes audit log commands
func initAuditCommands() {
	// Audit command
	auditCmd := &cobra.Command{
		Use:   "audit",
		Short: "Security audit log commands",
		Long:  `View and query security audit logs.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	accessControlCmd.AddCommand(auditCmd)

	// List audit logs command
	listAuditCmd := &cobra.Command{
		Use:   "list",
		Short: "List audit logs",
		Long:  `List security audit logs with optional filtering.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionSecurityAudit); err != nil {
				return err
			}

			// Create filter
			filter := &access.AuditLogFilter{
				UserID:     auditUserID,
				Username:   auditUsername,
				Action:     access.AuditAction(auditAction),
				Resource:   auditResource,
				ResourceID: auditResourceID,
				IPAddress:  auditIPAddress,
				Severity:   access.AuditSeverity(auditSeverity),
				Status:     auditStatus,
				SessionID:  auditSessionID,
				Limit:      auditLimit,
				Offset:     auditOffset,
			}

			// Parse time filters if provided
			if auditStartTime != "" {
				startTime, err := time.Parse(time.RFC3339, auditStartTime)
				if err != nil {
					return fmt.Errorf("invalid start time format: %v", err)
				}
				filter.StartTime = startTime
			}

			if auditEndTime != "" {
				endTime, err := time.Parse(time.RFC3339, auditEndTime)
				if err != nil {
					return fmt.Errorf("invalid end time format: %v", err)
				}
				filter.EndTime = endTime
			}

			// Query audit logs
			ctx := context.Background()
			logs, err := accessControlManager.QueryAuditLogs(ctx, filter)
			if err != nil {
				return fmt.Errorf("error querying audit logs: %v", err)
			}

			// Display logs
			if len(logs) == 0 {
				fmt.Println("No audit logs found matching the criteria")
				return nil
			}

			// Output format
			switch auditOutputFormat {
			case "json":
				outputAuditLogsJSON(logs)
			case "csv":
				outputAuditLogsCSV(logs)
			default:
				outputAuditLogsText(logs)
			}

			return nil
		},
	}

	// Add flags for filtering
	listAuditCmd.Flags().StringVar(&auditUserID, "user-id", "", "Filter by user ID")
	listAuditCmd.Flags().StringVar(&auditUsername, "username", "", "Filter by username")
	listAuditCmd.Flags().StringVar(&auditAction, "action", "", "Filter by action (login, logout, create, read, update, delete, execute, authorize, unauthorized, system, security)")
	listAuditCmd.Flags().StringVar(&auditResource, "resource", "", "Filter by resource type")
	listAuditCmd.Flags().StringVar(&auditResourceID, "resource-id", "", "Filter by resource ID")
	listAuditCmd.Flags().StringVar(&auditIPAddress, "ip", "", "Filter by IP address")
	listAuditCmd.Flags().StringVar(&auditSeverity, "severity", "", "Filter by severity (info, low, medium, high, critical, error)")
	listAuditCmd.Flags().StringVar(&auditStatus, "status", "", "Filter by status (success, failed)")
	listAuditCmd.Flags().StringVar(&auditSessionID, "session-id", "", "Filter by session ID")
	listAuditCmd.Flags().StringVar(&auditStartTime, "start", "", "Filter by start time (RFC3339 format)")
	listAuditCmd.Flags().StringVar(&auditEndTime, "end", "", "Filter by end time (RFC3339 format)")
	listAuditCmd.Flags().IntVar(&auditLimit, "limit", 100, "Limit number of results")
	listAuditCmd.Flags().IntVar(&auditOffset, "offset", 0, "Offset for pagination")
	listAuditCmd.Flags().StringVar(&auditOutputFormat, "format", "text", "Output format (text, json, csv)")

	auditCmd.AddCommand(listAuditCmd)

	// Get audit log command
	getAuditCmd := &cobra.Command{
		Use:   "get [log-id]",
		Short: "Get audit log details",
		Long:  `Get detailed information about a specific audit log entry.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionSecurityAudit); err != nil {
				return err
			}

			logID := args[0]

			// Create filter to get the specific log
			filter := &access.AuditLogFilter{
				Limit: 1,
			}

			// Query audit logs
			ctx := context.Background()
			logs, err := accessControlManager.QueryAuditLogs(ctx, filter)
			if err != nil {
				return fmt.Errorf("error querying audit logs: %v", err)
			}

			// Find the log with the specified ID
			var log *access.AuditLog
			for _, l := range logs {
				if l.ID == logID {
					log = l
					break
				}
			}

			if log == nil {
				return fmt.Errorf("audit log not found: %s", logID)
			}

			// Display log details
			fmt.Printf("Audit Log Details:\n")
			fmt.Printf("ID: %s\n", log.ID)
			fmt.Printf("Timestamp: %s\n", log.Timestamp.Format(time.RFC3339))
			if log.UserID != "" {
				fmt.Printf("User ID: %s\n", log.UserID)
			}
			if log.Username != "" {
				fmt.Printf("Username: %s\n", log.Username)
			}
			fmt.Printf("Action: %s\n", log.Action)
			fmt.Printf("Resource: %s\n", log.Resource)
			if log.ResourceID != "" {
				fmt.Printf("Resource ID: %s\n", log.ResourceID)
			}
			fmt.Printf("Description: %s\n", log.Description)
			if log.IPAddress != "" {
				fmt.Printf("IP Address: %s\n", log.IPAddress)
			}
			if log.UserAgent != "" {
				fmt.Printf("User Agent: %s\n", log.UserAgent)
			}
			fmt.Printf("Severity: %s\n", log.Severity)
			fmt.Printf("Status: %s\n", log.Status)
			if log.SessionID != "" {
				fmt.Printf("Session ID: %s\n", log.SessionID)
			}
			if log.Metadata != nil && len(log.Metadata) > 0 {
				fmt.Printf("Metadata:\n")
				for k, v := range log.Metadata {
					fmt.Printf("  %s: %v\n", k, v)
				}
			}

			return nil
		},
	}
	auditCmd.AddCommand(getAuditCmd)

	// Export audit logs command
	exportAuditCmd := &cobra.Command{
		Use:   "export [filename]",
		Short: "Export audit logs",
		Long:  `Export security audit logs to a file in the specified format.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionSecurityAudit); err != nil {
				return err
			}

			filename := args[0]

			// Create filter
			filter := &access.AuditLogFilter{
				UserID:     auditUserID,
				Username:   auditUsername,
				Action:     access.AuditAction(auditAction),
				Resource:   auditResource,
				ResourceID: auditResourceID,
				IPAddress:  auditIPAddress,
				Severity:   access.AuditSeverity(auditSeverity),
				Status:     auditStatus,
				SessionID:  auditSessionID,
				Limit:      auditLimit,
				Offset:     auditOffset,
			}

			// Parse time filters if provided
			if auditStartTime != "" {
				startTime, err := time.Parse(time.RFC3339, auditStartTime)
				if err != nil {
					return fmt.Errorf("invalid start time format: %v", err)
				}
				filter.StartTime = startTime
			}

			if auditEndTime != "" {
				endTime, err := time.Parse(time.RFC3339, auditEndTime)
				if err != nil {
					return fmt.Errorf("invalid end time format: %v", err)
				}
				filter.EndTime = endTime
			}

			// Query audit logs
			ctx := context.Background()
			logs, err := accessControlManager.QueryAuditLogs(ctx, filter)
			if err != nil {
				return fmt.Errorf("error querying audit logs: %v", err)
			}

			// Export logs to file
			if err := exportAuditLogs(logs, filename, auditOutputFormat); err != nil {
				return fmt.Errorf("error exporting audit logs: %v", err)
			}

			fmt.Printf("Exported %d audit logs to %s\n", len(logs), filename)

			return nil
		},
	}

	// Add flags for filtering
	exportAuditCmd.Flags().StringVar(&auditUserID, "user-id", "", "Filter by user ID")
	exportAuditCmd.Flags().StringVar(&auditUsername, "username", "", "Filter by username")
	exportAuditCmd.Flags().StringVar(&auditAction, "action", "", "Filter by action")
	exportAuditCmd.Flags().StringVar(&auditResource, "resource", "", "Filter by resource type")
	exportAuditCmd.Flags().StringVar(&auditResourceID, "resource-id", "", "Filter by resource ID")
	exportAuditCmd.Flags().StringVar(&auditIPAddress, "ip", "", "Filter by IP address")
	exportAuditCmd.Flags().StringVar(&auditSeverity, "severity", "", "Filter by severity")
	exportAuditCmd.Flags().StringVar(&auditStatus, "status", "", "Filter by status")
	exportAuditCmd.Flags().StringVar(&auditSessionID, "session-id", "", "Filter by session ID")
	exportAuditCmd.Flags().StringVar(&auditStartTime, "start", "", "Filter by start time (RFC3339 format)")
	exportAuditCmd.Flags().StringVar(&auditEndTime, "end", "", "Filter by end time (RFC3339 format)")
	exportAuditCmd.Flags().IntVar(&auditLimit, "limit", 1000, "Limit number of results")
	exportAuditCmd.Flags().IntVar(&auditOffset, "offset", 0, "Offset for pagination")
	exportAuditCmd.Flags().StringVar(&auditOutputFormat, "format", "csv", "Output format (json, csv)")

	auditCmd.AddCommand(exportAuditCmd)
}

// outputAuditLogsText outputs audit logs in text format
func outputAuditLogsText(logs []*access.AuditLog) {
	fmt.Printf("Found %d audit logs:\n\n", len(logs))
	for i, log := range logs {
		fmt.Printf("Log #%d:\n", i+1)
		fmt.Printf("  ID: %s\n", log.ID)
		fmt.Printf("  Timestamp: %s\n", log.Timestamp.Format(time.RFC3339))
		if log.UserID != "" {
			fmt.Printf("  User: %s", log.UserID)
			if log.Username != "" {
				fmt.Printf(" (%s)", log.Username)
			}
			fmt.Println()
		}
		fmt.Printf("  Action: %s\n", log.Action)
		fmt.Printf("  Resource: %s", log.Resource)
		if log.ResourceID != "" {
			fmt.Printf(" (%s)", log.ResourceID)
		}
		fmt.Println()
		fmt.Printf("  Description: %s\n", log.Description)
		fmt.Printf("  Severity: %s\n", log.Severity)
		fmt.Printf("  Status: %s\n", log.Status)
		fmt.Println()
	}
}

// outputAuditLogsJSON outputs audit logs in JSON format
func outputAuditLogsJSON(logs []*access.AuditLog) {
	// In a real implementation, this would use json.Marshal
	fmt.Println("JSON output not implemented in this example")
}

// outputAuditLogsCSV outputs audit logs in CSV format
func outputAuditLogsCSV(logs []*access.AuditLog) {
	// Print CSV header
	fmt.Println("ID,Timestamp,UserID,Username,Action,Resource,ResourceID,Description,IPAddress,Severity,Status,SessionID")
	
	// Print each log as a CSV row
	for _, log := range logs {
		fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			log.ID,
			log.Timestamp.Format(time.RFC3339),
			log.UserID,
			log.Username,
			log.Action,
			log.Resource,
			log.ResourceID,
			escapeCSV(log.Description),
			log.IPAddress,
			log.Severity,
			log.Status,
			log.SessionID,
		)
	}
}

// exportAuditLogs exports audit logs to a file
func exportAuditLogs(logs []*access.AuditLog, filename string, format string) error {
	// In a real implementation, this would write to a file
	fmt.Printf("Would export %d logs to %s in %s format\n", len(logs), filename, format)
	return nil
}

// escapeCSV escapes a string for CSV output
func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}
