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

// SQLIncidentStore is a SQL implementation of IncidentStore
type SQLIncidentStore struct {
	db *sql.DB
}

// NewSQLIncidentStore creates a new SQL-based incident store
func NewSQLIncidentStore(db *sql.DB) (adapter.IncidentStore, error) {
	store := &SQLIncidentStore{
		db: db,
	}

	// Initialize database schema if needed
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema initializes the database schema
func (s *SQLIncidentStore) initSchema() error {
	// Create incidents table
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS incidents (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			severity TEXT NOT NULL,
			status TEXT NOT NULL,
			reported_by TEXT,
			assigned_to TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			resolved_at TIMESTAMP,
			metadata TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create incidents table: %w", err)
	}

	// Create indexes
	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity)`)
	if err != nil {
		return fmt.Errorf("failed to create severity index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status)`)
	if err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_incidents_reported_by ON incidents(reported_by)`)
	if err != nil {
		return fmt.Errorf("failed to create reported_by index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_incidents_assigned_to ON incidents(assigned_to)`)
	if err != nil {
		return fmt.Errorf("failed to create assigned_to index: %w", err)
	}

	return nil
}

// CreateIncident creates a new security incident
func (s *SQLIncidentStore) CreateIncident(ctx context.Context, incident *adapter.IncidentEvent) error {
	// Serialize metadata
	metadataJSON, err := json.Marshal(incident.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	query := `
		INSERT INTO incidents (
			id, title, description, severity, status, reported_by, assigned_to,
			created_at, updated_at, resolved_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	createdAt := time.Now().Format(time.RFC3339)
	if incident.CreatedAt != "" {
		createdAt = incident.CreatedAt
	}

	updatedAt := createdAt
	if incident.UpdatedAt != "" {
		updatedAt = incident.UpdatedAt
	}

	var resolvedAt interface{}
	if incident.ResolvedAt != "" {
		resolvedAt = incident.ResolvedAt
	} else {
		resolvedAt = nil
	}

	_, err = s.db.ExecContext(
		ctx,
		query,
		incident.ID,
		incident.Title,
		incident.Description,
		incident.Severity,
		incident.Status,
		incident.ReportedBy,
		incident.AssignedTo,
		createdAt,
		updatedAt,
		resolvedAt,
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	return nil
}

// GetIncidentByID retrieves a security incident by ID
func (s *SQLIncidentStore) GetIncidentByID(ctx context.Context, id string) (*adapter.IncidentEvent, error) {
	query := `
		SELECT id, title, description, severity, status, reported_by, assigned_to,
		       created_at, updated_at, resolved_at, metadata
		FROM incidents
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanIncident(row)
}

// UpdateIncident updates an existing security incident
func (s *SQLIncidentStore) UpdateIncident(ctx context.Context, incident *adapter.IncidentEvent) error {
	// Check if incident exists
	_, err := s.GetIncidentByID(ctx, incident.ID)
	if err != nil {
		return err
	}

	// Serialize metadata
	metadataJSON, err := json.Marshal(incident.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	query := `
		UPDATE incidents
		SET title = ?, description = ?, severity = ?, status = ?, reported_by = ?,
		    assigned_to = ?, updated_at = ?, resolved_at = ?, metadata = ?
		WHERE id = ?
	`

	updatedAt := time.Now().Format(time.RFC3339)
	if incident.UpdatedAt != "" {
		updatedAt = incident.UpdatedAt
	}

	var resolvedAt interface{}
	if incident.ResolvedAt != "" {
		resolvedAt = incident.ResolvedAt
	} else {
		resolvedAt = nil
	}

	_, err = s.db.ExecContext(
		ctx,
		query,
		incident.Title,
		incident.Description,
		incident.Severity,
		incident.Status,
		incident.ReportedBy,
		incident.AssignedTo,
		updatedAt,
		resolvedAt,
		metadataJSON,
		incident.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	return nil
}

// DeleteIncident deletes a security incident
func (s *SQLIncidentStore) DeleteIncident(ctx context.Context, id string) error {
	query := `DELETE FROM incidents WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found")
	}

	return nil
}

// ListIncidents lists security incidents with optional filtering
func (s *SQLIncidentStore) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*adapter.IncidentEvent, int, error) {
	// Build query
	baseQuery := `
		SELECT id, title, description, severity, status, reported_by, assigned_to,
		       created_at, updated_at, resolved_at, metadata
		FROM incidents
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}

	// Add filter conditions
	if severity, ok := filter["severity"].(string); ok && severity != "" {
		conditions = append(conditions, "severity = ?")
		args = append(args, severity)
	}

	if status, ok := filter["status"].(string); ok && status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}

	if reportedBy, ok := filter["reported_by"].(string); ok && reportedBy != "" {
		conditions = append(conditions, "reported_by = ?")
		args = append(args, reportedBy)
	}

	if assignedTo, ok := filter["assigned_to"].(string); ok && assignedTo != "" {
		conditions = append(conditions, "assigned_to = ?")
		args = append(args, assignedTo)
	}

	if startDate, ok := filter["start_date"].(string); ok && startDate != "" {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, startDate)
	}

	if endDate, ok := filter["end_date"].(string); ok && endDate != "" {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, endDate)
	}

	// Build the final query
	query := baseQuery
	for _, condition := range conditions {
		query += " AND " + condition
	}

	// Add order by
	query += " ORDER BY created_at DESC"

	// Add limit and offset
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)

		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}

	// Execute query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	// Scan results
	var incidents []*adapter.IncidentEvent
	for rows.Next() {
		incident, err := s.scanIncidentFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		incidents = append(incidents, incident)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating incidents: %w", err)
	}

	// Count total incidents
	countQuery := `
		SELECT COUNT(*)
		FROM incidents
		WHERE 1=1
	`
	for _, condition := range conditions {
		countQuery += " AND " + condition
	}

	var count int
	err = s.db.QueryRowContext(ctx, countQuery, args[:len(args)-len(incidents)]...).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count incidents: %w", err)
	}

	return incidents, count, nil
}

// scanIncident scans a security incident from a database row
func (s *SQLIncidentStore) scanIncident(row *sql.Row) (*adapter.IncidentEvent, error) {
	var incident adapter.IncidentEvent
	var reportedBy, assignedTo, resolvedAt sql.NullString
	var metadataJSON string

	err := row.Scan(
		&incident.ID,
		&incident.Title,
		&incident.Description,
		&incident.Severity,
		&incident.Status,
		&reportedBy,
		&assignedTo,
		&incident.CreatedAt,
		&incident.UpdatedAt,
		&resolvedAt,
		&metadataJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("incident not found")
		}
		return nil, fmt.Errorf("failed to scan incident: %w", err)
	}

	// Handle nullable fields
	if reportedBy.Valid {
		incident.ReportedBy = reportedBy.String
	}
	if assignedTo.Valid {
		incident.AssignedTo = assignedTo.String
	}
	if resolvedAt.Valid {
		incident.ResolvedAt = resolvedAt.String
	}

	// Parse metadata
	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &incident.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	return &incident, nil
}

// scanIncidentFromRows scans a security incident from database rows
func (s *SQLIncidentStore) scanIncidentFromRows(rows *sql.Rows) (*adapter.IncidentEvent, error) {
	var incident adapter.IncidentEvent
	var reportedBy, assignedTo, resolvedAt sql.NullString
	var metadataJSON string

	err := rows.Scan(
		&incident.ID,
		&incident.Title,
		&incident.Description,
		&incident.Severity,
		&incident.Status,
		&reportedBy,
		&assignedTo,
		&incident.CreatedAt,
		&incident.UpdatedAt,
		&resolvedAt,
		&metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan incident: %w", err)
	}

	// Handle nullable fields
	if reportedBy.Valid {
		incident.ReportedBy = reportedBy.String
	}
	if assignedTo.Valid {
		incident.AssignedTo = assignedTo.String
	}
	if resolvedAt.Valid {
		incident.ResolvedAt = resolvedAt.String
	}

	// Parse metadata
	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &incident.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	return &incident, nil
}

// Close closes the SQL connection
func (s *SQLIncidentStore) Close() error {
	return s.db.Close()
}
