package service

import (
	"context"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type NewsService struct {
	repo repository.ArticleRepository
}

func NewNewsService(repo repository.ArticleRepository) *NewsService {
	return &NewsService{repo: repo}
}

func (s *NewsService) Create(ctx context.Context, title, content, category, source string) (domain.Article, error) {
	now := time.Now()
	a := domain.Article{ID: fmt.Sprintf("article-%d", now.UnixNano()), Title: title, Content: content,
		Summary: truncate(content, 100), Category: category, Source: source, Status: "draft",
		Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.repo.Create(ctx, a)
}

// Update 编辑资讯内容（保留 ID/状态/创建时间，摘要重新截取）。
func (s *NewsService) Update(ctx context.Context, id, title, content, category, source string) (domain.Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	a.Title = title
	a.Content = content
	a.Summary = truncate(content, 100)
	a.Category = category
	a.Source = source
	a.UpdatedAt = time.Now()
	return s.repo.Update(ctx, a)
}

func (s *NewsService) Publish(ctx context.Context, id string) (domain.Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	a.Status = "published"
	a.UpdatedAt = time.Now()
	return s.repo.Update(ctx, a)
}

func (s *NewsService) ListByCategory(ctx context.Context, category string, page, pageSize int) ([]domain.Article, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByCategory(ctx, category, offset, pageSize)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
