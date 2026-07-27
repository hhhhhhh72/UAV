package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// AuditEntry represents a single audit record.
type AuditEntry struct {
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	Result       string
	RequestID    string
	Metadata     map[string]any
}

// WriteAudit inserts an audit log entry.
func (s *Store) WriteAudit(ctx context.Context, e AuditEntry) error {
	meta, err := json.Marshal(e.Metadata)
	if err != nil { return fmt.Errorf("marshal audit metadata: %w", err) }
	id := fmt.Sprintf("audit-%d", time.Now().UnixNano())
	_, err = s.pool.Exec(ctx, `
		INSERT INTO audit_logs (id, actor_id, action, resource_type, resource_id, result, request_id, metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, e.ActorID, e.Action, e.ResourceType, e.ResourceID, e.Result, e.RequestID, meta, time.Now())
	if err != nil { return fmt.Errorf("write audit log: %w", err) }
	return nil
}
