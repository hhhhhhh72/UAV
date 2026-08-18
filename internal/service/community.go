package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type CommunityService struct {
	post    repository.PostRepository
	comment repository.CommentRepository
	report  repository.ReportRepository
}

func NewCommunityService(p repository.PostRepository, c repository.CommentRepository, r repository.ReportRepository) *CommunityService {
	return &CommunityService{post: p, comment: c, report: r}
}

// ---- Posts ----

func (s *CommunityService) CreatePost(ctx context.Context, a domain.Actor, title, content string, images []string) (domain.Post, error) {
	now := time.Now()
	p := domain.Post{ID: nextID("post"), AuthorID: a.ID, Title: title, Content: content, Images: images, Status: "published", Version: 1, CreatedAt: now, UpdatedAt: now}
	slog.Info("post created", "post_id", p.ID, "author_id", p.AuthorID)
	return s.post.Create(ctx, p)
}

func (s *CommunityService) PublishPost(ctx context.Context, a domain.Actor, id string) (domain.Post, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Post{}, errors.New("admin permission required")
	}
	p, err := s.post.FindByID(ctx, id)
	if err != nil {
		return domain.Post{}, err
	}
	p.Status = "published"
	p.UpdatedAt = time.Now()
	slog.Info("post updated", "post_id", id)
	return s.post.Update(ctx, id, p)
}

func (s *CommunityService) RemovePost(ctx context.Context, a domain.Actor, id string) (domain.Post, error) {
	p, err := s.post.FindByID(ctx, id)
	if err != nil {
		return domain.Post{}, err
	}
	if p.AuthorID != a.ID && a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Post{}, errors.New("permission denied")
	}
	p.Status = "removed"
	p.UpdatedAt = time.Now()
	slog.Info("post updated", "post_id", id)
	return s.post.Update(ctx, id, p)
}

func (s *CommunityService) ListPublishedPosts(ctx context.Context, offset, limit int) ([]domain.Post, int, error) {
	return s.post.ListPublished(ctx, offset, limit)
}

// ---- Comments ----

func (s *CommunityService) CreateComment(ctx context.Context, a domain.Actor, postID, content string) (domain.Comment, error) {
	c := domain.Comment{ID: nextID("comment"), PostID: postID, AuthorID: a.ID, Content: content, Status: "active", CreatedAt: time.Now()}
	return s.comment.Create(ctx, c)
}

func (s *CommunityService) ListComments(ctx context.Context, postID string) ([]domain.Comment, error) {
	return s.comment.ListByPost(ctx, postID)
}

// ---- Reports ----

func (s *CommunityService) CreateReport(ctx context.Context, a domain.Actor, resourceType, resourceID, reason string) (domain.Report, error) {
	rp := domain.Report{ID: nextID("report"), ReporterID: a.ID, ResourceType: resourceType, ResourceID: resourceID, Reason: reason, Status: "pending", CreatedAt: time.Now()}
	return s.report.Create(ctx, rp)
}

func (s *CommunityService) ListPendingReports(ctx context.Context, a domain.Actor, offset, limit int) ([]domain.Report, int, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, 0, errors.New("admin permission required")
	}
	return s.report.ListPending(ctx, offset, limit)
}
