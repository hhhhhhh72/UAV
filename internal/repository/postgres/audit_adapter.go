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
