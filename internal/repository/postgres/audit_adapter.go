package postgres

import (
	"context"

	"drone-platform/internal/repository"
)

// AuditAdapter bridges pg Store to repository.AuditWriter.
type AuditAdapter struct{ store *Store }

func NewAuditAdapter(s *Store) *AuditAdapter { return &AuditAdapter{store: s} }

func (a *AuditAdapter) WriteAudit(ctx context.Context, entry repository.AuditEntry) error {
	return a.store.WriteAudit(ctx, AuditEntry{
		ActorID:      entry.ActorID,
		Action:       entry.Action,
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID,
		Result:       entry.Result,
		RequestID:    entry.RequestID,
		Metadata:     entry.Metadata,
	})
}

// ListAudit 实现 repository.AuditReader（管理端审计追溯查询）。
func (a *AuditAdapter) ListAudit(ctx context.Context, f repository.AuditFilter, offset, limit int) ([]repository.AuditEntry, int, error) {
	items, total, err := a.store.ListAudit(ctx, f, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	out := make([]repository.AuditEntry, 0, len(items))
	for _, e := range items {
		out = append(out, repository.AuditEntry{
			ID:           e.ID,
			ActorID:      e.ActorID,
			Action:       e.Action,
			ResourceType: e.ResourceType,
			ResourceID:   e.ResourceID,
			Result:       e.Result,
			RequestID:    e.RequestID,
			Metadata:     e.Metadata,
			CreatedAt:    e.CreatedAt,
		})
	}
	return out, total, nil
}
