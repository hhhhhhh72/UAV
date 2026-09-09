package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"drone-platform/internal/repository"
)

// AuditEntry represents a single audit record.
type AuditEntry struct {
	ID           string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	Result       string
	RequestID    string
	Metadata     map[string]any
	CreatedAt    time.Time
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
// ListAudit 分页查询审计日志（管理端「操作审计」页）：按时间倒序，支持操作人/动作/
// 资源类型/时间范围过滤。返回 (记录, 总数)。
func (s *Store) ListAudit(ctx context.Context, f repository.AuditFilter, offset, limit int) ([]AuditEntry, int, error) {
	where := []string{"1=1"}
	args := []any{}
	idx := 1
	if f.ActorID != "" {
		where = append(where, fmt.Sprintf("actor_id = $%d", idx))
		args = append(args, f.ActorID)
		idx++
	}
	if f.Action != "" {
		where = append(where, fmt.Sprintf("action = $%d", idx))
		args = append(args, f.Action)
		idx++
	}
	if f.ResourceType != "" {
		where = append(where, fmt.Sprintf("resource_type = $%d", idx))
		args = append(args, f.ResourceType)
		idx++
	}
	if !f.Start.IsZero() {
		where = append(where, fmt.Sprintf("created_at >= $%d", idx))
		args = append(args, f.Start)
		idx++
	}
	if !f.End.IsZero() {
		where = append(where, fmt.Sprintf("created_at <= $%d", idx))
		args = append(args, f.End)
		idx++
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM audit_logs WHERE "+cond, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	q := "SELECT id, actor_id, action, resource_type, resource_id, result, request_id, COALESCE(metadata,'{}'::jsonb), created_at FROM audit_logs WHERE " +
		cond + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var meta []byte
		if err := rows.Scan(&e.ID, &e.ActorID, &e.Action, &e.ResourceType, &e.ResourceID, &e.Result, &e.RequestID, &meta, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		if len(meta) > 0 {
			_ = json.Unmarshal(meta, &e.Metadata)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

