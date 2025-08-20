// Package db provides database implementations of the access control interfaces
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/perplext/LLMrecon/src/security/access/db/adapter"
)

// SQLAuditStore is a SQL implementation of ModelsAuditStore
type SQLAuditStore struct {
	db *sql.DB
}

// NewSQLAuditStore creates a new SQL-based audit store
func NewSQLAuditStore(db *sql.DB) (adapter.ModelsAuditStore, error) {
	store := &SQLAuditStore{
		db: db,
	}

	// Initialize database schema if needed
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil

// initSchema initializes the database schema
func (s *SQLAuditStore) initSchema() error {
	// Create audit_logs table
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			timestamp TIMESTAMP NOT NULL,
			user_id TEXT,
			username TEXT,
			action TEXT NOT NULL,
			resource TEXT,
			resource_id TEXT,
			severity TEXT NOT NULL,
			status TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			details TEXT,
			metadata TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create audit_logs table: %w", err)
	}

	// Create indexes
	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp)`)
	if err != nil {
		return fmt.Errorf("failed to create timestamp index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)`)
	if err != nil {
		return fmt.Errorf("failed to create user_id index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action)`)
	if err != nil {
		return fmt.Errorf("failed to create action index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_logs_severity ON audit_logs(severity)`)
	if err != nil {
		return fmt.Errorf("failed to create severity index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource)`)
	if err != nil {
		return fmt.Errorf("failed to create resource index: %w", err)
	}

	return nil

// LogEvent logs an audit event
func (s *SQLAuditStore) LogEvent(ctx context.Context, event *adapter.AuditEvent) error {
	// Serialize details
	detailsJSON, err := json.Marshal(event.Details)
	if err != nil {
		return fmt.Errorf("failed to serialize details: %w", err)
	}

	// Serialize metadata
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	query := `
		INSERT INTO audit_logs (
			id, timestamp, user_id, username, action, resource, resource_id,
			severity, status, ip_address, user_agent, details, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(
		ctx,
		query,
		event.ID,
		formatTime(event.Timestamp),
		event.UserID,
		event.Username,
		event.Action,
		event.Resource,
		event.ResourceID,
		event.Severity,
		event.Status,
		event.IPAddress,
		event.UserAgent,
		detailsJSON,
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	return nil

// GetEventByID retrieves an audit event by ID
func (s *SQLAuditStore) GetEventByID(ctx context.Context, id string) (*adapter.AuditEvent, error) {
	query := `
		SELECT id, timestamp, user_id, username, action, resource, resource_id,
		       severity, status, ip_address, user_agent, details, metadata
		FROM audit_logs
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanAuditEvent(row)

// QueryEvents queries audit events based on filters
func (s *SQLAuditStore) QueryEvents(ctx context.Context, filter *adapter.AuditEventFilter, offset, limit int) ([]*adapter.AuditEvent, int, error) {
	// Build query
	baseQuery := `
		SELECT id, timestamp, user_id, username, action, resource, resource_id,
		       severity, status, ip_address, user_agent, details, metadata
		FROM audit_logs
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}

	// Add filter conditions
	if !filter.StartTime.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, formatTime(filter.StartTime))
	}

	if !filter.EndTime.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, formatTime(filter.EndTime))
	}

	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Username != "" {
		conditions = append(conditions, "username = ?")
		args = append(args, filter.Username)
	}

	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}

	if filter.Resource != "" {
		conditions = append(conditions, "resource = ?")
		args = append(args, filter.Resource)
	}

	if filter.ResourceID != "" {
		conditions = append(conditions, "resource_id = ?")
		args = append(args, filter.ResourceID)
	}

	if filter.Severity != "" {
		conditions = append(conditions, "severity = ?")
		args = append(args, filter.Severity)
	}

	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	if filter.IPAddress != "" {
		conditions = append(conditions, "ip_address = ?")
		args = append(args, filter.IPAddress)
	}

	// Build the final query
	query := baseQuery
	for _, condition := range conditions {
		query += " AND " + condition
	}

	// Add order by
	query += " ORDER BY timestamp DESC"

	// Add limit and offset
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	// Execute query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer func() { if err := rows.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	// Scan results
	var events []*adapter.AuditEvent
	for rows.Next() {
		event, err := s.scanAuditEventFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audit events: %w", err)
	}

	return events, len(events), nil

// ExportEvents exports audit events to a file
func (s *SQLAuditStore) ExportEvents(ctx context.Context, filter *adapter.AuditEventFilter, format string) (string, error) {
	// This is a placeholder implementation
	return "", fmt.Errorf("export not implemented")

// CountEvents counts audit events based on filters
func (s *SQLAuditStore) CountEvents(ctx context.Context, filter *adapter.AuditEventFilter) (int64, error) {
	// Build query
	baseQuery := `
		SELECT COUNT(*)
		FROM audit_logs
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}

	// Add filter conditions
	if !filter.StartTime.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, formatTime(filter.StartTime))
	}

	if !filter.EndTime.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, formatTime(filter.EndTime))
	}

	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Username != "" {
		conditions = append(conditions, "username = ?")
		args = append(args, filter.Username)
	}

	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}

	if filter.Resource != "" {
		conditions = append(conditions, "resource = ?")
		args = append(args, filter.Resource)
	}

	if filter.ResourceID != "" {
		conditions = append(conditions, "resource_id = ?")
		args = append(args, filter.ResourceID)
	}

	if filter.Severity != "" {
		conditions = append(conditions, "severity = ?")
		args = append(args, filter.Severity)
	}

	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	if filter.IPAddress != "" {
		conditions = append(conditions, "ip_address = ?")
		args = append(args, filter.IPAddress)
	}

	// Build the final query
	query := baseQuery
	for _, condition := range conditions {
		query += " AND " + condition
	}

	// Execute query
	var count int64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	return count, nil

// scanAuditEvent scans an audit event from a database row
func (s *SQLAuditStore) scanAuditEvent(row *sql.Row) (*adapter.AuditEvent, error) {
	var event adapter.AuditEvent
	var timestampStr string
	var userID, username, ipAddress, userAgent sql.NullString
	var detailsJSON, metadataJSON string

	err := row.Scan(
		&event.ID,
		&timestampStr,
		&userID,
		&username,
		&event.Action,
		&event.Resource,
		&event.ResourceID,
		&event.Severity,
		&event.Status,
		&ipAddress,
		&userAgent,
		&detailsJSON,
		&metadataJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("audit event not found")
		}
		return nil, fmt.Errorf("failed to scan audit event: %w", err)
	}

	// Parse timestamp
	event.Timestamp, _ = time.Parse(time.RFC3339, timestampStr)

	// Handle nullable fields
	if userID.Valid {
		event.UserID = userID.String
	}
	if username.Valid {
		event.Username = username.String
	}
	if ipAddress.Valid {
		event.IPAddress = ipAddress.String
	}
	if userAgent.Valid {
		event.UserAgent = userAgent.String
	}

	// Parse details
	if detailsJSON != "" {
		var details map[string]interface{}
		if err := json.Unmarshal([]byte(detailsJSON), &details); err != nil {
			return nil, fmt.Errorf("failed to parse details: %w", err)
		}
		event.Details = details
	}

	// Parse metadata
	if metadataJSON != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
		event.Metadata = metadata
	}

	return &event, nil

// scanAuditEventFromRows scans an audit event from database rows
func (s *SQLAuditStore) scanAuditEventFromRows(rows *sql.Rows) (*adapter.AuditEvent, error) {
	var event adapter.AuditEvent
	var timestampStr string
	var userID, username, ipAddress, userAgent sql.NullString
	var detailsJSON, metadataJSON string

	err := rows.Scan(
		&event.ID,
		&timestampStr,
		&userID,
		&username,
		&event.Action,
		&event.Resource,
		&event.ResourceID,
		&event.Severity,
		&event.Status,
		&ipAddress,
		&userAgent,
		&detailsJSON,
		&metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan audit event: %w", err)
	}

	// Parse timestamp
	event.Timestamp, _ = time.Parse(time.RFC3339, timestampStr)

	// Handle nullable fields
	if userID.Valid {
		event.UserID = userID.String
	}
	if username.Valid {
		event.Username = username.String
	}
	if ipAddress.Valid {
		event.IPAddress = ipAddress.String
	}
	if userAgent.Valid {
		event.UserAgent = userAgent.String
	}

	// Parse details
	if detailsJSON != "" {
		var details map[string]interface{}
		if err := json.Unmarshal([]byte(detailsJSON), &details); err != nil {
			return nil, fmt.Errorf("failed to parse details: %w", err)
		}
		event.Details = details
	}

	// Parse metadata
	if metadataJSON != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
		event.Metadata = metadata
	}

	return &event, nil

// Close closes the SQL connection
func (s *SQLAuditStore) Close() error {
	return s.db.Close()
}
}
}
}
}
}
}
}
}
